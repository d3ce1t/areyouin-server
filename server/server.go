package main

import (
	proto "areyouin/protocol"
	"fmt"
	fb "github.com/huandu/facebook"
	"log"
	"net"
	"time"
)

const MAX_WRITE_TIMEOUT = 10 * time.Second
const MAX_READ_TIMEOUT = 2000 * time.Millisecond

func NewServer() *Server {
	server := &Server{}
	server.init()
	return server
}

type Callback func(proto.PacketType, proto.Message, *AyiSession)

type Server struct {
	sessions       map[uint64]*AyiSession
	udb            *UsersDatabase
	edb            *EventsDatabase
	ds             *DeliverySystem
	uid_ch         chan uint64
	callbacks      map[proto.PacketType]Callback
	uid_generators map[uint16]*UIDGen
}

// Public methods
func (s *Server) GetNewUserID() uint64 {
	return <-s.uid_ch
}

func (s *Server) RegisterCallback(command proto.PacketType, f Callback) {
	if s.callbacks == nil {
		s.callbacks = make(map[proto.PacketType]Callback)
	}
	s.callbacks[command] = f
}

// Setup server components
func (s *Server) init() {
	// Allocate space for components
	s.uid_generators = make(map[uint16]*UIDGen)
	s.sessions = make(map[uint64]*AyiSession)
	s.udb = NewUserDatabase()
	s.edb = NewEventDatabase()
	s.uid_ch = make(chan uint64)
	s.ds = NewDeliverySystem(s)
	s.callbacks = make(map[proto.PacketType]Callback)

	// Start UserID generator
	go s.generatorTask(1)

	// Start Event Delivery
	s.ds.Run()
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
			fmt.Println("Couldn't accept:", err.Error())
			continue
		}

		session := NewSession(client, s)
		go s.handleSession(session)
	}
}

func (s *Server) RegisterSession(session *AyiSession) {
	s.sessions[session.UserId] = session
}

func (s *Server) UnregisterSession(session *AyiSession) {
	session.Close()
	delete(s.sessions, session.UserId)
}

// Private methods

// TODO: Close connection if no login for a while (maybe 30 seconds)
// TODO: Close connection if no PING, PONG dialog (each 15 minutes?)
func (s *Server) handleSession(session *AyiSession) {

	log.Println("New connection from", session)

	var packet *proto.AyiPacket
	var read_error *proto.ReadError
	exit := false

	for !exit {

		select {
		// Send Notifications
		case notification := <-session.NotificationChannel:
			session.ProcessNotification(notification)

		// Read messages and then write a reply (if needed)
		default:
			session.Conn.SetReadDeadline(time.Now().Add(MAX_READ_TIMEOUT))
			if packet, read_error = proto.ReadPacket(session.Conn); read_error == nil {
				err := s.serveMessage(packet, session) // may block until writes are performed
				if err != nil {                        // Errors may happen
					log.Println("Unexpected error happened while serving message", err)
				}
			}
		} // End select

		// Manage possible error
		if read_error != nil && read_error.ConnectionClosed() {
			log.Println("Connection closed by client")
			exit = true
		} else if read_error != nil && !read_error.Timeout() {
			log.Println(read_error.Error())
		}

		//time.Sleep(100 * time.Millisecond)
	}

	log.Println("Session closed")

	if uac, ok := s.udb.GetByID(session.UserId); ok {
		uac.last_connection = time.Now().UTC().Unix()
	}

	s.UnregisterSession(session)
}

func (s *Server) getUIDGenerator(id uint16) *UIDGen {

	generator, ok := s.uid_generators[id]

	if !ok {
		generator = NewUIDGen(id)
		s.uid_generators[id] = generator
	}

	return generator
}

func (s *Server) generatorTask(gid uint16) {

	gen := s.getUIDGenerator(gid)

	for {
		newId := gen.GenerateID()
		//log.Println("New ID created!", newId)
		s.uid_ch <- newId
	}
}

func (s *Server) serveMessage(packet *proto.AyiPacket, session *AyiSession) error {

	message := packet.DecodeMessage()

	if message == nil {
		log.Fatal("Unknown message", packet)
		return nil
	}

	if f, ok := s.callbacks[packet.Type()]; ok {
		f(packet.Type(), message, session)
	}

	return nil
}

func (s *Server) sendUserFriends(client *AyiSession) bool {

	result := false

	if uac, ok := s.udb.GetByID(client.UserId); ok {
		friends := uac.GetAllFriends()
		friends_proto := make([]*proto.Friend, len(friends))
		for i := range friends {
			friends_proto[i] = &proto.Friend{
				UserId: friends[i].id,
				Name:   friends[i].name,
			}
		}
		reply := proto.NewMessage().FriendsList(friends_proto).Marshal()
		log.Println("SEND USER FRIENDS to", client)
		writeReply(reply, client)
		result = true
	} else {
		log.Println("SendUserFriends failed because of an invalid UserID")
		result = false
	}

	return result
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

func writeReply(reply []byte, session *AyiSession) {
	client := session.Conn
	client.SetWriteDeadline(time.Now().Add(MAX_WRITE_TIMEOUT))
	_, err := client.Write(reply)
	if err != nil {
		log.Println("Coudn't send reply: ", err)
	}
}

func (s *Server) notifyUser(user_id uint64, message []byte, callback func()) {
	if session, ok := s.sessions[user_id]; ok {
		session.Notify(&Notification{
			Message:  message,
			Callback: callback,
		})
	}
}

/* Returns an event whose participant list has been filtered to protect
   privacy of users that will not assist the event and are not friends
	 of the user which the event is gonna be sent*/
func filterEventParticipants(user *UserAccount, dst_event *proto.Event, src_event *proto.Event) {

	if dst_event == nil {
		log.Println("filterEventParticipants() dst_event argument is nil")
		return
	}

	dst_event.Participants = make([]*proto.EventParticipant, 0, 4)

	for _, p := range src_event.Participants {
		// If the participant is a confirmed user (yes or cannot assist answer has been given)
		if p.Response == proto.AttendanceResponse_ASSIST ||
			p.Response == proto.AttendanceResponse_CANNOT_ASSIST ||
			user.IsFriend(p.UserId) ||
			user.id == p.UserId { // self-user
			dst_event.Participants = append(dst_event.Participants, p)
		}
	}
}

func main() {
	//fmt.Println("GOMAXPROCS is", runtime.GOMAXPROCS(0))
	server := NewServer() // Server is global
	initFakeUsers(server)
	server.RegisterCallback(proto.M_PING, onPing)
	server.RegisterCallback(proto.M_CREATE_EVENT, onCreateEvent)
	server.RegisterCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
	server.RegisterCallback(proto.M_USER_AUTH, onUserAuthentication)
	server.RegisterCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
	server.RegisterCallback(proto.M_USER_FRIENDS, onUserFriends)
	server.Run()
}
