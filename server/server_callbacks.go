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
	log.Printf("> (%v) USER CREATE ACCOUNT %v\n", session, msg)

	// Create new user account
	user := core.NewUserAccount(server.GetNewID(), msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)

	// Check if its a valid user, so the input was correct
	if _, err := user.IsValid(); err != nil {
		error_code := getNetErrorCode(err, proto.E_INVALID_INPUT)
		session.Write(proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, error_code).Marshal())
		log.Printf("< (%v) CREATE ACCOUNT INVALID USER: %v\n", session, err)
		return
	}

	// Because an attacker could use the create account feature in order to check if a user exists,
	// a security wait is introduced here to prevent an attacker massive e-mail checking. It's
	// ridiculous but it's my only anti-flood protection so far
	time.Sleep(SECURITY_WAIT_TIME)

	userDAO := server.NewUserDAO()
	var reply []byte

	// If it's a Facebook account (fbid and fbtoken are not empty) check token
	if user.HasFacebookCredentials() {

		fbsession := fb.NewSession(user.Fbtoken)

		if _, err := fb.CheckAccess(user.Fbid, fbsession); err != nil {
			reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_INVALID_ACCESS).Marshal()
			session.Write(reply)
			log.Printf("< (%v) CREATE ACCOUNT FB ERROR: %v\n", session, fb.GetErrorMessage(err))
			return
		}
	}

	// Insert into users database. Insert will fail if an existing user with the same
	// e-mail address already exists, or if the Facebook address is already being used
	// by another user. It also controls orphaned user_facebook_credentials rows due
	// to the way insertion is performed in Cassandra. When orphaned row is found and
	// grace period has not elapsed, an ErrGracePeriod error is triggered. A different
	// error message could be sent to the client whenever this happens. This way client
	// could be notified to wait grace period seconds and retry. However, an OPERATION
	// FAILED message is sent so far. 	Read UserDAO.insert for more info.
	if err := userDAO.Insert(user); err != nil {
		err_code := getNetErrorCode(err, proto.E_OPERATION_FAILED)
		reply = proto.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, err_code).Marshal()
		session.Write(reply)
		log.Printf("< (%v) CREATE ACCOUNT INSERT ERROR: %v\n", session, err)
		return
	}

	// Insert OK
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

	session.Write(reply)
	log.Printf("< (%v) CREATE ACCOUNT OK\n", session)
}

// FIXME: Renew token should also authenticate the user without needing to get the user to call
// authenticate.
func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkUnauthenticated(session)

	server := session.Server
	msg := message.(*proto.NewAuthToken)
	log.Printf("> (%v) USER NEW AUTH TOKEN %v\n", session, msg)

	var reply []byte

	if msg.Type != proto.AuthType_A_NATIVE && msg.Type != proto.AuthType_A_FACEBOOK {
		reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_MALFORMED_MESSAGE).Marshal()
		session.Write(reply)
		log.Printf("< (%v) USER NEW AUTH TOKEN MALFORMED MESSAGE\n", session)
		return
	}

	userDAO := server.NewUserDAO()

	// Get new token by e-mail and password
	// NOTE: Review
	if msg.Type == proto.AuthType_A_NATIVE {

		if user_id, err := userDAO.CheckEmailCredentials(msg.Pass1, msg.Pass2); err == nil {
			new_auth_token := uuid.NewV4()
			if err := userDAO.SetAuthToken(user_id, new_auth_token); err == nil {
				reply = proto.NewMessage().UserAccessGranted(user_id, new_auth_token).Marshal()
				log.Printf("< (%v) USER NEW AUTH ACCESS GRANTED\n", session)
			} else {
				reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
				log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			}
		} else if err == dao.ErrNotFound {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD).Marshal()
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD\n", session)
		} else {
			reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
		}

		session.Write(reply)

		// Get new token by Facebook User ID and Facebook Access Token
	} else if msg.Type == proto.AuthType_A_FACEBOOK {

		// In this context, E_FB_INVALID_USER_OR_PASSWORD means that account does not exist or
		// it is an invalid account.
		fbsession := fb.NewSession(msg.Pass2)

		if _, err := fb.CheckAccess(msg.Pass1, fbsession); err != nil {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_FB_INVALID_ACCESS).Marshal()
			session.Write(reply)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID FB ACCESS %v\n", session, fb.GetErrorMessage(err))
			return
		}

		user_id, err := userDAO.GetIDByFacebookID(msg.Pass1)

		if err == dao.ErrNotFound {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD).Marshal()
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD", session)
			session.Write(reply)
			return
		} else if err != nil {
			reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			session.Write(reply)
			return
		}

		// Check that account linked to given Facebook ID is valid, i.e. it has user_email_credentials (with or without
		// password). It may happen that row in user_email_credentials exists but does not have set a password. Moreover,
		// it may not have Facebook either and it would still be valid. This behaviour is preferred because if
		// this state is found, something had have to be wrong. Under normal conditions, that state should have never
		// happened. So, at this point only existence of e-mail are checked (credentials are ignored).
		if _, err := userDAO.CheckValidAccount(user_id, false); err != nil {
			reply = proto.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD).Marshal()
			session.Write(reply)
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			return
		}

		new_auth_token := uuid.NewV4()

		if err := userDAO.SetAuthTokenAndFBToken(user_id, new_auth_token, msg.Pass1, msg.Pass2); err != nil {
			reply = proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal()
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			session.Write(reply)
			return
		}

		reply = proto.NewMessage().UserAccessGranted(user_id, new_auth_token).Marshal()
		session.Write(reply)
		log.Printf("< (%v) USER NEW AUTH TOKEN ACCESS GRANTED\n", session)
	}
}

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkUnauthenticated(session)

	server := session.Server
	msg := message.(*proto.UserAuthentication)
	log.Printf("> (%v) USER AUTH %v\n", session, msg)

	userDAO := server.NewUserDAO()

	user_id := msg.UserId
	auth_token, _ := uuid.Parse(msg.AuthToken)

	ok, err := userDAO.CheckAuthToken(user_id, auth_token)

	if err != nil {
		if err == dao.ErrNotFound {
			sendAuthError(session)
		} else {
			session.Write(proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal())
			log.Printf("< (%v) AUTH FAILED %v\n", session, err)
		}
		return
	}

	if ok {
		session.IsAuth = true
		session.UserId = user_id
		session.Write(proto.NewMessage().Ok(packet_type).Marshal())
		log.Printf("< (%v) AUTH OK\n", session.UserId)
		server.RegisterSession(session)
		server.NewUserDAO().SetLastConnection(session.UserId, core.GetCurrentTimeMillis())
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
	log.Printf("> (%v) CREATE EVENT %v\n", session.UserId, msg)

	userDAO := server.NewUserDAO()

	author, err := userDAO.Load(session.UserId)
	if err != nil {
		session.Write(proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal())
		log.Printf("< (%v) CREATE EVENT AUTHOR ERROR %v\n", session.UserId, err)
		return
	}

	if _, err := author.IsValid(); err != nil {
		session.Write(proto.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED).Marshal())
		log.Printf("< (%v) CREATE EVENT AUTHOR ERROR %v\n", session.UserId, err)
		return
	}

	// Check event creation is inside creation window
	currentDate := core.UnixMillisToTime(core.GetCurrentTimeSeconds())
	createdDate := core.UnixMillisToTime(msg.CreatedDate)

	if createdDate.Before(currentDate.Add(-time.Minute)) || createdDate.After(currentDate.Add(time.Minute)) {
		session.Write(proto.NewMessage().Error(packet_type, proto.E_EVENT_OUT_OF_CREATE_WINDOW).Marshal())
		log.Printf("< (%v) CREATE EVENT ERROR OUT OF WINDOW\n", session.UserId)
		return
	}

	event := core.CreateNewEvent(server.GetNewID(), author.Id, author.Name, msg.CreatedDate, msg.StartDate, msg.EndDate, msg.Message)

	if _, err := event.IsValid(); err != nil {
		session.Write(proto.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_INVALID_INPUT)).Marshal())
		log.Printf("< (%v) CREATE EVENT ERROR %v\n", session.UserId, err)
		return
	}

	// Prepare participants
	participantsList, warning, err := server.createParticipantsList(author.Id, msg.Participants)

	if err != nil {
		session.Write(proto.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED)).Marshal())
		log.Printf("< (%v) CREATE EVENT ERROR %v\n", session.UserId, err)
		return
	}

	// Add author as another participant of the event and assume he or she
	// will assist by default
	participant := author.AsParticipant()
	participant.SetFields(core.AttendanceResponse_ASSIST, core.MessageStatus_NO_DELIVERED)
	participantsList[author.Id] = participant

	event.SetParticipants(participantsList)

	if err := server.PublishEvent(event); err == nil {
		session.Write(proto.NewMessage().Ok(packet_type).Marshal())
		log.Printf("< (%v) CREATE EVENT OK (eventId: %v Num.Participants: %v)\n", session.UserId, event.EventId, len(event.Participants))
	} else if err == ErrParticipantsRequired {
		session.Write(proto.NewMessage().Error(packet_type, getNetErrorCode(warning, proto.E_EVENT_PARTICIPANTS_REQUIRED)).Marshal())
		log.Printf("< (%v) CREATE EVENT ERROR %v %v\n", session.UserId, err, warning)
	} else {
		session.Write(proto.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED)).Marshal())
		log.Printf("< (%v) CREATE EVENT ERROR %v\n", session.UserId, err)
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
	log.Printf("> (%v) CONFIRM ATTENDANCE %v\n", session.UserId, msg)

	server := session.Server
	event_dao := server.NewEventDAO()

	// Preconditions: User must have received the invitationm, so user must be in the event participant list
	// or user has the event in his inbox
	participant, err := event_dao.LoadParticipant(msg.EventId, session.UserId)
	var reply []byte

	if err != nil {
		if err == dao.ErrNotFound {
			reply = proto.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_INVALID_EVENT_OR_PARTICIPANT).Marshal()
			log.Printf("< (%v) CONFIRM ATTENDANCE %v INVALID_EVENT_OR_PARTICIPANT\n", session.UserId, msg.EventId)
		} else {
			reply = proto.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_OPERATION_FAILED).Marshal()
			log.Printf("< (%v) CONFIRM ATTENDANCE %v ERROR %v\n", session.UserId, msg.EventId, err)
		}
		session.Write(reply)
		return
	}

	// If the stored response is the same as the provided, send OK response inmediately
	if participant.Response == msg.ActionCode {
		reply = proto.NewMessage().Ok(packet_type).Marshal()
		session.Write(reply)
		return
	}

	if err := event_dao.SetParticipantResponse(session.UserId, msg.EventId, msg.ActionCode); err != nil {
		reply = proto.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_OPERATION_FAILED).Marshal()
		session.Write(reply)
		log.Printf("< (%v) CONFIRM ATTENDANCE %v ERROR %v\n", session.UserId, msg.EventId, err)
		return
	}

	// Send OK Response
	participant.Response = msg.ActionCode
	reply = proto.NewMessage().Ok(packet_type).Marshal()
	session.Write(reply)
	log.Printf("< (%v) CONFIRM ATTENDANCE %v OK\n", session.UserId, msg.EventId)

	// Notify participants
	if event, err := event_dao.LoadEventAndParticipants(msg.EventId); err == nil && len(event) > 0 {
		task := &NotifyParticipantChange{
			Event: event[0],
		}
		task.ParticipantsChanged = append(task.ParticipantsChanged, session.UserId)
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

func onOk(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

	msg := message.(*proto.Ok)
	log.Println("> OK", msg.Type)

	/*switch msg.Type {
	case proto.M_INVITATION_RECEIVED:

	}*/

}

func onPing(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
	checkAuthenticated(session)
	msg := message.(*proto.Ping)
	log.Printf("> (%v) PING %v\n", session.UserId, msg.CurrentTime)
	reply := proto.NewMessage().Pong().Marshal()
	session.Write(reply)
}

func onPong(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
	checkAuthenticated(session)
	msg := message.(*proto.Pong)
	log.Printf("> (%v) PONG %v\n", session.UserId, msg.CurrentTime)
}
