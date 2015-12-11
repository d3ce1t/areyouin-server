package main

import (
	proto "areyouin/protocol"
	"github.com/twinj/uuid"
	"log"
	"time"
)

func onCreateEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateEvent)
	log.Println("CREATE EVENT", msg)

	if !session.IsAuth {
		log.Println("Received CREATE EVENT message from unauthenticated session", session)
		return
	}

	author, _ := server.udb.GetByID(session.UserId)

	// TODO: Validate input data
	// TODO: Check overlapping with other own published events
	event := &proto.Event{
		EventId:            server.GetNewUserID(), // Maybe a bottleneck here
		AuthorId:           author.id,
		AuthorName:         author.name,
		CreationDate:       time.Now().UTC().Unix(), // Seconds
		StartDate:          msg.StartDate,
		EndDate:            msg.EndDate,
		Message:            msg.Message,
		IsPublic:           false,
		NumberParticipants: 1, // The own author
	}

	// Put participants info into the event
	allright := true

	for _, user_id := range msg.Participants {

		uac, ok := server.udb.GetByID(user_id)

		if !ok {
			log.Println("Trying to add into event", event.EventId, "a participant that does not exist")
			allright = false
			break
		}

		participant := &proto.EventParticipant{
			UserId:    uac.id,
			Name:      uac.name,
			Response:  proto.AttendanceResponse_NO_RESPONSE,
			Delivered: proto.MessageStatus_NO_DELIVERED,
		}

		event.Participants = append(event.Participants, participant)
	}

	if allright {
		// Add author as another participant of the event and assume he or she
		// will assist by default
		event.Participants = append(event.Participants, &proto.EventParticipant{
			UserId:    author.id,
			Name:      author.name,
			Response:  proto.AttendanceResponse_ASSIST,
			Delivered: proto.MessageStatus_NO_DELIVERED,
		})

		event.NumberParticipants = uint32(len(event.Participants))

		if ok := server.edb.Insert(event); ok { // Insert is not thread-safe
			server.ds.Submit(event)
			writeReply(proto.NewMessage().Ok(proto.OK_ACK).Marshal(), session)
			log.Println("EVENT STORED BUT NOT PUBLISHED", event.EventId)
		} else {
			writeReply(proto.NewMessage().Error(proto.M_CREATE_EVENT, proto.E_EVENT_CREATION_ERROR).Marshal(), session)
			log.Println("EVENT CREATION ERROR")
		}

	} else {
		writeReply(proto.NewMessage().Error(proto.M_CREATE_EVENT, proto.E_EVENT_CREATION_ERROR).Marshal(), session)
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

func onCreateAccount(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateUserAccount)
	log.Println("USER CREATE ACCOUNT", msg)

	var reply []byte

	// User exists
	if server.udb.ExistEmail(msg.Email) {
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_USER_EXISTS).Marshal()
		writeReply(reply, session)
		return
	}

	// TODO: Validate user date

	// Create new user account
	user := NewUserAccount(server.GetNewUserID(), msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)

	// If it's a Facebook account (fbid and fbtoken not empty) then check token
	if user.IsFacebook() {
		if fbaccount, ok := checkFacebookAccess(user.fbid, user.fbtoken); ok {
			// Trust on Facebook e-mail verification
			if user.email == fbaccount.email {
				user.email_verified = true
			}
		} else {
			reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_INVALID_TOKEN).Marshal()
			writeReply(reply, session)
			return
		}
	}

	// Insert into users database
	if server.udb.Insert(user) {
		reply = proto.NewMessage().UserAccessGranted(user.id, user.auth_token).Marshal()
	} else { // Facebook account may already be linked to another user
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_USER_EXISTS).Marshal()
	}

	writeReply(reply, session)
}

func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.NewAuthToken)
	log.Println("USER NEW AUTH TOKEN", msg)

	var reply []byte

	// Get new token by e-mail and password
	if msg.Type == proto.AuthType_A_NATIVE {
		if userAccount, ok := server.udb.GetByEmail(msg.Pass1); ok && msg.Pass2 == userAccount.password {
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
		} else if userAccount, ok := server.udb.GetByFBUID(msg.Pass1); ok {
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

	writeReply(reply, session)
}

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.UserAuthentication)
	log.Println("USER AUTH", msg)

	user_id := msg.UserId
	auth_token, _ := uuid.Parse(msg.AuthToken)

	if server.udb.CheckAccess(user_id, auth_token) {
		writeReply(proto.NewMessage().Ok(proto.OK_AUTH).Marshal(), session)
		session.IsAuth = true
		session.UserId = user_id
		server.RegisterSession(session)
		log.Println("AUTH OK")
		server.sendUserFriends(session)
		// FIXME: Do not send all of the private events, but limit to a fixed number
		sendPrivateEvents(session)
	} else {
		writeReply(proto.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER).Marshal(), session)
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

func sendPrivateEvents(session *AyiSession) bool {

	server := session.Server
	result := false

	if uac, ok := server.udb.GetByID(session.UserId); ok {
		events := uac.GetAllEvents()
		reply := proto.NewMessage().EventsList(events).Marshal()
		log.Println("SEND PRIVATE EVENTS to", session)
		writeReply(reply, session)
		result = true
	} else {
		log.Println("SendPrivateEvents failed because of an invalid UserID")
		result = false
	}

	return result
}

func onListPublicEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onHistoryAuthoredEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onHistoryPrivateEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onHistoryPublicEvents(packet_type proto.PacketType, message proto.Message, client *AyiSession) {

}

func onUserFriends(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	log.Println("USER FRIENDS") // Message does not has payload

	if !session.IsAuth {
		log.Println("Received USER FRIENDS message from unauthenticated client", session)
		return
	}

	server := session.Server
	var reply []byte

	if !server.udb.ExistID(session.UserId) {
		reply = proto.NewMessage().Error(proto.M_USER_FRIENDS, proto.E_MALFORMED_MESSAGE).Marshal()
		writeReply(reply, session)
		log.Println("FIXME: Received USER FRIENDS message from authenticated user but non-existent")
	} else if ok := server.sendUserFriends(session); !ok {
		reply = proto.NewMessage().Error(proto.M_USER_FRIENDS, proto.E_INVALID_USER).Marshal()
		writeReply(reply, session)
	}
}
