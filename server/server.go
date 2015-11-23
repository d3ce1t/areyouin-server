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

func onCreateEvent(packet_type proto.PacketType, message proto.Message, client net.Conn) {
	log.Println("CREATE EVENT", message)
}

func onCancelEvent(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onInviteUsers(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onCancelUsersInvitation(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onConfirmAttendance(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onModifyEvent(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onVoteChange(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onUserPosition(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onUserPositionRange(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onCreateAccount(packet_type proto.PacketType, message proto.Message, client net.Conn) {

	msg := message.(*proto.CreateUserAccount)
	log.Println("USER CREATE ACCOUNT", msg)

	var reply []byte

	// User exists
	if udb.Exist(msg.Email) {
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

func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, client net.Conn) {
	msg := message.(*proto.NewAuthToken)
	log.Println("USER NEW AUTH TOKEN", msg)

	var reply []byte

	// Get new token by e-mail and password
	/*if msg.Type == proto.AuthType_A_NATIVE {
		if userAccount, ok := udb.GetByEmail(msg.Pass1); ok && msg.Pass2 == userAccount.password {
			userAccount.auth_token = uuid.NewV4()
			reply = proto.NewMessage().UserAccessGranted(userAccount.id, userAccount.auth_token).Marshal()
		} else {
			reply = proto.NewMessage().Error(proto.E_INVALID_USER).Marshal()
		}
		// Get new token by Facebook User ID and Facebook Access Token
	} else*/
	if msg.Type == proto.AuthType_A_FACEBOOK {

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

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, client net.Conn) {

	msg := message.(*proto.UserAuthentication)
	log.Println("USER AUTH", msg)

	user_id, _ := uuid.Parse(msg.UserId)
	auth_token, _ := uuid.Parse(msg.AuthToken)

	var reply []byte

	if udb.CheckAccess(user_id, auth_token) {
		reply = proto.NewMessage().Ok(proto.OK_AUTH).Marshal()
		log.Println("AUTH OK")
	} else {
		reply = proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER).Marshal()
		log.Println("INVALID USER")
	}

	writeReply(reply, client)
}

func onPing(packet_type proto.PacketType, message proto.Message, client net.Conn) {
	msg := message.(*proto.Ping)
	log.Println("PING", msg.CurrentTime, client)
	reply := proto.NewMessage().Pong().Marshal()
	writeReply(reply, client)
}

func onReadEvent(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onListAuthoredEvents(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onListPrivateEvents(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onListPublicEvents(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onHistoryAuthoredEvents(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onHistoryPrivateEvents(packet_type proto.PacketType, message proto.Message, client net.Conn) {

}

func onHistoryPublicEvents(packet_type proto.PacketType, message proto.Message, client net.Conn) {

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

func handleConnection(id int, server *proto.AyiListener, client net.Conn) {

	log.Println("New connection", id, client)

	for {
		// Read messages and then write (if needed)
		// Sync behaviour
		msg := proto.ReadPacket(client)

		if msg == nil {
			log.Println("Session closed")
			return
		}

		//log.Println("Type:", msg.Header.Type, "Size:", msg.Header.Size)
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

	id := 0

	for {
		client, err := server.Accept()

		if err != nil {
			fmt.Println("Couldn't accept:", err.Error())
			continue
		}

		go handleConnection(id, server, client)
		id++
	}
}
