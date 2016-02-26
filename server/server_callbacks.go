package main

import (
	"crypto/sha256"
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

	server := session.Server
	msg := message.(*proto.CreateUserAccount)
	log.Printf("> (%v) USER CREATE ACCOUNT %v\n", session, msg)

	checkUnauthenticated(session)

	// Create new user account
	user := core.NewUserAccount(server.GetNewID(), msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)

	// Check if its a valid user, so the input was correct
	if _, err := user.IsValid(); err != nil {
		error_code := getNetErrorCode(err, proto.E_INVALID_INPUT)
		session.Write(session.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, error_code))
		log.Printf("< (%v) CREATE ACCOUNT INVALID USER: %v\n", session, err)
		return
	}

	// Because an attacker could use the create account feature in order to check if a user exists,
	// a security wait is introduced here to prevent an attacker massive e-mail checking. It's
	// ridiculous but it's my only anti-flood protection so far
	time.Sleep(SECURITY_WAIT_TIME)

	userDAO := server.NewUserDAO()
	var reply *proto.AyiPacket

	// If it's a Facebook account (fbid and fbtoken are not empty) check token
	if user.HasFacebookCredentials() {

		fbsession := fb.NewSession(user.Fbtoken)

		if _, err := fb.CheckAccess(user.Fbid, fbsession); err != nil {
			reply = session.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, proto.E_FB_INVALID_ACCESS)
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
	// FAILED message is sent so far. Read UserDAO.insert for more info.
	if err := userDAO.Insert(user); err != nil {
		err_code := getNetErrorCode(err, proto.E_OPERATION_FAILED)
		reply = session.NewMessage().Error(proto.M_USER_CREATE_ACCOUNT, err_code)
		session.Write(reply)
		log.Printf("< (%v) CREATE ACCOUNT INSERT ERROR: %v\n", session, err)
		return
	}

	// Insert OK
	reply = session.NewMessage().UserAccessGranted(user.Id, user.AuthToken)

	// Import Facebook friends that uses AreYouIN if needed
	if user.HasFacebookCredentials() {
		task := &ImportFacebookFriends{
			TargetUser: user,
			Fbtoken:    user.Fbtoken,
		}
		server.task_executor.Submit(task)
	}

	session.Write(reply)
	log.Printf("< (%v) CREATE ACCOUNT OK\n", session)
}

// FIXME: Renew token should also authenticate the user without needing to get the user to call
// authenticate.
func onUserNewAuthToken(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.NewAuthToken)
	log.Printf("> (%v) USER NEW AUTH TOKEN %v\n", session, msg)

	checkUnauthenticated(session)

	var reply *proto.AyiPacket

	if msg.Type != proto.AuthType_A_NATIVE && msg.Type != proto.AuthType_A_FACEBOOK {
		reply = session.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_MALFORMED_MESSAGE)
		session.Write(reply)
		log.Printf("< (%v) USER NEW AUTH TOKEN MALFORMED MESSAGE\n", session)
		return
	}

	userDAO := server.NewUserDAO()

	// Get new token by e-mail and password
	// NOTE: Review
	if msg.Type == proto.AuthType_A_NATIVE {

		if user_id, err := userDAO.GetIDByEmailAndPassword(msg.Pass1, msg.Pass2); err == nil {
			new_auth_token := uuid.NewV4()
			if err := userDAO.SetAuthToken(user_id, new_auth_token); err == nil {
				reply = session.NewMessage().UserAccessGranted(user_id, new_auth_token)
				log.Printf("< (%v) USER NEW AUTH ACCESS GRANTED\n", session)
			} else {
				reply = session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED)
				log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			}
		} else if err == dao.ErrNotFound {
			reply = session.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD\n", session)
		} else {
			reply = session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED)
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
		}

		session.Write(reply)

	} else if msg.Type == proto.AuthType_A_FACEBOOK {

		// Get new token by Facebook User ID and Facebook Access Token
		// In this context, E_FB_INVALID_USER_OR_PASSWORD means that account does not exist or
		// it is an invalid account.

		// Use Facebook servers to check if the id and token are valid

		fbsession := fb.NewSession(msg.Pass2)

		if _, err := fb.CheckAccess(msg.Pass1, fbsession); err != nil {
			reply = session.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_FB_INVALID_ACCESS)
			session.Write(reply)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID FB ACCESS %v\n", session, fb.GetErrorMessage(err))
			return
		}

		// Check if Facebook user exists also in AreYouIN, i.e. there is a Fbid pointing
		// to a user id

		user_id, err := userDAO.GetIDByFacebookID(msg.Pass1)

		if err == dao.ErrNotFound {
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD", session)
			session.Write(session.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD))
			return
		} else if err != nil {
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			session.Write(session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED))
			return
		}

		// Moreover, check if this user.fbid match provided fb id

		user, err := userDAO.Load(user_id)
		if err != nil {
			session.Write(session.NewMessage().Error(packet_type, proto.E_INVALID_USER_OR_PASSWORD))
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			return
		}

		if user.Fbid != msg.Pass1 {
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD", session)
			session.Write(session.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD))
		}

		// Check that account linked to given Facebook ID is valid, i.e. it has user_email_credentials (with or without
		// password). It may happen that a row in user_email_credentials exists but does not have set a password. Moreover,
		// it may not have Facebook either and it would still be valid. This behaviour is preferred because if
		// this state is found, something had have to be wrong. Under normal conditions, that state should have never
		// happened. So, at this point only existence of e-mail are checked (credentials are ignored).

		if _, err := userDAO.CheckValidAccountObject(user.Id, user.Email, user.Fbid, false); err != nil {
			reply = session.NewMessage().Error(proto.M_USER_NEW_AUTH_TOKEN, proto.E_INVALID_USER_OR_PASSWORD)
			session.Write(reply)
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			return
		}

		new_auth_token := uuid.NewV4()

		if err := userDAO.SetAuthTokenAndFBToken(user_id, new_auth_token, msg.Pass1, msg.Pass2); err != nil {
			reply = session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED)
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			session.Write(reply)
			return
		}

		reply = session.NewMessage().UserAccessGranted(user_id, new_auth_token)
		session.Write(reply)
		log.Printf("< (%v) USER NEW AUTH TOKEN ACCESS GRANTED\n", session)
	}
}

func onUserAuthentication(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.AccessToken)
	log.Printf("> (%v) USER AUTH %v\n", session, msg)

	checkUnauthenticated(session)

	userDAO := server.NewUserDAO()
	user_id := msg.UserId

	user_account, err := userDAO.Load(user_id)
	if err == dao.ErrNotFound || (err == nil && user_account.AuthToken.String() != msg.AuthToken) {
		sendAuthError(session)
		return
	} else if err != nil {
		session.Write(session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) AUTH ERROR %v\n", session, err)
		return
	}

	session.IsAuth = true
	session.UserId = user_id
	session.IIDToken = user_account.IIDtoken
	session.Write(session.NewMessage().Ok(packet_type))
	log.Printf("< (%v) AUTH OK\n", session.UserId)
	server.RegisterSession(session)
	server.NewUserDAO().SetLastConnection(session.UserId, core.GetCurrentTimeMillis())
	server.task_executor.Submit(&SendUserFriends{UserId: user_id})
	// FIXME: Do not send all of the private events, but limit to a fixed number
	sendPrivateEvents(session)
}

func onNewAccessToken(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) REQUEST NEW ACCESS TOKEN\n", session.UserId)

	checkAuthenticated(session)

	accessTokenDAO := server.NewAccessTokenDAO()
	new_access_token := uuid.NewV4()

	// Overwrites previous one if exists
	err := accessTokenDAO.Insert(session.UserId, new_access_token.String())
	if err != nil {
		reply := session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED)
		log.Printf("< (%v) REQUEST NEW ACCESS TOKEN ERROR: %v\n", session.UserId, err)
		session.Write(reply)
		return
	}

	session.Write(session.NewMessage().NewAccessToken(session.UserId, new_access_token))
	log.Printf("< (%v) ACCESS TOKEN: %v\n", session.UserId, new_access_token)
}

func onIIDTokenReceived(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.InstanceIDToken)
	log.Printf("> (%v) IID TOKEN %v\n", session.UserId, msg)

	checkAuthenticated(session)

	if msg.Token == "" || len(msg.Token) < 10 {
		reply := session.NewMessage().Error(packet_type, proto.E_INVALID_INPUT)
		log.Printf("< (%v) IID TOKEN ERROR: INVALID INPUT\n", session.UserId)
		session.Write(reply)
		return
	}

	userDAO := server.NewUserDAO()
	if err := userDAO.SetIIDToken(session.UserId, msg.Token); err != nil {
		reply := session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED)
		log.Printf("< (%v) IID TOKEN ERROR: %v\n", session.UserId, err)
		session.Write(reply)
		return
	}

	session.Write(session.NewMessage().Ok(packet_type))
	log.Printf("< (%v) IID TOKEN OK\n", session.UserId)
}

func onGetUserAccount(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) GET USER ACCOUNT\n", session.UserId)

	checkAuthenticated(session)

	userDAO := server.NewUserDAO()
	user, err := userDAO.LoadWithPicture(session.UserId)

	if err != nil {
		reply := session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) GET USER ACCOUNT ERROR: %v\n", session.UserId, err)
		session.Write(reply)
		return
	}

	packet := session.NewMessage().UserAccount(user)
	session.Write(packet)
	log.Printf("< (%v) SEND USER ACCOUNT INFO (%v bytes)\n", session.UserId, len(user.Picture))
}

func onChangeProfilePicture(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.UserAccount)
	log.Printf("> (%v) CHANGE PROFILE PICTURE (%v bytes)\n", session.UserId, len(msg.Picture))

	checkAuthenticated(session)

	// Compute digest for picture
	digest := sha256.Sum256(msg.Picture)

	picture := &core.Picture{
		RawData: msg.Picture,
		Digest:  digest[:],
	}

	var err error

	// Add or remove profile picture
	if len(msg.Picture) != 0 {
		err = server.saveProfilePicture(session.UserId, picture)
	} else {
		picture.Digest = nil
		err = server.removeProfilePicture(session.UserId, picture)
	}

	if err != nil {
		reply := session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) CHANGE PROFILE PICTURE ERROR: %v\n", session.UserId, err)
		session.Write(reply)
		return
	}

	session.Write(session.NewMessage().OkWithPayload(packet_type, picture.Digest))
	log.Printf("< (%v) PROFILE PICTURE CHANGED\n", session.UserId)
}

func onCreateEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateEvent)
	log.Printf("> (%v) CREATE EVENT %v\n", session.UserId, msg)

	checkAuthenticated(session)

	userDAO := server.NewUserDAO()

	author, err := userDAO.Load(session.UserId)
	if err != nil {
		session.Write(session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) CREATE EVENT AUTHOR ERROR %v\n", session.UserId, err)
		return
	}

	if _, err := author.IsValid(); err != nil {
		session.Write(session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) CREATE EVENT AUTHOR ERROR %v\n", session.UserId, err)
		return
	}

	// Check event creation is inside creation window
	currentDate := core.UnixMillisToTime(core.GetCurrentTimeSeconds())
	createdDate := core.UnixMillisToTime(msg.CreatedDate)

	if createdDate.Before(currentDate.Add(-time.Minute)) || createdDate.After(currentDate.Add(time.Minute)) {
		session.Write(session.NewMessage().Error(packet_type, proto.E_EVENT_OUT_OF_CREATE_WINDOW))
		log.Printf("< (%v) CREATE EVENT ERROR OUT OF WINDOW\n", session.UserId)
		return
	}

	event := core.CreateNewEvent(server.GetNewID(), author.Id, author.Name, msg.CreatedDate, msg.StartDate, msg.EndDate, msg.Message)

	if _, err := event.IsValid(); err != nil {
		session.Write(session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_INVALID_INPUT)))
		log.Printf("< (%v) CREATE EVENT ERROR %v\n", session.UserId, err)
		return
	}

	// Load users participants
	userParticipants, warning, err := server.loadUserParticipants(author.Id, msg.Participants)
	if err != nil {
		session.Write(session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED)))
		log.Printf("< (%v) CREATE EVENT ERROR %v\n", session.UserId, err)
		return
	}

	if len(userParticipants) == 0 {
		session.Write(session.NewMessage().Error(packet_type, getNetErrorCode(warning, proto.E_EVENT_PARTICIPANTS_REQUIRED)))
		log.Printf("< (%v) CREATE EVENT ERROR %v %v\n", session.UserId, err, warning)
		return
	}

	// Add author as another participant of the event and assume he or she
	// will assist by default
	participantsList := server.createParticipantList(userParticipants)
	authorParticipant := author.AsParticipant()
	authorParticipant.SetFields(core.AttendanceResponse_ASSIST, core.MessageStatus_NO_DELIVERED) // Assume it's delivered because author created it
	participantsList[author.Id] = authorParticipant

	event.SetParticipants(participantsList)

	if err := server.PublishEvent(event); err == nil {

		ok_msg := session.NewMessage().Ok(packet_type)
		session.Write(ok_msg)
		log.Printf("< (%v) CREATE EVENT OK (eventId: %v, Num.Participants: %v)\n", session.UserId, event.EventId, len(event.Participants))

		event_created_msg := session.NewMessage().EventCreated(event.GetEventWithoutParticipants())
		status_msg := session.NewMessage().AttendanceStatus(event.EventId, server.createParticipantListFromMap(event.Participants))
		session.Write(event_created_msg, status_msg)
		log.Printf("< (%v) SEND NEW EVENT %v\n", author.Id, event.EventId)

		notification := &NotifyEventInvitation{
			Event:  event,
			Target: userParticipants,
		}
		server.task_executor.Submit(notification)

	} else {
		session.Write(session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED)))
		log.Printf("< (%v) CREATE EVENT ERROR %v\n", session.UserId, err)
	}
}

func onCancelEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

func onInviteUsers(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.InviteUsers)
	log.Printf("> (%v) INVITE USERS %v\n", session.UserId, msg)

	checkAuthenticated(session)

	eventDAO := server.NewEventDAO()
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	var reply *proto.AyiPacket

	// Operation failed
	if err != nil {
		reply = session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED)
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v): %v\n", session.UserId, msg.EventId, err)
		session.Write(reply)
		return
	}

	// Event does not exist
	if len(events) == 0 {
		reply = session.NewMessage().Error(packet_type, proto.E_INVALID_EVENT)
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v): EVENT DOES NOT EXIST\n", session.UserId, msg.EventId)
		session.Write(reply)
		return
	}

	// Author mismath
	event := events[0]
	author_id := session.UserId
	if event.AuthorId != author_id {
		reply = session.NewMessage().Error(packet_type, proto.E_EVENT_AUTHOR_MISMATCH)
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v, author_id=%v): AUTHOR MISMATCH\n", session.UserId, msg.EventId, event.AuthorId)
		session.Write(reply)
		return
	}

	// Event started or finished
	current_time := core.GetCurrentTimeMillis()

	if event.StartDate < current_time {
		reply = session.NewMessage().Error(packet_type, proto.E_EVENT_CANNOT_BE_MODIFIED)
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) EVENT CANNOT BE MODIFIED\n", session.UserId, msg.EventId)
		session.Write(reply)
		return
	}

	// Load user participants
	if len(msg.Participants) == 0 {
		session.Write(session.NewMessage().Error(packet_type, proto.E_EVENT_PARTICIPANTS_REQUIRED))
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) PARTICIPANTS REQUIRED\n", session.UserId, msg.EventId)
		return
	}

	participants_id := server.GetNewParticipants(msg.Participants, event)
	userParticipants, warning, err := server.loadUserParticipants(author_id, participants_id)

	if err != nil {
		session.Write(session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED)))
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) %v\n", session.UserId, msg.EventId, err)
		return
	}

	if len(userParticipants) == 0 {
		session.Write(session.NewMessage().Error(packet_type, getNetErrorCode(warning, proto.E_INVALID_PARTICIPANT)))
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) INVALID PARTICIPANTS %v\n", session.UserId, msg.EventId, warning)
		return
	}

	// After check all of the possible erros, finally participants are inserted into the event
	// and users inboxes
	var succeedCounter int
	new_participants := make([]uint64, 0, len(userParticipants))
	old_participants := core.GetParticipantsIdSlice(event.Participants)

	for _, user := range userParticipants {
		participant := user.AsParticipant()
		err = eventDAO.AddOrUpdateEventToUserInbox(participant, event)
		if err == nil {
			event.Participants[participant.UserId] = participant // Change event
			new_participants = append(new_participants, user.Id)
			succeedCounter++
		} else {
			log.Printf("Error inviting participant %v: %v\n", participant.UserId, err)
			delete(userParticipants, user.Id) // Keep only succesfully added participants
		}
	}

	if succeedCounter > 0 {

		session.Write(session.NewMessage().Ok(packet_type))
		log.Printf("< (%v) INVITE USERS OK (event_id=%v, invitations_send=%v, total=%v)\n", session.UserId, msg.EventId, succeedCounter, len(msg.Participants))

		// Notify previous participants of the new ones added
		notify_event_changed := &NotifyParticipantChange{
			Event:               event,
			ParticipantsChanged: new_participants,
			Target:              old_participants,
		}

		server.task_executor.Submit(notify_event_changed)

		notify_invitation := &NotifyEventInvitation{
			Event:  event,
			Target: userParticipants,
		}
		server.task_executor.Submit(notify_invitation)

	} else {
		session.Write(session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED)))
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) Couldn't invite at least one participant\n", session.UserId, msg.EventId)
	}
}

func onCancelUsersInvitation(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	checkAuthenticated(session)

}

// When a ConfirmAttendance message is received, the attendance response of the participant
// in the participant event list is changed and notified to the other participants. It is
// important to note that num_attendees is not changed server-side till the event has started.
// Clients are cool counting attendees :)
func onConfirmAttendance(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ConfirmAttendance)
	log.Printf("> (%v) CONFIRM ATTENDANCE %v\n", session.UserId, msg)

	checkAuthenticated(session)

	event_dao := server.NewEventDAO()

	// Preconditions: User must have received the invitationm, so user must be in the event participant list
	// or user has the event in his inbox
	participant, err := event_dao.LoadParticipant(msg.EventId, session.UserId)
	var reply *proto.AyiPacket

	if err != nil {
		if err == dao.ErrNotFound {
			reply = session.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_INVALID_EVENT_OR_PARTICIPANT)
			log.Printf("< (%v) CONFIRM ATTENDANCE %v INVALID_EVENT_OR_PARTICIPANT\n", session.UserId, msg.EventId)
		} else {
			reply = session.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_OPERATION_FAILED)
			log.Printf("< (%v) CONFIRM ATTENDANCE %v ERROR %v\n", session.UserId, msg.EventId, err)
		}
		session.Write(reply)
		return
	}

	current_time := core.GetCurrentTimeMillis()

	if participant.EventStartDate < current_time {
		reply = session.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_EVENT_CANNOT_BE_MODIFIED)
		log.Printf("< (%v) CONFIRM ATTENDANCE %v EVENT CANNOT BE MODIFIED\n", session.UserId, msg.EventId)
		session.Write(reply)
		return
	}

	// If the stored response is the same as the provided, send OK response inmediately
	if participant.Response == msg.ActionCode {
		reply = session.NewMessage().Ok(packet_type)
		session.Write(reply)
		return
	}

	if err := event_dao.SetParticipantResponse(participant, msg.ActionCode); err != nil {
		reply = session.NewMessage().Error(proto.M_CONFIRM_ATTENDANCE, proto.E_OPERATION_FAILED)
		session.Write(reply)
		log.Printf("< (%v) CONFIRM ATTENDANCE %v ERROR %v\n", session.UserId, msg.EventId, err)
		return
	}

	// Send OK Response
	participant.Response = msg.ActionCode
	reply = session.NewMessage().Ok(packet_type)
	session.Write(reply)
	log.Printf("< (%v) CONFIRM ATTENDANCE %v OK\n", session.UserId, msg.EventId)

	// Notify participants
	if event, err := event_dao.LoadEventAndParticipants(msg.EventId); err == nil && len(event) == 1 {
		task := &NotifyParticipantChange{
			Event:               event[0],
			ParticipantsChanged: []uint64{session.UserId},
			Target:              core.GetParticipantsIdSlice(event[0].Participants),
		}

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

	log.Println("> USER FRIENDS") // Message does not has payload

	checkAuthenticated(session)

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

	msg := message.(*proto.Ok)
	log.Println("> OK", msg.Type)

	checkAuthenticated(session)

	/*switch msg.Type {
	case proto.M_INVITATION_RECEIVED:

	}*/
}

func onClockRequest(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
	log.Printf("> (%v) CLOCK REQUEST\n", session.UserId)
	checkAuthenticated(session)
	session.Write(session.NewMessage().ClockResponse())
	log.Printf("< (%v) CLOCK RESPONSE\n", session.UserId)
}

func onPing(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
	msg := message.(*proto.TimeInfo)
	log.Printf("> (%v) PING %v\n", session.UserId, msg.CurrentTime)
	checkAuthenticated(session)
	session.Write(session.NewMessage().Pong())
	log.Printf("< (%v) PONG\n", session.UserId)
}

func onPong(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
	msg := message.(*proto.TimeInfo)
	checkAuthenticated(session)
	log.Printf("> (%v) PONG %v\n", session.UserId, msg.CurrentTime)
}
