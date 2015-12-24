package main

import (
	"github.com/twinj/uuid"
	"log"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	proto "peeple/areyouin/protocol"
	"time"
)

const (
	SECURITY_WAIT_TIME = 1 * time.Second
)

func onCreateAccount(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkUnauthenticated(session)

	server := session.Server
	msg := message.(*proto.CreateUserAccount)
	log.Println("> USER CREATE ACCOUNT", msg)

	// Create new user account
	user := core.NewUserAccount(server.GetNewID(), msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)

	// Check if its a valid user, so the input was correct
	if !user.IsValid() {
		session.WriteReply(proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_INVALID_INPUT).Marshal())
		return
	}

	// Because an attacker could use the create account feature in order to check if a user exists,
	// a security wait is introduced here to prevent attackers of massive e-mails checking
	time.Sleep(SECURITY_WAIT_TIME)

	dao := server.NewUserDAO()

	// Check if user exists. If user e-mail exists may be orphan due to the way users are
	// inserted into cassandra. So it's needed to check if the user related to this e-mail
	// also exists. In case it doesn't exist, then delete it in order to avoid a collision
	// when inserting later.
	if exists, _ := dao.ExistWithSanity(user); !exists {
		session.WriteReply(proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_EMAIL_EXISTS).Marshal())
		return
	}

	var reply []byte

	// If it's a Facebook account (fbid and fbtoken not empty) then check token
	if user.HasFacebookCredentials() {
		if fbaccount, ok := checkFacebookAccess(user.Fbid, user.Fbtoken); ok {
			// Trust on Facebook e-mail verification
			if user.Email == fbaccount.email {
				user.EmailVerified = true
			}
		} else {
			reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_INVALID_ACCESS).Marshal()
			session.WriteReply(reply)
			return
		}
	}

	// Insert into users database
	if ok, _ := dao.Insert(user); ok {
		reply = proto.NewMessage().UserAccessGranted(user.Id, user.AuthToken).Marshal()
	} else { // Facebook account may already be linked to another user
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_EXISTS).Marshal()
	}

	session.WriteReply(reply)
}

func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkUnauthenticated(session)

	server := session.Server
	msg := message.(*proto.NewAuthToken)
	log.Println("> USER NEW AUTH TOKEN", msg)

	userDAO := server.NewUserDAO()

	var reply []byte

	// Get new token by e-mail and password
	if msg.Type == proto.AuthType_A_NATIVE {

		if user_id, err := userDAO.CheckEmailCredentials(msg.Pass1, msg.Pass2); err == nil {
			new_auth_token := uuid.NewV4()
			if err := userDAO.SetAuthToken(user_id, new_auth_token); err == nil {
				reply = proto.NewMessage().UserAccessGranted(user_id, new_auth_token).Marshal()
				log.Println("< ACCESS GRANTED")
			} else {
				reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
				log.Println("onUserNewAuthToken:", err)
			}
		} else if err == dao.ErrNotFound {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD).Marshal()
			log.Println("< INVALID USER OR PASSWORD")
		} else {
			reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
			log.Println("onUserNewAuthToken:", err)
		}

		// Get new token by Facebook User ID and Facebook Access Token
	} else if msg.Type == proto.AuthType_A_FACEBOOK {

		_, valid_token := checkFacebookAccess(msg.Pass1, msg.Pass2)

		if !valid_token {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_FB_INVALID_ACCESS).Marshal()
			log.Println("< INVALID FBID OR ACCESS TOKEN")
		} else if user_id, err := userDAO.GetIDByFacebookID(msg.Pass1); err == nil {
			new_auth_token := uuid.NewV4()
			if err := userDAO.SetAuthTokenAndFBToken(user_id, new_auth_token, msg.Pass1, msg.Pass2); err == nil {
				reply = proto.NewMessage().UserAccessGranted(user_id, new_auth_token).Marshal()
				log.Println("< ACCESS GRANTED")
			} else {
				reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
				log.Println("onUserNewAuthToken:", err)
			}
		} else if err == dao.ErrNotFound {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD).Marshal()
			log.Println("< INVALID USER OR PASSWORD")
		} else {
			reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
			log.Println("onUserNewAuthToken:", err)
		}
	} else {
		log.Println("< USER NEW AUTH TOKEN malformed message")
		reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_MALFORMED_MESSAGE).Marshal()
	}

	session.WriteReply(reply)
}

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkUnauthenticated(session)

	server := session.Server
	msg := message.(*proto.UserAuthentication)
	log.Println("> USER AUTH", msg)

	userDAO := server.NewUserDAO()

	user_id := msg.UserId
	auth_token, _ := uuid.Parse(msg.AuthToken)

	if ok, err := userDAO.CheckAuthToken(user_id, auth_token); ok {
		session.IsAuth = true
		session.UserId = user_id
		server.RegisterSession(session)
		session.WriteReply(proto.NewMessage().Ok(packet_type).Marshal())
		log.Println("< AUTH OK")
		sendUserFriends(session)
		// FIXME: Do not send all of the private events, but limit to a fixed number
		sendPrivateEvents(session)
	} else if err == nil || err == dao.ErrNotFound {
		sendAuthError(session)
	} else {
		session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal())
		log.Println("onUserAuthentication:", err)
	}
}

func onCreateEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

	server := session.Server
	msg := message.(*proto.CreateEvent)
	log.Println("> CREATE EVENT", msg)

	userDAO := server.NewUserDAO()

	author, err := userDAO.Load(session.UserId)
	if err != nil {
		session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal())
		log.Println("onCreateEvent", err)
		return
	}

	// TODO: Check overlapping with other own published events
	event := core.CreateNewEvent(server.GetNewID(), author.Id, author.Name, msg.StartDate, msg.EndDate, msg.Message)

	if !event.IsValid() {
		session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_INVALID_INPUT).Marshal())
		return
	}

	// Prepare participants
	participantsList := server.createParticipantsList(author.Id, msg.Participants)

	// Add author as another participant of the event and assume he or she
	// will assist by default
	participant := author.AsParticipant()
	participant.SetFields(core.AttendanceResponse_ASSIST, core.MessageStatus_NO_DELIVERED)
	participantsList = append(participantsList, participant)

	// Only proceed if there are more participants than the only author
	if len(participantsList) > 1 {

		if ok := server.PublishEvent(event, participantsList); ok {
			session.WriteReply(proto.NewMessage().Ok(packet_type).Marshal())
			log.Println("< EVENT STORED BUT NOT PUBLISHED", event.EventId)
		} else {
			session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal())
			log.Println("< EVENT CREATION ERROR")
		}

	} else {
		session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_EVENT_PARTICIPANTS_REQUIRED).Marshal())
		log.Println("< EVENT CREATION ERROR INVALID PARTICIPANTS")
	}
}

func onCancelEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onInviteUsers(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onCancelUsersInvitation(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

// When a ConfirmAttendance message is received, the attendance response of the participant
// in the participant event list is changed and notified to the other participants. It is
// important to note that num_attendees is not changed server-side till the event has started.
// Clients are cool counting attendees :)
func onConfirmAttendance(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

	msg := message.(*proto.ConfirmAttendance)
	log.Println("> CONFIRM ATTENDANCE", msg)

	server := session.Server
	event_dao := server.NewEventDAO()

	// Preconditions: User must have received the invitationm, so user must be in the event participant list
	// or user has the event in his inbox
	participant, err := event_dao.LoadParticipant(msg.EventId, session.UserId)
	var reply []byte

	if err != nil {
		reply = proto.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_INVALID_EVENT_OR_PARTICIPANT).Marshal()
		session.WriteReply(reply)
		log.Println("< ConfirmAttendance:", err)
		return
	}

	// If the stored response is the same as the provided, send OK response inmediately
	if participant.Response == msg.ActionCode {
		reply = proto.NewMessage().Ok(packet_type).Marshal()
		session.WriteReply(reply)
		return
	}

	if err := event_dao.SetParticipantResponse(session.UserId, msg.EventId, msg.ActionCode); err != nil {
		reply = proto.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_OPERATION_FAILED).Marshal()
		session.WriteReply(reply)
		return
	}

	// Send OK Response
	participant.Response = msg.ActionCode
	reply = proto.NewMessage().Ok(packet_type).Marshal()
	session.WriteReply(reply)

	// Notify participants
	task := &NotifyParticipantChange{
		EventId:  msg.EventId,
		UserId:   session.UserId,
		Name:     participant.Name,
		Response: participant.Response,
		Status:   participant.Delivered,
	}

	server.task_executor.Submit(task)
}

func onModifyEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onVoteChange(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onUserPosition(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onUserPositionRange(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onReadEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onListAuthoredEvents(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onListPrivateEvents(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onListPublicEvents(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onHistoryAuthoredEvents(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onHistoryPrivateEvents(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onHistoryPublicEvents(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onUserFriends(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

	log.Println("> USER FRIENDS") // Message does not has payload

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

func onPing(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
	checkAuthenticated(session)
	msg := message.(*proto.Ping)
	log.Println("> PING", msg.CurrentTime, session)
	reply := proto.NewMessage().Pong().Marshal()
	session.WriteReply(reply)
}
