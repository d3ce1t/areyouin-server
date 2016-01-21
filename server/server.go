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
	"time"
)

const (
	ALL_CONTACTS_GROUP = 0 // Id for the main friend group of a user
	//MAX_READ_TIMEOUT   = 1 * time.Second
	MAX_IDLE_TIME    = 30 * time.Minute // 30m
	MAX_LOGIN_TIME   = 30 * time.Second // 30s
	PING_INTERVAL_MS = 29 * time.Minute // 29m
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

	fmt.Println("---------------------------------------!")
	fmt.Println("! WARNING WARNING WARNING              !")
	fmt.Println("! You have started a testing server    !")
	fmt.Println("! WARNING WARNING WARNING              !")
	fmt.Println("----------------------------------------")

	server := &Server{
		DbAddress: "192.168.1.10",
		Keyspace:  "areyouin",
	}
	server.init()
	return server
}

type Callback func(proto.PacketType, proto.Message, *AyiSession)

type Server struct {
	sessions      map[uint64]*AyiSession
	task_executor *TaskExecutor
	ds            *DeliverySystem
	id_gen_ch     chan uint64
	callbacks     map[proto.PacketType]Callback
	id_generators map[uint16]*core.IDGen
	cluster       *gocql.ClusterConfig
	dbsession     *gocql.Session
	Keyspace      string
	webhook       *wh.WebHookServer
	DbAddress     string
}

func (s *Server) DbSession() *gocql.Session {
	return s.dbsession
}

// Setup server components
func (s *Server) init() {

	s.sessions = make(map[uint64]*AyiSession)
	s.callbacks = make(map[proto.PacketType]Callback)

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

	// Start Event Delivery
	s.ds = NewDeliverySystem(s)
	s.ds.Start()
}

func (s *Server) connectToDB() {
	if session, err := s.cluster.CreateSession(); err == nil {
		s.dbsession = session
	} else {
		log.Println("Error connecting to cassandra:", err)
		return
	}
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

	if oldSession, ok := s.sessions[session.UserId]; ok {
		s.UnregisterSession(oldSession)
		oldSession.Close()
	}

	s.sessions[session.UserId] = session
}

func (s *Server) UnregisterSession(session *AyiSession) {
	user_id := session.UserId
	if !session.IsClosed {
		session.Close()
	}
	delete(s.sessions, user_id)
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
func (s *Server) handleSession(session *AyiSession) {

	// Defer session close
	defer func() {

		defer func() { // updateLastconnection may also panic
			if r := recover(); r != nil {
				log.Printf("Session %v Defer Panic: %v\n", session, r)
			}
			s.UnregisterSession(session)
			log.Println("Session closed:", session)
		}()

		if r := recover(); r != nil {
			log.Printf("Session %v Panic: %v\n", session, r)
		}

		session.updateLastConnection()
	}()

	log.Println("New connection from", session)

	pingTime := time.Now().Add(PING_INTERVAL_MS)

	exit := false

	for !session.IsClosed && !exit {

		select {
		// Send Notifications
		case notification := <-session.NotificationChannel:
			session.ProcessNotification(notification)
			continue

		// Read messages
		case packet := <-session.SocketChannel:
			session.lastRecvMsg = time.Now()
			pingTime = session.lastRecvMsg.Add(PING_INTERVAL_MS)
			if err := s.serveMessage(packet, session); err != nil { // may block until writes are performed
				log.Println("ServeMessage Panic:", err)
				log.Println("Involved packet:", packet)
				session.WriteReply(proto.NewMessage().Error(packet.Type(), proto.E_OPERATION_FAILED).Marshal())
			}

		// Manage errors
		case err := <-session.SocketError:
			if err == proto.ErrConnectionClosed {
				log.Println("Connection closed by client:", session)
				exit = true
			} else if err != proto.ErrTimeout {
				log.Println("Error:", err)
			}

		default:
			current_time := time.Now()
			if !session.IsAuth {
				if current_time.After(session.lastRecvMsg.Add(MAX_LOGIN_TIME)) {
					session.lastRecvMsg = time.Now()
					log.Println("Connection IDLE", session)
					exit = true
				}
			} else {
				if current_time.After(session.lastRecvMsg.Add(MAX_IDLE_TIME)) {
					session.lastRecvMsg = time.Now()
					log.Println("Connection IDLE", session)
					exit = true
				} else if current_time.After(pingTime) {
					session.SendPing()
					pingTime = time.Now().Add(18 * time.Second)
					log.Println("< PING to", session)
				}
			}

			time.Sleep(250 * time.Millisecond)

		} // End select

	} // End loop
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

func (s *Server) notifyUser(user_id uint64, message []byte, callback func()) {
	if session, ok := s.sessions[user_id]; ok {
		session.Notify(&Notification{
			Message:  message,
			Callback: callback,
		})
	}
}

/*
Insert an event into database, add participants to it and send it to users' inbox.
NOTE: This function isn't thread-safe
*/
func (s *Server) PublishEvent(event *core.Event, participants []*core.EventParticipant) bool {

	result := false
	dao := s.NewEventDAO()

	if len(participants) > 0 {
		// FIXME: Insert uses lightweight-transaction but actually may be not needed because
		// EventID (primary key) is unique if, and only if, IDGen ID do not overlap with
		// others IDGen running concurrently. In other words, if each IDGen produces keys
		// of its assigned space, then EventID is unique.
		if ok, err := dao.InsertEventCAS(event); ok {
			if err := dao.AddOrUpdateParticipants(event.EventId, participants); err == nil {
				event.NumGuests = int32(len(participants))
				// Should use Compare-and-set version of SetNumGuests
				if err := dao.SetNumGuests(event.EventId, event.NumGuests); err != nil {
					log.Println("PublishEvent", err)
				}
				// FIXME: DeliverySystem Submit must be persistent in order to continue the job
				// in case of failure
				s.ds.Submit(event) // put event into users' inbox
				result = true
			} else {
				log.Println("PublishEvent", err)
			}

		} else {
			log.Println("PublishEvent:", err)
		}
	} else {
		log.Println("Trying to publish an event with no participants")
	}

	return result
}

func (s *Server) createParticipantsList(author_id uint64, participants_id []uint64) ([]*core.EventParticipant, error) {

	var warning error
	result := make([]*core.EventParticipant, 0, len(participants_id))
	dao := s.NewUserDAO()

	// TODO: Optimise this path
	for _, user_id := range participants_id {
		if ok, _ := dao.AreFriends(author_id, user_id); ok {
			if uac, _ := dao.Load(user_id); uac != nil {
				result = append(result, uac.AsParticipant())
			} else {
				log.Println("createParticipantList() participant", user_id, "does not exist")
				warning = ErrUnregisteredFriendsIgnored
			}
		} else {
			log.Println("createParticipantList() Not friends", author_id, "and", user_id, "or doesn't exist")
			warning = ErrNonFriendsIgnored
		}
	}

	return result, warning
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
func sendPrivateEvents(session *AyiSession) {

	server := session.Server
	dao := server.NewEventDAO()
	events, err := dao.LoadUserEventsAndParticipants(session.UserId, core.GetCurrentTimeMillis())

	if err != nil {
		log.Println("sendPrivateEvents()", err)
		return
	}

	// For compatibility, split events in event info and participants
	participants_map := make(map[uint64][]*core.EventParticipant)
	client_status_map := make(map[uint64]*core.EventParticipant)

	for _, event := range events {
		participants_map[event.EventId] = event.Participants
		for _, p := range event.Participants {
			if p.UserId == session.UserId {
				client_status_map[event.EventId] = p
				break
			}
		}
		event.Participants = nil
	}

	if len(events) > 0 {

		// Send events list to user
		log.Println("SEND PRIVATE EVENTS to", session)
		session.WriteReply(proto.NewMessage().EventsList(events).Marshal())

		// Send participants info of each event, update participant status as delivered and notify
		for _, event := range events {

			participant := client_status_map[event.EventId]
			event_participants, _ := participants_map[event.EventId]
			participants_filtered := session.Server.filterParticipants(session.UserId, event_participants)

			// Send attendance info
			msg := proto.NewMessage().AttendanceStatus(event.EventId, participants_filtered).Marshal()
			session.WriteReply(msg)

			// Update participant status
			if participant.Delivered != core.MessageStatus_CLIENT_DELIVERED {

				dao.SetParticipantStatus(session.UserId, event.EventId, core.MessageStatus_CLIENT_DELIVERED) //FIXME: Probably could do it in only one operation

				// Notify change in participant status to the other participants
				task := &NotifyParticipantChange{
					EventId:  event.EventId,
					UserId:   session.UserId,
					Name:     participant.Name,
					Response: participant.Response,
					Status:   core.MessageStatus_CLIENT_DELIVERED,
				}

				task.AddParticipantsDst(event_participants) // I'm also sending notification to the author. Could avoid this because author already knows
				server.task_executor.Submit(task)           // that the event has been send to him
			}
		}
	}
}

func sendAuthError(session *AyiSession) {
	session.WriteReply(proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD).Marshal())
	log.Println("SEND INVALID USER OR PASSWORD")
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

/* Returns a participant list where users that will not assist the event or aren't
   friends of the given user are removed */
func (s *Server) filterParticipants(participant uint64, participants []*core.EventParticipant) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants))

	for _, p := range participants {
		// If the participant is a confirmed user (yes or cannot assist answer has been given)
		if s.canSee(participant, p) {
			result = append(result, p)
		} else {
			result = append(result, p.AsAnonym())
		}
	}

	return result
}

/*
 Tells if participant p1 can see changes of participant p2
*/
// FIXME: Maybe is better to cache this
func (s *Server) canSee(p1 uint64, p2 *core.EventParticipant) bool {
	dao := s.NewUserDAO()
	if p2.Response == core.AttendanceResponse_ASSIST ||
		/*p2.Response == core.AttendanceResponse_CANNOT_ASSIST ||*/
		p1 == p2.UserId {
		return true
	} else if ok, _ := dao.AreFriends(p1, p2.UserId); ok {
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

	//server := NewTestServer()
	server := NewServer() // Server is global
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

	shell := &Shell{Server: server}
	go shell.Execute()

	server.Run() // start server loop
}
