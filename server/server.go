package main

import (
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"net"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
	proto "peeple/areyouin/protocol"
	wh "peeple/areyouin/webhook"
)

const (
	ALL_CONTACTS_GROUP = 0 // Id for the main friend group of a user
	//MAX_READ_TIMEOUT   = 1 * time.Second
)

func NewServer() *Server {
	server := &Server{
		DbAddress: "192.168.1.2",
		Keyspace:  "areyouin",
	}
	server.init()
	return server
}

func NewTestServer() *Server {

	fmt.Println("----------------------------------------")
	fmt.Println("! WARNING WARNING WARNING              !")
	fmt.Println("! You have started a testing server    !")
	fmt.Println("! WARNING WARNING WARNING              !")
	fmt.Println("----------------------------------------")

	server := &Server{
		DbAddress: "192.168.1.10",
		Keyspace:  "areyouin",
		Testing:   true,
	}
	server.init()
	return server
}

type Callback func(proto.PacketType, proto.Message, *AyiSession)

type Server struct {
	sessions      *SessionsMap
	task_executor *TaskExecutor
	id_gen_ch     chan uint64
	callbacks     map[proto.PacketType]Callback
	id_generators map[uint16]*core.IDGen
	cluster       *gocql.ClusterConfig
	dbsession     *gocql.Session
	Keyspace      string
	webhook       *wh.WebHookServer
	DbAddress     string
	serialChannel chan func()
	Testing       bool
}

func (s *Server) DbSession() *gocql.Session {
	return s.dbsession
}

// Setup server components
func (s *Server) init() {

	s.sessions = NewSessionsMap()
	s.callbacks = make(map[proto.PacketType]Callback)

	// Serial execution
	s.serialChannel = make(chan func(), 8)
	go func() {
		for f := range s.serialChannel {
			f()
		}
	}()

	// ID generator
	s.id_generators = make(map[uint16]*core.IDGen)
	s.id_gen_ch = make(chan uint64)
	go s.generatorTask(1)

	// Connect to Cassandra
	s.cluster = gocql.NewCluster(s.DbAddress /*"192.168.1.3"*/)
	s.cluster.Keyspace = s.Keyspace
	s.cluster.Consistency = gocql.LocalQuorum
	s.connectToDB()

	// Task Executor
	s.task_executor = NewTaskExecutor(s)
	s.task_executor.Start()
}

func (s *Server) connectToDB() {
	if session, err := s.cluster.CreateSession(); err == nil {
		s.dbsession = session
	} else {
		log.Println("Error connecting to cassandra:", err)
		return
	}
}

func (s *Server) executeSerial(f func()) {
	s.serialChannel <- f
}

func (s *Server) Run() {

	// Start webhook
	s.webhook = wh.New(fb.FB_APP_SECRET)
	s.webhook.RegisterCallback(s.onFacebookUpdate)
	s.webhook.Run()

	// Start up server listener
	listener, err := net.Listen("tcp", ":1822")

	if err != nil {
		panic("Couldn't start listening: " + err.Error())
	}

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

func (s *Server) GetNewID() uint64 {
	return <-s.id_gen_ch
}

func (s *Server) RegisterCallback(command proto.PacketType, f Callback) {
	if s.callbacks == nil {
		s.callbacks = make(map[proto.PacketType]Callback)
	}
	s.callbacks[command] = f
}

func (s *Server) RegisterSession(session *AyiSession) {
	s.executeSerial(func() {
		if oldSession, ok := s.sessions.Get(session.UserId); ok {
			oldSession.WriteSync(proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD).Marshal())
			log.Printf("< (%v) SEND INVALID USER OR PASSWORD\n", oldSession)
			oldSession.Exit()
			log.Printf("Closing old session %v for user %v\n", oldSession, oldSession.UserId)
		}
		s.sessions.Put(session.UserId, session)
		log.Printf("Register session %v for user %v\n", session, session.UserId)
	})
}

func (s *Server) UnregisterSession(session *AyiSession) {
	s.executeSerial(func() {
		user_id := session.UserId
		if !session.IsClosed() {
			session.Exit()
		}
		oldSession, ok := s.sessions.Get(user_id)
		if ok && oldSession == session {
			s.sessions.Remove(user_id)
			log.Printf("Unregister session %v for user %v\n", session, user_id)
		}
	})
}

func (s *Server) NewUserDAO() core.UserDAO {
	if s.dbsession == nil {
		s.connectToDB()
	}
	return dao.NewUserDAO(s.dbsession)
}

func (s *Server) NewEventDAO() core.EventDAO {
	if s.dbsession == nil {
		s.connectToDB()
	}
	return dao.NewEventDAO(s.dbsession)
}

// Private methods
func (server *Server) handleSession(session *AyiSession) {

	defer func() { // session.RunLoop() may throw panic
		if r := recover(); r != nil {
			log.Printf("Session %v Panic: %v\n", session, r)
		}
	}()

	log.Println("New connection from", session)

	session.OnRead = func(s *AyiSession, packet *proto.AyiPacket) {
		if err := s.Server.serveMessage(packet, s); err != nil { // may block until writes are performed
			log.Println("ServeMessage Panic:", err)
			log.Println("Involved packet:", packet)
			s.Write(proto.NewMessage().Error(packet.Type(), proto.E_OPERATION_FAILED).Marshal())
		}
	}

	session.OnError = func(s *AyiSession, err error) {
		if err != proto.ErrTimeout {
			log.Println("Session Error:", err)
		}
	}

	session.OnClosed = func(s *AyiSession, peer bool) {
		if s.IsAuth {
			//NOTE: If a user is deleted from user_account while it is still connected,
			// a row in invalid state will be created when updating last connection
			s.Server.NewUserDAO().SetLastConnection(s.UserId, core.GetCurrentTimeMillis())
		}

		s.Server.UnregisterSession(s)

		if peer {
			log.Printf("Session closed by client: %v %v\n", s.UserId, s)
		} else {
			log.Printf("Session closed %v %v\n", s.UserId, s)
		}
	}

	session.RunLoop() // Block here
}

func (s *Server) getIDGenerator(id uint16) *core.IDGen {

	generator, ok := s.id_generators[id]

	if !ok {
		generator = core.NewIDGen(id)
		s.id_generators[id] = generator
	}

	return generator
}

func (s *Server) generatorTask(gid uint16) {

	gen := s.getIDGenerator(gid)

	for {
		newId := gen.GenerateID()
		s.id_gen_ch <- newId
	}
}

func (s *Server) serveMessage(packet *proto.AyiPacket, session *AyiSession) (err error) {

	message := packet.DecodeMessage()
	err = nil

	if message != nil {

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
			f(packet.Type(), message, session)
		} else {
			err = ErrUnhandledMessage
		}

	} else {
		err = ErrUnknownMessage
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

	userDao := s.NewUserDAO()

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
					UserId: user.Id,
					Name:   user.Name,
					//Fbid:    user.Fbid,
					Fbtoken: user.Fbtoken,
				})
			}
		} // End inner loop
	} // End outter loop
}

func (s *Server) SendMessage(user_id uint64, message []byte) bool {
	session := s.GetSession(user_id)
	if session == nil {
		return false
	}
	return session.Write(message)
}

func (s *Server) GetSession(user_id uint64) *AyiSession {
	if session, ok := s.sessions.Get(user_id); ok {
		return session
	} else {
		return nil
	}
}

// Insert an event into database, add participants to it and send it to users' inbox.
func (s *Server) PublishEvent(event *core.Event) error {

	eventDAO := s.NewEventDAO()

	if len(event.Participants) <= 1 { // I need more than the only author
		return ErrParticipantsRequired
	}

	if err := eventDAO.InsertEventAndParticipants(event); err != nil {
		return err
	}

	// NOTE: DeliverySystem Submit must be persistent in order to continue the job
	// in case of failure. So, in the meanwhile publishing events to inbox will be
	// managed here. Only notification will be managed async.
	//s.ds.Submit(event) // put event into users' inbox

	// Author is the last participant. Add it first in order to get the author receive
	// the event to add other participants if something fails
	tmp_error := s.inviteParticipantToEvent(event, event.Participants[event.AuthorId])
	if tmp_error != nil {
		log.Println("PublishEvent Error:", tmp_error)
		return ErrAuthorDeliveryError
	}

	for k, participant := range event.Participants {
		if k != event.AuthorId {
			s.inviteParticipantToEvent(event, participant) // TODO: Implement retry but do not return on error
		}
	}

	notification := &NotifyEventInvitation{
		Event: event,
	}

	s.task_executor.Submit(notification)

	return nil
}

func (s *Server) inviteParticipantToEvent(event *core.Event, participant *core.EventParticipant) error {

	eventDAO := s.NewEventDAO()

	if err := eventDAO.AddEventToUserInbox(participant.UserId, event); err != nil {
		log.Println("Coudn't deliver event", event.EventId, err)
		return err
	}

	participant.Delivered = core.MessageStatus_SERVER_DELIVERED
	log.Println("Event", event.EventId, "delivered to user", participant.UserId)
	return nil
}

// This function is used to create a participant list that will be added to an event.
// This event will be published on behalf of the author. By this reason, participants
// can only be current friends of the author. In this code, it is assumed that
// participants are already friends of the author (author has-friend way). However,
// it must be checked if participants have also the author as a friend (friend has-author way)
func (s *Server) createParticipantsList(author_id uint64, participants_id []uint64) (participant map[uint64]*core.EventParticipant, warn error, err error) {

	var last_warning error
	result := make(map[uint64]*core.EventParticipant)
	userDAO := s.NewUserDAO()

	for _, p_id := range participants_id {

		if ok, err := userDAO.IsFriend(p_id, author_id); ok {

			if uac, err := userDAO.Load(p_id); err == dao.ErrNotFound { // FIXME: Load several participants in one operation
				last_warning = ErrUnregisteredFriendsIgnored
				log.Printf("createParticipantList() Warning at userid %v: %v\n", p_id, err)
			} else if err != nil {
				return nil, nil, err
			} else {
				result[uac.Id] = uac.AsParticipant()
			}

		} else if err != nil {
			return nil, nil, err
		} else {
			log.Println("createParticipantList() Not friends", author_id, "and", p_id)
			last_warning = ErrNonFriendsIgnored
		}
	}

	return result, last_warning, nil
}

func (s *Server) createParticipantsFromFriends(author_id uint64) []*core.EventParticipant {

	dao := s.NewUserDAO()
	friends, _ := dao.LoadFriends(author_id, ALL_CONTACTS_GROUP)

	if friends != nil {
		return core.CreateParticipantsFromFriends(author_id, friends)
	} else {
		log.Println("createParticipantsFromFriends() no friends or error")
		return nil
	}
}

// Called from multiple threads
// TODO: Update to send events + participants in one single message
func sendPrivateEvents(session *AyiSession) {

	server := session.Server
	dao := server.NewEventDAO()
	events, err := dao.LoadUserEventsAndParticipants(session.UserId, core.GetCurrentTimeMillis())

	if err != nil {
		log.Printf("sendPrivateEvents() to %v Error: %v\n", session.UserId, err)
		return
	}

	// For compatibility, split events into event info and participants
	half_events := make([]*core.Event, 0, len(events))
	for _, event := range events {
		half_events = append(half_events, event.GetEventWithoutParticipants())
	}

	// Send events list to user
	log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session.UserId, len(half_events))
	session.Write(proto.NewMessage().EventsList(half_events).Marshal()) // TODO: Change after remove compatibility

	// Send participants info of each event, update participant status as delivered and notify
	for _, event := range events {

		participants_filtered := session.Server.filterParticipantsMap(session.UserId, event.Participants)

		// Send attendance info
		msg := proto.NewMessage().AttendanceStatus(event.EventId, participants_filtered).Marshal()
		session.Write(msg)

		// Update participant status of the session user
		ownParticipant := event.Participants[session.UserId]

		if ownParticipant.Delivered != core.MessageStatus_CLIENT_DELIVERED {
			ownParticipant.Delivered = core.MessageStatus_CLIENT_DELIVERED
			dao.SetParticipantStatus(session.UserId, event.EventId, ownParticipant.Delivered)

			// Notify change in participant status to the other participants
			task := &NotifyParticipantChange{
				Event:               event,
				ParticipantsChanged: []uint64{session.UserId},
			}

			// I'm also sending notification to the author. Could avoid this because author already knows
			// that the event has been send to him
			server.task_executor.Submit(task)
		}
	}
}

func sendAuthError(session *AyiSession) {
	session.Write(proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD).Marshal())
	log.Printf("< (%v) SEND INVALID USER OR PASSWORD\n", session)
}

func checkAuthenticated(session *AyiSession) {
	if !session.IsAuth {
		panic(ErrAuthRequired)
	}
}

func checkUnauthenticated(session *AyiSession) {
	if session.IsAuth {
		panic(ErrAuthRequired)
	}
}

// Returns a participant list where users that will not assist the event or aren't
// friends of the given user are removed */
func (s *Server) filterParticipantsMap(participant uint64, participants map[uint64]*core.EventParticipant) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants))

	for _, p := range participants {
		if s.canSee(participant, p) {
			result = append(result, p)
		} else {
			result = append(result, p.AsAnonym())
		}
	}

	return result
}

func (s *Server) filterParticipantsSlice(participant uint64, participants []*core.EventParticipant) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants))

	for _, p := range participants {
		if s.canSee(participant, p) {
			result = append(result, p)
		} else {
			result = append(result, p.AsAnonym())
		}
	}

	return result
}

// Tells if participant p1 can see changes of participant p2
func (s *Server) canSee(p1 uint64, p2 *core.EventParticipant) bool {
	dao := s.NewUserDAO()
	if p2.Response == core.AttendanceResponse_ASSIST ||
		/*p2.Response == core.AttendanceResponse_CANNOT_ASSIST ||*/
		p1 == p2.UserId {
		return true
	} else if ok, _ := dao.IsFriend(p2.UserId, p1); ok {
		return true
	}
	return false
}

/*func createFbTestUsers() {
	fb.CreateTestUser("Test User One", true)
	fb.CreateTestUser("Test User Two", true)
	fb.CreateTestUser("Test User Three", true)
	fb.CreateTestUser("Test User Four", true)
	fb.CreateTestUser("Test User Five", true)
	fb.CreateTestUser("Test User Six", true)
	fb.CreateTestUser("Test User Seven", true)
	fb.CreateTestUser("Test User Eight", true)
}*/

func main() {

	server := NewTestServer()
	//server := NewServer() // Server is global
	//createFbTestUsers()
	/*if server.DbSession() != nil {
		core.AddFriendsToFbTestUserOne(server.NewUserDAO())
		//core.ClearUserAccounts(server.DbSession())
		//core.CreateFakeUsers(server.NewUserDAO())
	}*/
	server.RegisterCallback(proto.M_PING, onPing)
	server.RegisterCallback(proto.M_PONG, onPong)
	server.RegisterCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
	server.RegisterCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
	server.RegisterCallback(proto.M_USER_AUTH, onUserAuthentication)
	server.RegisterCallback(proto.M_CREATE_EVENT, onCreateEvent)
	server.RegisterCallback(proto.M_USER_FRIENDS, onUserFriends)
	server.RegisterCallback(proto.M_CONFIRM_ATTENDANCE, onConfirmAttendance)

	shell := NewShell(server)
	go shell.StartTermSSH()

	server.Run() // start server loop
}
