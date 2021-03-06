package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/d3ce1t/areyouin-server/api"
	fb "github.com/d3ce1t/areyouin-server/facebook"
	"github.com/d3ce1t/areyouin-server/model"
	proto "github.com/d3ce1t/areyouin-server/protocol"
	wh "github.com/d3ce1t/areyouin-server/webhook"
)

type Callback func(*proto.AyiPacket, proto.Message, *AyiSession)

const (
	SERVER_VERSION = "1.0.3"
)

var (
	BUILD_TIME = "" // Filled by build.sh
)

type Server struct {
	TLSConfig     *tls.Config
	sessions      *SessionsMap
	callbacks     map[proto.PacketType]Callback
	Model         *model.AyiModel
	modelObserver *ModelObserver
	DbSession     api.DbSession
	webhook       *wh.WebHookServer
	Config        api.Config
	version       string
	buildTime     string
}

func NewServer(session api.DbSession, model *model.AyiModel, config api.Config) *Server {
	server := &Server{
		DbSession: session,
		Model:     model,
		Config:    config,
		version:   SERVER_VERSION,
		buildTime: BUILD_TIME,
	}
	server.init()
	return server
}

func (s *Server) Version() string {
	return s.version
}

func (s *Server) BuildTime() string {
	return s.buildTime
}

// Setup server components
func (s *Server) init() {

	// Init TLS config
	cert, err := tls.LoadX509KeyPair(s.Config.CertFile(), s.Config.CertKey())
	if err != nil {
		panic(err)
	}

	s.TLSConfig = &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{cert},
		ServerName:   s.Config.DomainName(),
	}

	// Init sessions holder
	s.sessions = NewSessionsMap()
	s.callbacks = make(map[proto.PacketType]Callback)
}

func (s *Server) bootstrapServer() {

	log.Println("Bootstraping server...")

	err := s.Model.Events.BuildEventsTimeLine()
	if err != nil {
		fmt.Println(err)
	}

	err = s.Model.Events.BuildEventsHistory()
	if err != nil {
		fmt.Println(err)
	}

	log.Println("Bootstraping done")
}

func (s *Server) run(bootstrap bool) {

	if !s.Config.MaintenanceMode() {

		// Start webhook

		if s.Config.FBWebHookEnabled() {
			s.webhook = wh.New(fb.FB_APP_SECRET, s.Config)
			s.webhook.RegisterCallback(s.onFacebookUpdate)
			s.webhook.Run()
		}

		// Start model background tasks

		s.Model.StartBackgroundTasks()

		// Start up model observer

		s.modelObserver = newModelObserver(s)
		go s.modelObserver.run()

	} else {

		log.Println("Server running in MAINTENANCE MODE")

		if bootstrap {
			go s.bootstrapServer()
		}
	}

	// Start up server listener

	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", s.Config.ListenAddress(), s.Config.ListenPort()))
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
	session.OnRead = func(s *AyiSession, recvPacket *proto.AyiPacket) {

		packetToken := recvPacket.Header.GetToken()

		if !s.Server.Config.MaintenanceMode() {

			// Normal operation

			if err := s.Server.serveMessage(recvPacket, s); err != nil { // may block until writes are performed
				errorCode := getNetErrorCode(err, proto.E_OPERATION_FAILED)
				log.Printf("< (%v) ERROR %v: %v\n", session, errorCode, err)
				s.WriteResponse(packetToken, s.NewMessage().Error(recvPacket.Type(), errorCode))
			}

		} else {

			// Maintenance Mode

			log.Printf("> (%v) Packet %v received but ignored\n", session, recvPacket.Type())
			s.WriteResponseSync(packetToken, s.NewMessage().Error(recvPacket.Type(), proto.E_SERVER_MAINTENANCE))
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

	// Called after IDLE_INTERVAL time of inactivity
	session.OnIdle = func(session *AyiSession) {

		currentTime := time.Now()

		if !session.IsAuth {

			if currentTime.After(session.lastRecvMsg.Add(MAX_LOGIN_TIME)) {
				log.Printf("* (%v) Connection IDLE\n", session)
				session.Exit()
			}

		} else {

			if currentTime.After(session.lastRecvMsg.Add(MAX_IDLE_TIME)) {

				log.Printf("* (%v) Connection IDLE\n", session)
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

				go func() {
					addedFriends, err := s.Model.Friends.ImportFacebookFriends(user, false)
					if err != nil {
						log.Printf("* IMPORT FACEBOOK FRIENDS (userID: %v) ERROR: %v", user.Id(), err)
						return
					}
					log.Printf("* IMPORT FACEBOOK FRIENDS SUCCESS (userID: %v, added: %v)", user.Id(), len(addedFriends))
				}()
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
		panic(ErrUnauthorized)
	}
}

func checkUnauthenticated(session *AyiSession) {
	if session.IsAuth {
		panic(ErrForbidden)
	}
}

func checkNoErrorOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func checkAccessOrPanic(userID int64, event *model.Event) {
	if event.AuthorID() != userID {
		panic(ErrForbidden)
	}
}
