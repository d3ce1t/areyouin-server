package main

import (
	proto "areyouin/protocol"
	"fmt"
	fb "github.com/huandu/facebook"
	"github.com/twinj/uuid"
	"log"
	"net"
	"runtime"
	"time"
)

const MAX_WRITE_TIMEOUT = 10 * time.Second

var users_inbox = make(map[uint64]*Inbox)
var sessions = make(map[uint64]*AyiSession)
var udb = NewUserDatabase()
var edb = NewEventDatabase()
var uid_ch = make(chan uint64)
var ds = NewDeliverySystem()
var callbacks = make(map[proto.PacketType]Callback)

type Callback func(proto.PacketType, proto.Message, *AyiSession)

func onCreateEvent(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

	msg := message.(*proto.CreateEvent)
	log.Println("CREATE EVENT", msg)

	if !client.IsAuth {
		log.Println("Received CREATE EVENT message from unauthenticated client", client)
		return
	}

	uac, _ := udb.GetByID(client.UserId)

	// TODO: Validate input data
	// TODO: Check overlapping with other own published events
	event := &proto.Event{
		EventId:            getNewUserID(), // Maybe a bottleneck here
		AuthorId:           client.UserId,
		AuthorName:         uac.name,
		CreationDate:       time.Now().UTC().Unix(),
		StartDate:          msg.StartDate,
		EndDate:            msg.EndDate,
		Message:            msg.Message,
		IsPublic:           false,
		NumberParticipants: 1, // The own author
	}

	if udb.ExistAllIDs(msg.Participants) {
		if ok := edb.Insert(event); ok {
			ds.Submit(event, msg.Participants)
			writeReply(proto.NewMessage().Ok(proto.OK_ACK).Marshal(), client)
			log.Println("EVENT STORED BUT NOT PUBLISHED", event.EventId)
		} else {
			writeReply(proto.NewMessage().Error(proto.M_CREATE_EVENT, proto.E_EVENT_CREATION_ERROR).Marshal(), client)
			log.Println("EVENT CREATION ERROR")
		}
	} else {
		writeReply(proto.NewMessage().Error(proto.M_CREATE_EVENT, proto.E_EVENT_CREATION_ERROR).Marshal(), client)
		log.Println("INVALID PARTICIPANTS")
	}
}

func onCancelEvent(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onInviteUsers(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onCancelUsersInvitation(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onConfirmAttendance(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onModifyEvent(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onVoteChange(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onUserPosition(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onUserPositionRange(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onCreateAccount(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

	msg := message.(*proto.CreateUserAccount)
	log.Println("USER CREATE ACCOUNT", msg)

	var reply []byte

	// User exists
	if udb.ExistEmail(msg.Email) {
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_USER_EXISTS).Marshal()
		writeReply(reply, client)
		return
	}

	// TODO: Validate user date

	// Create new user account
	user := NewUserAccount(msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)

	// If it's a Facebook account (fbid and fbtoken not empty) then check token
	if user.IsFacebook() {
		if fbaccount, ok := checkFacebookAccess(user.fbid, user.fbtoken); ok {
			// Trust on Facebook e-mail verification
			if user.email == fbaccount.email {
				user.email_verified = true
			}
		} else {
			reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_INVALID_TOKEN).Marshal()
			writeReply(reply, client)
			return
		}
	}

	// Insert into users database
	if udb.Insert(user) {
		reply = proto.NewMessage().UserAccessGranted(user.id, user.auth_token).Marshal()
	} else { // Facebook account may already be linked to another user
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_USER_EXISTS).Marshal()
	}

	writeReply(reply, client)
}

func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, client *AyiSession) {
	msg := message.(*proto.NewAuthToken)
	log.Println("USER NEW AUTH TOKEN", msg)

	var reply []byte

	// Get new token by e-mail and password
	if msg.Type == proto.AuthType_A_NATIVE {
		if userAccount, ok := udb.GetByEmail(msg.Pass1); ok && msg.Pass2 == userAccount.password {
			userAccount.auth_token = uuid.NewV4()
			reply = proto.NewMessage().UserAccessGranted(userAccount.id, userAccount.auth_token).Marshal()
			log.Println("ACCESS GRANTED")
		} else {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER).Marshal()
			log.Println("INVALID USER")
		}
		// Get new token by Facebook User ID and Facebook Access Token
	} else if msg.Type == proto.AuthType_A_FACEBOOK {

		_, valid_token := checkFacebookAccess(msg.Pass1, msg.Pass2)

		if !valid_token {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_FB_INVALID_TOKEN).Marshal()
			log.Println("INVALID TOKEN")
		} else if userAccount, ok := udb.GetByFBUID(msg.Pass1); ok {
			userAccount.fbtoken = msg.Pass2
			userAccount.auth_token = uuid.NewV4()
			reply = proto.NewMessage().UserAccessGranted(userAccount.id, userAccount.auth_token).Marshal()
			log.Println("ACCESS GRANTED")
		} else {
			// User do not exist
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER).Marshal()
			log.Println("INVALID USER")
		}
	} else {
		log.Println("USER NEW AUTH TOKEN malformed message")
		reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_MALFORMED_MESSAGE).Marshal()
	}

	writeReply(reply, client)
}

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

	msg := message.(*proto.UserAuthentication)
	log.Println("USER AUTH", msg)

	user_id := msg.UserId
	auth_token, _ := uuid.Parse(msg.AuthToken)

	var reply []byte

	if udb.CheckAccess(user_id, auth_token) {
		msg := proto.NewMessage().Ok(proto.OK_AUTH)
		reply = msg.Marshal()
		writeReply(reply, client)
		client.IsAuth = true
		client.UserId = user_id
		sessions[user_id] = client
		log.Println("AUTH OK")
		sendUserFriends(client)
		// TODO: Send list of current events
	} else {
		reply = proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER).Marshal()
		writeReply(reply, client)
		log.Println("INVALID USER")
	}
}

func onPing(packet_type proto.PacketType, message proto.Message, client *AyiSession) {
	msg := message.(*proto.Ping)
	log.Println("PING", msg.CurrentTime, client)
	reply := proto.NewMessage().Pong().Marshal()
	writeReply(reply, client)
}

func onReadEvent(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onListAuthoredEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onListPrivateEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onListPublicEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onHistoryAuthoredEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onHistoryPrivateEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onHistoryPublicEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onUserFriends(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

	log.Println("USER FRIENDS") // Message does not has payload

	if !client.IsAuth {
		log.Println("Received USER FRIENDS message from unauthenticated client", client)
		return
	}

	var reply []byte

	if !udb.ExistID(client.UserId) {
		reply = proto.NewMessage().Error(proto.M_USER_FRIENDS, proto.E_MALFORMED_MESSAGE).Marshal()
		writeReply(reply, client)
		log.Println("FIXME: Received USER FRIENDS message from authenticated user but non-existent")
	} else if ok := sendUserFriends(client); !ok {
		reply = proto.NewMessage().Error(proto.M_USER_FRIENDS, proto.E_INVALID_USER).Marshal()
		writeReply(reply, client)
	}
}

func sendUserFriends(client *AyiSession) bool {

	result := false

	if uac, ok := udb.GetByID(client.UserId); ok {
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

// TODO: Close connection if no login for a while (maybe 30 seconds)
// TODO: Close connection if no PING, PONG dialog (each 15 minutes?)
func handleSession(session *AyiSession) {

	log.Println("New connection from", session)

	for {
		// Read messages and then write (if needed)
		// Sync behaviour
		msg := proto.ReadPacket(session.Conn)

		if msg == nil {
			log.Println("Session closed")
			if uac, ok := udb.GetByID(session.UserId); ok {
				uac.last_connection = time.Now().UTC().Unix()
			}
			delete(sessions, session.UserId)
			return
		}

		err := ServeMessage(msg, session) // may block until writes are performed
		if err != nil {
			// Errors may happen
			log.Println("Unexpected error happened while serving message", err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	//client.Close()
}

// TODO: Notify user through the active connection and from its own thread
func notifyUser(user_id uint64, message []byte) {

	/*client, ok := sessions[user_id]

	if !ok {
		return
	}*/

}

func getNewUserID() uint64 {
	return <-uid_ch
}

func GeneratorTask(gid uint16, uid_ch chan uint64) {

	uidgen := CreateGenerator(gid)

	for {
		newId := uidgen()
		log.Println("New ID created!", newId)
		uid_ch <- newId
	}
}

func initDummyUsers() {
	user1 := NewUserAccount("User 1", "user1@foo.com", "12345", "", "", "")
	user2 := NewUserAccount("User 2B", "user2@foo.com", "12345", "", "", "")
	user3 := NewUserAccount("User 3A", "user3@foo.com", "12345", "", "", "")

	user1.id = 10745351749240831
	user1.auth_token, _ = uuid.Parse("119376ac-c58e-4704-850a-66a6f9663eaa")
	user2.id = getNewUserID()
	user3.id = getNewUserID()

	udb.Insert(user1)
	udb.Insert(user2)
	udb.Insert(user3)

	user1.AddFriend(user2.id)
	user1.AddFriend(user3.id)

	user2.AddFriend(user1.id)
	user2.AddFriend(user3.id)

	user3.AddFriend(user1.id)
	user3.AddFriend(user2.id)
}

func RegisterCallback(command proto.PacketType, f Callback) {
	if callbacks == nil {
		callbacks = make(map[proto.PacketType]Callback)
	}
	callbacks[command] = f
}

func ServeMessage(packet *proto.AyiPacket, session *AyiSession) error {

	message := packet.DecodeMessage()

	if message == nil {
		log.Fatal("Unknown message", packet)
		return nil
	}

	if f, ok := callbacks[packet.Type()]; ok {
		f(packet.Type(), message, session)
	}

	return nil
}

func main() {

	fmt.Println("GOMAXPROCS is", runtime.GOMAXPROCS(0))

	// Start up server listener
	listener, err := net.Listen("tcp", ":1822")

	if err != nil {
		panic("Couldn't start listening: " + err.Error())
	}

	RegisterCallback(proto.M_PING, onPing)
	RegisterCallback(proto.M_CREATE_EVENT, onCreateEvent)
	RegisterCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
	RegisterCallback(proto.M_USER_AUTH, onUserAuthentication)
	RegisterCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
	RegisterCallback(proto.M_USER_FRIENDS, onUserFriends)

	// Setup server components
	go GeneratorTask(1, uid_ch)
	initDummyUsers()
	ds.Run()

	for {
		client, err := listener.Accept()

		if err != nil {
			fmt.Println("Couldn't accept:", err.Error())
			continue
		}

		go handleSession(NewSession(client))
	}
}
