package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	"peeple/areyouin/model"
	fb "peeple/areyouin/facebook"
	proto "peeple/areyouin/protocol"
	wh "peeple/areyouin/webhook"
	"strings"
	"time"
)

const (
	GCM_API_KEY                = "AIzaSyAf-h1zJCRWNDt-dI3liL1yx4NEYjOq5GQ"
	GCM_MAX_TTL                = 2419200
	//MAX_READ_TIMEOUT   = 1 * time.Second
)

func NewServer(session core.DbSession, model *model.AyiModel) *Server {
	server := &Server{
		DbSession: session.(*dao.GocqlSession),
		Model: model,
	}
	server.init()
	return server
}

type Callback func(*proto.AyiPacket, proto.Message, *AyiSession)

type Server struct {
	TLSConfig     *tls.Config
	sessions      *SessionsMap
	task_executor *TaskExecutor
	callbacks     map[proto.PacketType]Callback
	Model         *model.AyiModel
	DbSession     *dao.GocqlSession
	webhook       *wh.WebHookServer
	serialChannel chan func()
	Testing       bool
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

	// Serial execution
	s.serialChannel = make(chan func(), 8)
	go func() {
		for f := range s.serialChannel {
			f()
		}
	}()

	// Task Executor
	s.task_executor = NewTaskExecutor(s)
	s.task_executor.Start()
}

func (s *Server) executeSerial(f func()) {
	s.serialChannel <- f
}

func (s *Server) Run() {

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

func (s *Server) RegisterCallback(command proto.PacketType, f Callback) {
	if s.callbacks == nil {
		s.callbacks = make(map[proto.PacketType]Callback)
	}
	s.callbacks[command] = f
}

func (s *Server) RegisterSession(session *AyiSession) {

	if oldSession, ok := s.sessions.Get(session.UserId); ok {
		oldSession.WriteSync(oldSession.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD))
		log.Printf("< (%v) SEND INVALID USER OR PASSWORD\n", oldSession)
		oldSession.Exit()
		log.Printf("* (%v) Closing old session for endpoint %v\n", oldSession, oldSession.Conn.RemoteAddr())
	}

	s.sessions.Put(session.UserId, session)
	log.Printf("* (%v) Register session for endpoint %v\n", session, session.Conn.RemoteAddr())
}

func (s *Server) UnregisterSession(session *AyiSession) {

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
func (server *Server) handleSession(session *AyiSession) {

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
				error_code := getNetErrorCode(err, proto.E_OPERATION_FAILED)
				log.Printf("< (%v) ERROR %v: %v\n", session, error_code, err)
				s.WriteResponse(packet.Header.GetToken(), s.NewMessage().Error(packet.Type(), error_code))
			}

		} else {

			// Maintenance Mode
			log.Printf("> (%v) Packet %v received but ignored\n", session, packet.Type())
			s.WriteSync(s.NewMessage().Error(packet.Type(), proto.E_SERVER_MAINTENANCE ))
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
	session.OnClosed = func(s *AyiSession, peer bool) {

		if s.IsAuth {
			//NOTE: If a user is deleted from user_account while it is still connected,
			// a row in invalid state will be created when updating last connection
			dao.NewUserDAO(s.Server.DbSession).SetLastConnection(s.UserId, core.GetCurrentTimeMillis())
			s.Server.UnregisterSession(s)
		}

		if peer {
			log.Printf("* (%v) Session closed by remote peer\n", s)
		} else {
			log.Printf("* (%v) Session closed by server\n", s)
		}
	}

	session.RunLoop() // Block here
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

	userDao := dao.NewUserDAO(s.DbSession)

	for _, entry := range updateInfo.Entries {

		user_id, err := userDao.GetIDByFacebookID(entry.Id)
		if err != nil {
			log.Printf("onFacebookUpdate Error 1 (%v): %v\n", user_id, err)
			continue
		}

		user, err := userDao.Load(user_id)
		if err != nil {
			log.Printf("onFacebookUpdate Error 2 (%v): %v\n", user_id, err)
			continue
		}

		for _, changedField := range entry.ChangedFields {
			if changedField == "friends" {
				s.task_executor.Submit(&ImportFacebookFriends{
					TargetUser: user,
					Fbtoken:    user.Fbtoken,
				})
			}
		} // End inner loop
	} // End outter loop
}

func (s *Server) GetSession(user_id int64) *AyiSession {
	if session, ok := s.sessions.Get(user_id); ok {
		return session
	} else {
		return nil
	}
}

// TODO: Remove. It's only being used for v0 and v1 compatibility
func sendPrivateEvents(session *AyiSession) {

	server := session.Server

	events, err := server.Model.Events.GetRecentEvents(session.UserId)
	if err != nil {
		log.Printf("sendPrivateEvents() to %v Error: %v\n", session, err)
		return
	}

	// Split events into event info and participants
	half_events := make([]*core.Event, 0, len(events))
	for _, event := range events {
		half_events = append(half_events, event.GetEventWithoutParticipants())
	}

	// Send events list to user
	log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session, len(half_events))
	session.Write(session.NewMessage().EventsList(half_events))


	for _, event := range events {

		// Send participants info for each event, update participant status as delivered and notify it

		if len(event.Participants) == 0 {
			log.Printf("SEND PRIVATE EVENTS WARNING: Event %v has zero participants\n", event.EventId)
			continue
		}

		// Send attendance info
		filteredParticipants := server.Model.Events.GetFilteredParticipants(event, session.UserId)
		session.Write(session.NewMessage().AttendanceStatus(event.EventId, filteredParticipants))

		// Update delivery state for session's user in this event
		changed, _ := server.Model.Events.ChangeDeliveryState(event, session.UserId, core.MessageStatus_CLIENT_DELIVERED)

		if changed {

			// Notify change in participant status to the other participants
			task := &NotifyParticipantChange{
				Event:               event,
				ParticipantsChanged: []int64{session.UserId},
				Target:              core.GetParticipantKeys(event.Participants),
			}

			// I'm also sending notification to the author. Could avoid this because author already knows
			// that the event has been send to him
			server.task_executor.Submit(task)
		}
	}
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

func checkAtLeastOneEventOrPanic(events []*core.Event) {
	if len(events) == 0 {
		panic(ErrEventNotFound)
	}
}

func checkEventWritableOrPanic(event *core.Event) {
	if event.GetStatus() != core.EventState_NOT_STARTED {
		panic(model.ErrEventNotWritable)
	}
}

func checkEventAuthorOrPanic(author_id int64, event *core.Event) {
	if event.AuthorId != author_id {
		panic(ErrAuthorMismatch)
	}
}

func (s *Server) isFriend(user1 int64, user2 int64) bool {
	ok, err := s.Model.Accounts.IsFriend(user2, user1)
	checkNoErrorOrPanic(err)
	return ok
}

// Sync server-side friends groups with client-side friends groups. If groups
// provided by the client contains all of the groups, then a full sync is required,
// i.e. server-side groups that does not exist client-side are removed. Otherwise,
// if provided groups are only a subset, a merge of client and server data is
// performed. Conversely to full sync, merging process does not remove existing
// groups from the server but add new groups and modify existing ones. Regarding
// full sync, it is assumed that clientGroups contains all of the groups in client.
// Hence, if a group doesn't exist in client, it will be removed from server. Like
// a regular sync, new groups in client will be added to server. In whatever case,
// if a group already exists server-side, it will be updated with members from client
// group, removing those members that does not exist in client group (client is master).
// In other words, groups at server will be equal to groups at client at the end of the
// synchornisation process.
func (s *Server) syncFriendGroups(owner int64, serverGroups []*core.Group,
	clientGroups []*core.Group, syncBehaviour core.SyncBehaviour) {

	friendsDAO := dao.NewFriendDAO(s.DbSession)

	// Copy map because it's gonna be modified
	clientGroupsCopy := make(map[int32]*core.Group)
	for _, group := range clientGroups {
		clientGroupsCopy[group.Id] = group
	}

	// Loop through server groups in order to know what
	// to do: update/replace or remove group from server
	for _, group := range serverGroups {

		if clientGroup, ok := clientGroupsCopy[group.Id]; ok {

			// Group exists.

			if clientGroup.Size == -1 && len(clientGroup.Members) == 0 {

				// Special case

				if clientGroup.Name == "" {

					// Group is marked for removal. So remove it from server

					err := friendsDAO.DeleteGroup(owner, group.Id)
					checkNoErrorOrPanic(err)

				} else if group.Name != clientGroup.Name {

					// Only Rename group
					err := friendsDAO.SetGroupName(owner, group.Id, clientGroup.Name)
					checkNoErrorOrPanic(err)

				}

			} else {

				// Update case

				if group.Name != clientGroup.Name {
					err := friendsDAO.SetGroupName(owner, group.Id, clientGroup.Name)
					checkNoErrorOrPanic(err)
				}

				s.syncGroupMembers(owner, group.Id, group.Members, clientGroup.Members)
			}

			// Delete also from copy because it has been processed
			delete(clientGroupsCopy, group.Id)

		} else if syncBehaviour == core.SyncBehaviour_TRUNCATE {

			// Remove

			err := friendsDAO.DeleteGroup(owner, group.Id)
			checkNoErrorOrPanic(err)
		}
	}

	// clientIndex contains only new groups. So add groups to server

	// Filter groups to remove non-friends.
	for _, group := range clientGroupsCopy {

		newMembers := make([]int64, 0, group.Size)

		for _, friendId := range group.Members {
			if s.isFriend(friendId, owner) {
				newMembers = append(newMembers, friendId)
			}
		}

		group.Members = newMembers
		group.Size = int32(len(newMembers))
	}

	for _, group := range clientGroupsCopy {
		err := friendsDAO.AddGroup(owner, group)
		checkNoErrorOrPanic(err)
	}
}

func (s *Server) syncGroupMembers(user_id int64, group_id int32, serverMembers []int64, clientMembers []int64) {

	// Index client members
	index := make(map[int64]bool)
	for _, id := range clientMembers {
		index[id] = true
	}

	// Loop through all members group owned by the server.
	// If member also exists client side, then keep it. Otherwise,
	// remove it.
	remove_ids := make([]int64, 0, len(serverMembers)/2)

	for _, serverMemberId := range serverMembers {
		if _, ok := index[serverMemberId]; !ok {
			remove_ids = append(remove_ids, serverMemberId)
		} else {
			delete(index, serverMemberId)
		}
	}

	// After removing already existing members. Index contains only new members,
	// so add them.
	friendDAO := dao.NewFriendDAO(s.DbSession)

	add_ids := make([]int64, 0, len(clientMembers))
	for id, _ := range index {
		if s.isFriend(id, user_id) {
			add_ids = append(add_ids, id)
		}
	}

	// Proceed database I/O
	if len(remove_ids) > 0 {
		err := friendDAO.DeleteMembers(user_id, group_id, remove_ids...)
		checkNoErrorOrPanic(err)
	}

	if len(add_ids) > 0 {
		err := friendDAO.AddMembers(user_id, group_id, add_ids...)
		checkNoErrorOrPanic(err)
	}
}
