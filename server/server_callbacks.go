package main

import (
	core "areyouin/common"
	proto "areyouin/protocol"
	"github.com/twinj/uuid"
	"log"
)

func onCreateAccount(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateUserAccount)
	log.Println("USER CREATE ACCOUNT", msg)

	var reply []byte

	dao := server.NewUserDAO()

	// TODO: Validate user data

	// Create new user account
	user := core.NewUserAccount(server.GetNewID(), msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)

	// Check if user exists. If user e-mail exists maybe orphan due to the way users are
	// inserted into cassandra. So it's needed to check if the user related to this e-mail
	// also exists. In case it doesn't exist, then delete it in order to avoid a collision
	// when inserting later.
	if user_id := dao.GetIDByEmail(user.Email); user_id != 0 {
		if dao.Exists(user_id) {
			reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_USER_EXISTS).Marshal()
			writeReply(reply, session)
			return
		} else {
			if user.HasEmailCredentials() && dao.GetIDByFacebookID(user.Fbid) == user_id {
				dao.DeleteFacebookCredentials(user.Fbid)
			}
			dao.DeleteEmailCredentials(msg.Email)
		}
	}

	// If it's a Facebook account (fbid and fbtoken not empty) then check token
	if user.HasFacebookCredentials() {
		if fbaccount, ok := checkFacebookAccess(user.Fbid, user.Fbtoken); ok {
			// Trust on Facebook e-mail verification
			if user.Email == fbaccount.email {
				user.EmailVerified = true
			}
		} else {
			reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_INVALID_TOKEN).Marshal()
			writeReply(reply, session)
			return
		}
	}

	// Insert into users database
	if ok, _ := dao.Insert(user); ok {
		reply = proto.NewMessage().UserAccessGranted(user.Id, user.AuthToken).Marshal()
	} else { // Facebook account may already be linked to another user
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_USER_EXISTS).Marshal()
	}

	writeReply(reply, session)
}

func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.NewAuthToken)
	log.Println("USER NEW AUTH TOKEN", msg)

	dao := server.NewUserDAO()

	var reply []byte

	// Get new token by e-mail and password
	if msg.Type == proto.AuthType_A_NATIVE {
		if user_id := dao.CheckEmailCredentials(msg.Pass1, msg.Pass2); user_id != 0 {
			new_auth_token := uuid.NewV4()
			if err := dao.SetAuthToken(user_id, new_auth_token); err != nil {
				log.Println("onUserNewAuthToken:", err)
			}
			reply = proto.NewMessage().UserAccessGranted(user_id, new_auth_token).Marshal()
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
		} else if user_id := dao.GetIDByFacebookID(msg.Pass1); user_id != 0 {
			new_auth_token := uuid.NewV4()
			dao.SetAuthTokenAndFBToken(user_id, new_auth_token, msg.Pass1, msg.Pass2)
			reply = proto.NewMessage().UserAccessGranted(user_id, new_auth_token).Marshal()
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

	dao := server.NewUserDAO()

	user_id := msg.UserId
	auth_token, _ := uuid.Parse(msg.AuthToken)

	if dao.CheckAuthToken(user_id, auth_token) {
		writeReply(proto.NewMessage().Ok(proto.OK_AUTH).Marshal(), session)
		session.IsAuth = true
		session.UserId = user_id
		server.RegisterSession(session)
		log.Println("AUTH OK")
		sendUserFriends(session)
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

func onCreateEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateEvent)
	log.Println("CREATE EVENT", msg)

	if !session.IsAuth {
		log.Println("Received CREATE EVENT message from unauthenticated session", session)
		return
	}

	dao := server.NewUserDAO()

	author := dao.Load(session.UserId)
	if author == nil {
		log.Println("Author should exist but it seems it didn't on an error ocurred")
		writeReply(proto.NewMessage().Error(proto.M_CREATE_EVENT, proto.E_EVENT_CREATION_ERROR).Marshal(), session)
		return
	}

	// TODO: Validate input data
	// TODO: Check overlapping with other own published events
	event := core.CreateNewEvent(server.GetNewID(), author.Id, author.Name, msg.StartDate, msg.EndDate, msg.Message)

	// Prepare participants
	participantsList := server.createParticipantsList(author.Id, msg.Participants)

	// Add author as another participant of the event and assume he or she
	// will assist by default
	participant := author.AsParticipant()
	participant.SetFields(proto.AttendanceResponse_ASSIST, proto.MessageStatus_NO_DELIVERED)
	participantsList = append(participantsList, participant)

	// Only proceed if there are more participants than the only author
	if len(participantsList) > 1 {

		if ok := server.PublishEvent(event, participantsList); ok {
			writeReply(proto.NewMessage().Ok(proto.OK_ACK).Marshal(), session)
			log.Println("EVENT STORED BUT NOT PUBLISHED", event.EventId)
		} else {
			writeReply(proto.NewMessage().Error(proto.M_CREATE_EVENT, proto.E_EVENT_CREATION_ERROR).Marshal(), session)
			log.Println("EVENT CREATION ERROR")
		}

	} else {
		writeReply(proto.NewMessage().Error(proto.M_CREATE_EVENT, proto.E_EVENT_CREATION_ERROR).Marshal(), session)
		log.Println("EVENT CREATION ERROR INVALID PARTICIPANTS")
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

func onUserFriends(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	log.Println("USER FRIENDS") // Message does not has payload

	if !session.IsAuth {
		log.Println("Received USER FRIENDS message from unauthenticated client", session)
		return
	}

	/*server := session.Server
	var reply []byte

	if !server.udb.ExistID(session.UserId) {
		reply = proto.NewMessage().Error(proto.M_USER_FRIENDS, proto.E_MALFORMED_MESSAGE).Marshal()
		writeReply(reply, session)
		log.Println("FIXME: Received USER FRIENDS message from authenticated user but non-existent")
	} else if ok := sendUserFriends(session); !ok {
		reply = proto.NewMessage().Error(proto.M_USER_FRIENDS, proto.E_INVALID_USER).Marshal()
		writeReply(reply, session)
	}*/
}
