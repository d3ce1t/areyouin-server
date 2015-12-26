package main

import (
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	fb "github.com/huandu/facebook"
	"log"
	"net"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	proto "peeple/areyouin/protocol"
	"time"
)

const (
	ALL_CONTACTS_GROUP = 0 // Id for the main friend group of a user
	MAX_WRITE_TIMEOUT  = 15 * time.Second
	//MAX_READ_TIMEOUT   = 1 * time.Second
	MAX_IDLE_TIME  = 30 * 60 * time.Second // 30m
	MAX_LOGIN_TIME = 30 * time.Second      // 30s
)

func NewServer() *Server {
	server := &Server{Keyspace: "areyouin"}
	server.init()
	return server
}

func NewTestServer() *Server {
	server := &Server{Keyspace: "areyouin_demo"}
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
	s.cluster = gocql.NewCluster("192.168.1.2" /*"192.168.1.3"*/)
	s.cluster.Keyspace = s.Keyspace
	s.cluster.Consistency = gocql.LocalQuorum
	s.connectToDB()

	// Task Executor
	s.task_executor = NewTaskExecutor(s)
	s.task_executor.Start()

	// Start Event Delivery
	s.ds = NewDeliverySystem(s)
	s.ds.Run()
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
	}

	s.sessions[session.UserId] = session
}

func (s *Server) UnregisterSession(session *AyiSession) {
	user_id := session.UserId
	session.Close()
	delete(s.sessions, user_id)
}

func (s *Server) NewUserDAO() core.UserDAO {
	return dao.NewUserDAO(s.dbsession)
}

func (s *Server) NewEventDAO() core.EventDAO {
	return dao.NewEventDAO(s.dbsession)
}

func (s *Server) AddFriend(user_id uint64, friend *core.Friend) error {
	return dao.NewUserDAO(s.dbsession).AddFriend(user_id, friend, ALL_CONTACTS_GROUP)
}

// Private methods
func (s *Server) handleSession(session *AyiSession) {

	// Defer session close
	defer func() {

		defer func() { // SetLastConnection may also panic
			if r := recover(); r != nil {
				log.Printf("Session %v Defer Panic: %v\n", session, r)
			}
			s.UnregisterSession(session)
			log.Println("Session closed:", session)
		}()

		if r := recover(); r != nil {
			log.Printf("Session %v Panic: %v\n", session, r)
		}
		last_connection := core.GetCurrentTimeMillis()
		session.Server.NewUserDAO().SetLastConnection(session.UserId, last_connection)
	}()

	log.Println("New connection from", session)

	exit := false

	for !exit {

		select {
		// Send Notifications
		case notification := <-session.NotificationChannel:
			session.ProcessNotification(notification)
			continue

		// Read messages
		case packet := <-session.SocketChannel:
			session.lastRecvMsg = time.Now().UTC()
			if err := s.serveMessage(packet, session); err != nil { // may block until writes are performed
				log.Println("Error:", err)
				log.Println("Involved packet:", packet)
			}

		// Manage errors
		case err := <-session.SocketError:
			if err == proto.ErrConnectionClosed {
				log.Println("Connection closed by client:", session)
				exit = true
			} else if err != proto.ErrTimeout {
				if session.isClosed {
					exit = true
				}
				log.Println("Error:", err)
			}

		default:
			current_time := time.Now().UTC()
			if !session.IsAuth {
				if current_time.After(session.lastRecvMsg.Add(MAX_LOGIN_TIME)) {
					session.lastRecvMsg = time.Now().UTC()
					log.Println("Connection IDLE", session)
					exit = true
				}
			} else {
				if session.IsAuth && current_time.After(session.lastRecvMsg.Add(MAX_IDLE_TIME)) {
					session.lastRecvMsg = time.Now().UTC()
					log.Println("Connection IDLE", session)
					exit = true
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
		if ok, err := dao.Insert(event); ok {
			if err := dao.AddOrUpdateParticipants(event.EventId, participants); err == nil {
				event.NumGuests = int32(len(participants))
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

func (s *Server) createParticipantsList(author_id uint64, participants_id []uint64) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants_id))

	dao := s.NewUserDAO()

	// TODO: Optimise this path
	for _, user_id := range participants_id {
		if ok, _ := dao.AreFriends(author_id, user_id); ok {
			if uac, _ := dao.Load(user_id); uac != nil {
				result = append(result, uac.AsParticipant())
			} else {
				log.Println("createParticipantList() participant", user_id, "does not exist")
			}
		} else {
			log.Println("createParticipantList() Not friends", author_id, "and", user_id, "or doesn't exist")
		}
	}

	return result
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

func sendUserFriends(session *AyiSession) {

	server := session.Server
	dao := server.NewUserDAO()

	friends, err := dao.LoadFriends(session.UserId, ALL_CONTACTS_GROUP)

	if err == nil {
		reply := proto.NewMessage().FriendsList(friends).Marshal()
		log.Println("SEND USER FRIENDS to", session)
		session.WriteReply(reply)
	}
}

// Called from multiple threads
func sendPrivateEvents(session *AyiSession) {

	server := session.Server
	dao := server.NewEventDAO()
	events, err := dao.LoadUserEvents(session.UserId, core.GetCurrentTimeMillis())

	if err != nil {
		log.Println("sendPrivateEvents()", err)
		return
	}

	if len(events) > 0 {

		// Send events list to user
		log.Println("SEND PRIVATE EVENTS to", session)
		session.WriteReply(proto.NewMessage().EventsList(events).Marshal())

		// Send participants info of each event and update participant status as delivered
		for _, event := range events {
			event_participants, _ := dao.LoadAllParticipants(event.EventId)
			event_participants = session.Server.filterParticipants(session.UserId, event_participants)
			msg := proto.NewMessage().AttendanceStatus(event.EventId, event_participants).Marshal()
			session.WriteReply(msg)
			// FIXME: Probably could do so in only one operation
			dao.SetParticipantStatus(session.UserId, event.EventId, core.MessageStatus_CLIENT_DELIVERED)
		}
	}
}

func sendAuthError(session *AyiSession) {
	session.WriteReply(proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD).Marshal())
	log.Println("SEND INVALID USER OR PASSWORD")
}

func checkFacebookAccess(id string, access_token string) (fbaccount *FacebookAccount, ok bool) {

	// Contact Facebook
	res, err := fb.Get("/me", fb.Params{
		"fields":       "id,name,email",
		"access_token": access_token,
	})

	if err != nil {
		log.Println("Server error when connecting to Facebook", err)
		return nil, false
	}

	// Get info
	account := &FacebookAccount{}

	if fbid, ok := res["id"]; ok {
		account.id = fbid.(string)
	}

	if name, ok := res["name"]; ok {
		account.name = name.(string)
	}

	if email, ok := res["email"]; ok {
		account.email = email.(string)
	}

	if account.id != id {
		log.Println("Fbid does not match provided User ID")
		return account, false
	} else {
		return account, true
	}
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
		p2.Response == core.AttendanceResponse_CANNOT_ASSIST ||
		p1 == p2.UserId {
		return true
	} else if ok, _ := dao.AreFriends(p1, p2.UserId); ok {
		return true
	}
	return false
}

func main() {

	server := NewServer() // Server is global

	if server.DbSession() != nil {
		core.CreateFakeUsers(server.NewUserDAO())
	}
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
