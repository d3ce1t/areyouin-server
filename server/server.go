package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"peeple/areyouin/api"
	fb "peeple/areyouin/facebook"
	"peeple/areyouin/model"
	proto "peeple/areyouin/protocol"
	wh "peeple/areyouin/webhook"
	"strings"
	"time"
)

const (
	GCM_API_KEY = "AIzaSyAf-h1zJCRWNDt-dI3liL1yx4NEYjOq5GQ"
	GCM_MAX_TTL = 2419200
)

func NewServer(session api.DbSession, model *model.AyiModel) *Server {
	server := &Server{
		DbSession: session,
		Model:     model,
	}
	server.init()
	return server
}

type Callback func(*proto.AyiPacket, proto.Message, *AyiSession)

type Server struct {
	TLSConfig       *tls.Config
	sessions        *SessionsMap
	task_executor   *TaskExecutor
	callbacks       map[proto.PacketType]Callback
	Model           *model.AyiModel
	DbSession       api.DbSession
	webhook         *wh.WebHookServer
	Testing         bool
	MaintenanceMode bool
}

// Setup server components
func (s *Server) init() {

	// Init TLS config
	cert, err := tls.LoadX509KeyPair("cert/fullchain.pem", "cert/privkey.pem")
	if err != nil {
		panic(err)
	}

	s.TLSConfig = &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{cert},
		ServerName:   "service.peeple.es",
	}

	// Init sessions holder
	s.sessions = NewSessionsMap()
	s.callbacks = make(map[proto.PacketType]Callback)

	// Task Executor
	s.task_executor = NewTaskExecutor(s)
	s.task_executor.Start()
}

func (s *Server) run() {

	// Start webhook
	if !s.MaintenanceMode {
		s.webhook = wh.New(fb.FB_APP_SECRET)
		s.webhook.RegisterCallback(s.onFacebookUpdate)
		s.webhook.Run()
	} else {
		log.Println("Server running in MAINTENANCE MODE")
	}

	// Start up server listener
	listener, err := net.Listen("tcp", ":1822")

	if err != nil {
		panic("Couldn't start listening: " + err.Error())
	}

	defer listener.Close()

	// Main Loop
	for {
		client, err := listener.Accept()

		if err != nil {
			log.Println("Couldn't accept:", err.Error())
			continue
		}

		session := NewSession(client, s)
		go s.handleSession(session)
	}
}

func (s *Server) registerCallback(command proto.PacketType, f Callback) {
	if s.callbacks == nil {
		s.callbacks = make(map[proto.PacketType]Callback)
	}
	s.callbacks[command] = f
}

func (s *Server) registerSession(session *AyiSession) {

	if oldSession, ok := s.sessions.Get(session.UserId); ok {
		oldSession.WriteSync(oldSession.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD))
		log.Printf("< (%v) SEND INVALID USER OR PASSWORD\n", oldSession)
		oldSession.Exit()
		log.Printf("* (%v) Closing old session for endpoint %v\n", oldSession, oldSession.Conn.RemoteAddr())
	}

	s.sessions.Put(session.UserId, session)
	log.Printf("* (%v) Register session for endpoint %v\n", session, session.Conn.RemoteAddr())
}

func (s *Server) unregisterSession(session *AyiSession) {

	user_id := session.UserId
	if !session.IsClosed() {
		session.Exit()
	}

	oldSession, ok := s.sessions.Get(user_id)
	if ok && oldSession == session {
		s.sessions.Remove(user_id)
		log.Printf("* (%v) Unregister session for endpoint %v\n", session, session.Conn.RemoteAddr())
	}

}

// Private methods
func (s *Server) handleSession(session *AyiSession) {

	defer func() { // session.RunLoop() may throw panic
		if r := recover(); r != nil {
			log.Printf("Session %v Panic: %v\n", session, r)
		}
		log.Printf("* (%v) Session finished\n", session)
	}()

	log.Printf("* (%v) New connection\n", session)

	// Packet received
	session.OnRead = func(s *AyiSession, packet *proto.AyiPacket) {

		if !s.Server.MaintenanceMode {

			// Normal operation

			if err := s.Server.serveMessage(packet, s); err != nil { // may block until writes are performed
				errorCode := getNetErrorCode(err, proto.E_OPERATION_FAILED)
				log.Printf("< (%v) ERROR %v: %v\n", session, errorCode, err)
				s.WriteResponse(packet.Header.GetToken(), s.NewMessage().Error(packet.Type(), errorCode))
			}

		} else {

			// Maintenance Mode
			log.Printf("> (%v) Packet %v received but ignored\n", session, packet.Type())
			s.WriteSync(s.NewMessage().Error(packet.Type(), proto.E_SERVER_MAINTENANCE))
			log.Printf("< (%v) SERVER IS IN MAINTENANCE MODE\n", session)
			s.Exit()

		}

	}

	// Error
	session.OnError = func() func(s *AyiSession, err error) {

		num_errors := 0
		last_error_time := time.Now()
		return func(s *AyiSession, err error) {

			log.Printf("* (%v) Session Error: %v (num_errors = %v)\n", s, err, num_errors)

			if err == proto.ErrTimeout {
				s.Exit()

				// HACK: Compare error string because there is no ErrTlsXXXXX or alike
			} else if strings.Contains(err.Error(), "tls: first record does not look like a TLS handshake") {
				s.Exit()
			} else if strings.Contains(err.Error(), "unknown certificate") {
				s.Exit()
			}

			// Protect agains error flood
			current_time := time.Now()
			if last_error_time.Add(1 * time.Second).Before(current_time) {
				last_error_time = current_time
				num_errors = 1
			} else {
				num_errors++
				if num_errors == 10 {
					s.Exit()
				}
			}
		}
	}()

	// Closed connection
	session.OnClosed = func(session *AyiSession, peer bool) {

		if session.IsAuth {
			//NOTE: If a user is deleted from user_account while it is still connected,
			// a row in invalid state will be created when updating last connection
			session.Server.refreshSessionActivity(session)
			session.Server.unregisterSession(session)
		}

		if peer {
			log.Printf("* (%v) Session closed by remote peer\n", session)
		} else {
			log.Printf("* (%v) Session closed by server\n", session)
		}
	}

	// Called periodically (each 10 minutes)
	session.OnIdle = func(session *AyiSession) {

		currentTime := time.Now()

		if !session.IsAuth {

			if currentTime.After(session.lastRecvMsg.Add(MAX_LOGIN_TIME)) {
				log.Println("Connection IDLE", s)
				session.Exit()
			}

		} else {

			if currentTime.After(session.lastRecvMsg.Add(MAX_IDLE_TIME)) {

				log.Println("Connection IDLE", s)
				session.Exit()

			} else {

				if currentTime.After(session.pingTime) {
					session.ping()
					log.Printf("< (%v) PING", session)
				}

				session.Server.refreshSessionActivity(session)
			}
		}
	}

	session.RunLoop() // Block here
}

func (s *Server) refreshSessionActivity(session *AyiSession) {
	err := s.Model.Accounts.RefreshSessionActivity(session.UserId)
	if err != nil {
		log.Printf("* (%v) REFRESH SESSION ACTIVITY ERROR: %v", session, err)
	}
}

func (s *Server) serveMessage(packet *proto.AyiPacket, session *AyiSession) (err error) {

	// Decodes payload. If message does not have payload, ignore that
	message, decode_err := packet.DecodeMessage()
	if err != nil && err != proto.ErrNoPayload {
		return decode_err
	}

	// Defer recovery
	defer func() {
		if r := recover(); r != nil {
			if err_tmp, ok := r.(error); ok {
				err = err_tmp
			} else {
				err = errors.New(fmt.Sprintf("%v", r))
			}
		}
	}()

	// Call function to manage this message
	if f, ok := s.callbacks[packet.Type()]; ok {
		f(packet, message, session)
	} else {
		err = ErrUnregisteredMessage
	}

	return
}

/*map[
	object:user
	entry:[
				map[uid:100212827024403
						id:100212827024403
						time:1.451828949e+09
						changed_fields:[friends]]
				map[uid:101253640253369
						id:101253640253369
						time:1.451828949e+09
						changed_fields:[friends]]
	 ]
]*/
func (s *Server) onFacebookUpdate(updateInfo *wh.FacebookUpdate) {

	if updateInfo.Object != "user" {
		log.Println("onFacebookUpdate: Invalid update type")
		return
	}

	for _, entry := range updateInfo.Entries {

		user, err := s.Model.Accounts.GetUserAccountByFacebook(entry.Id)
		if err != nil {
			log.Printf("onFacebookUpdate Error: FbId -> %v, Err -> %v\n", entry.Id, err)
			continue
		}

		for _, changedField := range entry.ChangedFields {
			if changedField == "friends" {
				s.task_executor.Submit(&ImportFacebookFriends{
					TargetUser: user,
					Fbtoken:    user.FbToken(),
				})
			}
		} // End inner loop
	} // End outter loop
}

func (s *Server) getSession(userID int64) *AyiSession {
	if session, ok := s.sessions.Get(userID); ok {
		return session
	}
	return nil
}

func checkAuthenticated(session *AyiSession) {
	if !session.IsAuth {
		panic(ErrAuthRequired)
	}
}

func checkUnauthenticated(session *AyiSession) {
	if session.IsAuth {
		panic(ErrNoAuthRequired)
	}
}

func checkNoErrorOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func checkAtLeastOneEventOrPanic(events []*model.Event) {
	if len(events) == 0 {
		panic(ErrEventNotFound)
	}
}

func checkEventWritableOrPanic(event *model.Event) {
	if event.Status() != api.EventState_NOT_STARTED {
		panic(model.ErrEventNotWritable)
	}
}

func checkEventAuthorOrPanic(authorID int64, event *model.Event) {
	if event.AuthorId() != authorID {
		panic(ErrAuthorMismatch)
	}
}
