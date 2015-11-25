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

var udb = newUserDatabase()

func initDummyUsers() {
	user1 := newUserAccount("User 1", "user1@example.com", "12345", "", "", "")
	user2 := newUserAccount("User 2", "user2@example.com", "12345", "", "", "")
	user3 := newUserAccount("User 3", "user3@example.com", "12345", "", "", "")

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

func onCreateEvent(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {
	log.Println("CREATE EVENT", message)
}

func onCancelEvent(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onInviteUsers(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onCancelUsersInvitation(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onConfirmAttendance(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onModifyEvent(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onVoteChange(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onUserPosition(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onUserPositionRange(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onCreateAccount(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

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
	user := newUserAccount(msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)

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

func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {
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

	time.Sleep(2000 * time.Millisecond) // FIXME: Remove this
	writeReply(reply, client)
}

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

	msg := message.(*proto.UserAuthentication)
	log.Println("USER AUTH", msg)

	user_id, _ := uuid.Parse(msg.UserId)
	auth_token, _ := uuid.Parse(msg.AuthToken)

	var reply []byte

	if udb.CheckAccess(user_id, auth_token) {
		reply = proto.NewMessage().Ok(proto.OK_AUTH).Marshal()
		client.SetAuthenticated(true)
		client.SetUserId(user_id.String())
		log.Println("AUTH OK")
	} else {
		reply = proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER).Marshal()
		log.Println("INVALID USER")
	}

	writeReply(reply, client)

	// TODO: Send list of friends
	// TODO: Send list of current events
}

func onPing(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {
	msg := message.(*proto.Ping)
	log.Println("PING", msg.CurrentTime, client)
	reply := proto.NewMessage().Pong().Marshal()
	writeReply(reply, client)
}

func onReadEvent(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onListAuthoredEvents(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onListPrivateEvents(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onListPublicEvents(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onHistoryAuthoredEvents(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onHistoryPrivateEvents(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onHistoryPublicEvents(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {

}

func onUserFriends(packet_type proto.PacketType, message proto.Message, client *proto.AyiClient) {
	log.Println("USER FRIENDS")
	//reply := proto.NewMessage().Pong().Marshal()
	//writeReply(reply, client)
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

func writeReply(reply []byte, client net.Conn) {
	client.SetWriteDeadline(time.Now().Add(MAX_WRITE_TIMEOUT))
	_, err := client.Write(reply)
	if err != nil {
		log.Println("Coudn't send reply: ", err)
	}
}

func handleConnection(server *proto.AyiListener, client *proto.AyiClient) {

	log.Println("New connection from", client)

	for {
		// Read messages and then write (if needed)
		// Sync behaviour
		msg := proto.ReadPacket(client)

		if msg == nil {
			log.Println("Session closed")
			return
		}

		err := server.ServeMessage(msg, client) // may block until writes are performed
		if err != nil {
			// Errors may happen
		}

		time.Sleep(100 * time.Millisecond)
	}

	//client.Close()
}

func main() {

	fmt.Println("GOMAXPROCS is", runtime.GOMAXPROCS(0))

	server, err := proto.Listen("tcp", ":1822")

	if err != nil {
		panic("Couldn't start listening: " + err.Error())
	}

	server.RegisterCallback(proto.M_PING, onPing)
	server.RegisterCallback(proto.M_CREATE_EVENT, onCreateEvent)
	server.RegisterCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
	server.RegisterCallback(proto.M_USER_AUTH, onUserAuthentication)
	server.RegisterCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
	server.RegisterCallback(proto.M_USER_FRIENDS, onUserFriends)

	initDummyUsers()

	for {
		client, err := server.Accept()

		if err != nil {
			fmt.Println("Couldn't accept:", err.Error())
			continue
		}

		go handleConnection(server, client)
	}
}
