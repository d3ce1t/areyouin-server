package main

import (
	"github.com/twinj/uuid"
	"log"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
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
	// a security wait is introduced here to prevent an attacker massive e-mail checking
	time.Sleep(SECURITY_WAIT_TIME)

	dao := server.NewUserDAO()

	// Check if user exists and performs some sanity of data if needed
	if exists, err := dao.ExistWithSanity(user); exists {
		session.WriteReply(proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_EMAIL_EXISTS).Marshal())
		return
	} else if err != nil { // If an error happen, assume it may exist so, cancel operation
		session.WriteReply(proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_OPERATION_FAILED).Marshal())
		return
	}

	var reply []byte

	// If it's a Facebook account (fbid and fbtoken are not empty) check token
	if user.HasFacebookCredentials() {

		fbsession := fb.NewSession(user.Fbtoken)

		if _, err := fb.CheckAccess(user.Fbid, fbsession); err != nil {
			reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_INVALID_ACCESS).Marshal()
			session.WriteReply(reply)
			fb.LogError(err)
			return
		}
	}

	// Insert into users database
	if err := dao.Insert(user); err != nil {
		// Facebook account may already be linked to another user
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_EXISTS).Marshal()
		session.WriteReply(reply)
		log.Println("onCreateUserAccount Error:", err)
		return
	}

	reply = proto.NewMessage().UserAccessGranted(user.Id, user.AuthToken).Marshal()

	// Import Facebook friends that uses AreYouIN if needed
	if user.HasFacebookCredentials() {
		task := &ImportFacebookFriends{
			UserId: user.Id,
			Name:   user.Name,
			//Fbid:    user.Fbid,
			Fbtoken: user.Fbtoken,
		}
		server.task_executor.Submit(task)
	}

	session.WriteReply(reply)
}

// FIXME: Renew token should also authenticate the user without needing to get the user to call
// authenticate.
func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkUnauthenticated(session)

	server := session.Server
	msg := message.(*proto.NewAuthToken)
	log.Println("> USER NEW AUTH TOKEN", msg)

	userDAO := server.NewUserDAO()

	var reply []byte

	if msg.Type != proto.AuthType_A_NATIVE && msg.Type != proto.AuthType_A_FACEBOOK {
		reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_MALFORMED_MESSAGE).Marshal()
		session.WriteReply(reply)
		log.Println("< USER NEW AUTH TOKEN malformed message")
		return
	}

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

		session.WriteReply(reply)

		// Get new token by Facebook User ID and Facebook Access Token
	} else if msg.Type == proto.AuthType_A_FACEBOOK {

		fbsession := fb.NewSession(msg.Pass2)

		if _, err := fb.CheckAccess(msg.Pass1, fbsession); err != nil {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_FB_INVALID_ACCESS).Marshal()
			session.WriteReply(reply)
			log.Println("< INVALID FBID OR ACCESS TOKEN")
			fb.LogError(err)
			return
		}

		user_id, err := userDAO.GetIDByFacebookID(msg.Pass1)

		if err != nil {
			if err == dao.ErrNotFound {
				reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD).Marshal()
				log.Println("< INVALID USER OR PASSWORD")
			} else {
				reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
				log.Println("onUserNewAuthToken:", err)
			}
			session.WriteReply(reply)
			return
		}

		new_auth_token := uuid.NewV4()

		if err := userDAO.SetAuthTokenAndFBToken(user_id, new_auth_token, msg.Pass1, msg.Pass2); err != nil {
			reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
			log.Println("onUserNewAuthToken:", err)
			session.WriteReply(reply)
			return
		}

		reply = proto.NewMessage().UserAccessGranted(user_id, new_auth_token).Marshal()
		session.WriteReply(reply)
		log.Println("< ACCESS GRANTED")
	}

}

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkUnauthenticated(session)

	server := session.Server
	msg := message.(*proto.UserAuthentication)
	log.Println("> USER AUTH", msg)

	userDAO := server.NewUserDAO()

	user_id := msg.UserId
	auth_token, _ := uuid.Parse(msg.AuthToken)

	ok, err := userDAO.CheckAuthToken(user_id, auth_token)

	if err != nil {
		if err == dao.ErrNotFound {
			sendAuthError(session)
		} else {
			session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal())
			log.Println("onUserAuthentication Failed:", err)
		}
		return
	}

	if ok {
		session.IsAuth = true
		session.UserId = user_id
		server.RegisterSession(session)
		session.updateLastConnection()
		session.WriteReply(proto.NewMessage().Ok(packet_type).Marshal())
		log.Println("< AUTH OK")
		server.task_executor.Submit(&SendUserFriends{UserId: user_id})
		// FIXME: Do not send all of the private events, but limit to a fixed number
		sendPrivateEvents(session)
	} else {
		sendAuthError(session)
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
		log.Println("onCreateEvent Failed", err)
		return
	}

	// Check event creation is inside creation window
	currentDate := core.UnixMillisToTime(core.GetCurrentTimeSeconds())
	createdDate := core.UnixMillisToTime(msg.CreatedDate)

	if createdDate.Before(currentDate.Add(-time.Minute)) || createdDate.After(currentDate.Add(time.Minute)) {
		session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_EVENT_OUT_OF_CREATE_WINDOW).Marshal())
		log.Println("onCreateEvent Failed: Event is out of the allowed creating window")
		return
	}

	event := core.CreateNewEvent(server.GetNewID(), author.Id, author.Name, msg.CreatedDate, msg.StartDate, msg.EndDate, msg.Message)

	if _, err := event.IsValid(); err != nil {

		switch err {
		case core.ErrInvalidStartDate:
			session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_EVENT_INVALID_START_DATE).Marshal())
		case core.ErrInvalidEndDate:
			session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_EVENT_INVALID_END_DATE).Marshal())
		default:
			session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_INVALID_INPUT).Marshal())
		}

		log.Println("onCreateEvent Error:", err)
		return
	}

	// Prepare participants
	participantsList, warning := server.createParticipantsList(author.Id, msg.Participants)

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
		if warning == ErrNonFriendsIgnored || warning == ErrUnregisteredFriendsIgnored {
			session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_INVALID_PARTICIPANT).Marshal())
		} else {
			session.WriteReply(proto.NewMessage().Error(packet_type, proto.E_EVENT_PARTICIPANTS_REQUIRED).Marshal())
		}
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
		if err == dao.ErrNotFound {
			reply = proto.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_INVALID_EVENT_OR_PARTICIPANT).Marshal()
			log.Println("< CONFIRM ATTENDANCE INVALID_EVENT_OR_PARTICIPANT")
		} else {
			reply = proto.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_OPERATION_FAILED).Marshal()
			log.Println("< CONFIRM ATTENDANCE OPERATION FAILED")
		}
		session.WriteReply(reply)
		log.Println("ConfirmAttendance:", err)
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

	if participants, err := event_dao.LoadAllParticipants(task.EventId); err == nil {
		log.Println("Num.Participants:", len(participants))
		task.AddParticipantsDst(participants)
		server.task_executor.Submit(task)
	} else {
		log.Println("onConfirmAttendance:", err)
	}
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

func onPong(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
	checkAuthenticated(session)
	msg := message.(*proto.Pong)
	log.Println("> PONG", msg.CurrentTime, session)
}
