package main

import (
	"crypto/sha256"
	"log"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
	proto "peeple/areyouin/protocol"
	"time"

	"github.com/twinj/uuid"
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
		session.Write(session.NewMessage().Error(packet_type, error_code))
		log.Printf("< (%v) CREATE ACCOUNT INVALID USER: %v\n", session, err)
		return
	}

	userDAO := server.NewUserDAO()
	var reply *proto.AyiPacket

	// If it's a Facebook account (fbid and fbtoken are not empty) check token
	if user.HasFacebookCredentials() {

		fbsession := fb.NewSession(user.Fbtoken)

		if _, err := fb.CheckAccess(user.Fbid, fbsession); err != nil {
			reply = session.NewMessage().Error(packet_type, proto.E_FB_INVALID_ACCESS)
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
		reply = session.NewMessage().Error(packet_type, err_code)
		session.Write(reply)
		log.Printf("< (%v) CREATE ACCOUNT INSERT ERROR: %v\n", session, err)
		return
	}

	// Insert OK

	if user.HasFacebookCredentials() {
		// Load profie picture from Facebook
		server.task_executor.Submit(&LoadFacebookProfilePicture{
			User:    user,
			Fbtoken: user.Fbtoken,
		})
		// Import Facebook friends that uses AreYouIN if needed
		server.task_executor.Submit(&ImportFacebookFriends{
			TargetUser: user,
			Fbtoken:    user.Fbtoken,
		})
	}

	reply = session.NewMessage().UserAccessGranted(user.Id, user.AuthToken)
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
		reply = session.NewMessage().Error(packet_type, proto.E_MALFORMED_MESSAGE)
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
			if err = userDAO.SetAuthToken(user_id, new_auth_token); err == nil {
				reply = session.NewMessage().UserAccessGranted(user_id, new_auth_token)
				log.Printf("< (%v) USER NEW AUTH ACCESS GRANTED\n", session)
			} else {
				reply = session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED)
				log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
			}
		} else if err == dao.ErrNotFound {
			reply = session.NewMessage().Error(packet_type, proto.E_INVALID_USER_OR_PASSWORD)
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
			reply = session.NewMessage().Error(packet_type, proto.E_FB_INVALID_ACCESS)
			session.Write(reply)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID FB ACCESS %v\n", session, fb.GetErrorMessage(err))
			return
		}

		// Check if Facebook user exists also in AreYouIN, i.e. there is a Fbid pointing
		// to a user id

		user_id, err := userDAO.GetIDByFacebookID(msg.Pass1)

		if err == dao.ErrNotFound {
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD (1)", session)
			session.Write(session.NewMessage().Error(packet_type, proto.E_INVALID_USER_OR_PASSWORD))
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
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD (2)", session)
			session.Write(session.NewMessage().Error(packet_type, proto.E_INVALID_USER_OR_PASSWORD))
		}

		// Check that account linked to given Facebook ID is valid, i.e. it has user_email_credentials (with or without
		// password). It may happen that a row in user_email_credentials exists but does not have set a password. Moreover,
		// it may not have Facebook either and it would still be valid. This behaviour is preferred because if
		// this state is found, something had have to be wrong. Under normal conditions, that state should have never
		// happened. So, at this point only existence of e-mail are checked (credentials are ignored).

		if _, err := userDAO.CheckValidAccountObject(user.Id, user.Email, user.Fbid, false); err != nil {
			reply = session.NewMessage().Error(packet_type, proto.E_INVALID_USER_OR_PASSWORD)
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

	if session.ProtocolVersion < 2 {
		// FIXME: Do not send all of the private events, but limit to a fixed number
		sendPrivateEvents(session)
	}
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

	var user *core.UserAccount
	var err error

	if session.ProtocolVersion < 2 {

		// Keep this one for compatibility

		user, err = userDAO.LoadWithPicture(session.UserId)

	} else {

		// User account should not include the image but only its digest

		user, err = userDAO.Load(session.UserId)
	}

	if err != nil {
		reply := session.NewMessage().Error(packet_type, getNetErrorCode(err, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) GET USER ACCOUNT ERROR: %v\n", session.UserId, err)
		session.Write(reply)
		return
	}

	session.Write(session.NewMessage().UserAccount(user))
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

func onSyncGroups(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.SyncGroups)
	log.Printf("> (%v) SYNC GROUPS %v\n", session.UserId, msg)

	checkAuthenticated(session)

	// Load server groups

	friendDAO := server.NewFriendDAO()

	// FIXME: All groups are always loaded. However, a subset could be loaded when sync
	// behaviour is not TRUNCATE.
	serverGroups, err := friendDAO.LoadGroupsAndMembers(session.UserId)
	checkNoErrorOrPanic(err)

	// Sync
	server.syncFriendGroups(msg.Owner, serverGroups, msg.Groups, msg.SyncBehaviour)

	session.Write(session.NewMessage().Ok(packet_type))
	log.Printf("< (%v) SYNC GROUPS OK\n", session.UserId)
}

func onGetGroups(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) GET GROUPS\n", session.UserId)

	checkAuthenticated(session)

	friendDAO := server.NewFriendDAO()
	groups, err := friendDAO.LoadGroupsAndMembers(session.UserId)
	checkNoErrorOrPanic(err)

	session.Write(session.NewMessage().GroupsList(groups))
	log.Printf("< (%v) GROUPS LIST (num.groups: %v)\n", session.UserId, len(groups))
}

func onCreateEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateEvent)
	log.Printf("> (%v) CREATE EVENT (message: %v, start: %v, end: %v, invitations: %v, picture: %v bytes)\n",
		session.UserId, msg.Message, msg.StartDate, msg.EndDate, len(msg.Participants), len(msg.Picture))

	checkAuthenticated(session)

	userDAO := server.NewUserDAO()
	author, err := userDAO.Load(session.UserId)
	checkNoErrorOrPanic(err)

	valid, err := userDAO.CheckValidAccountObject(author.Id, author.Email, author.Fbid, true)
	if !valid || err != nil {
		session.Write(session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) CREATE EVENT AUTHOR ERROR (2) %v\n", session.UserId, err)
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

	// Create event object
	event := core.CreateNewEvent(server.GetNewID(), author.Id, author.Name, msg.CreatedDate, msg.StartDate, msg.EndDate, msg.Message)

	if _, err = event.IsValid(); err != nil {
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

	// Set participants
	event.SetParticipants(participantsList)

	// Compute digest for event picture
	digest := sha256.Sum256(msg.Picture)

	picture := &core.Picture{
		RawData: msg.Picture,
		Digest:  digest[:],
	}

	// Finally publish event
	if err := server.PublishEvent(event); err == nil {

		// After that, set event picture if received

		if len(msg.Picture) != 0 {
			err = server.saveEventPicture(event.EventId, picture)
			if err == nil {
				event.PictureDigest = picture.Digest
			} else {
				// Only log error but do nothiing. Event has already been published.
				log.Printf("* (%v) Error saving picture for event %v (%v)\n", session.UserId, event.EventId, err)
			}
		}

		if session.ProtocolVersion >= 2 {

			// From protocol v2 onward, OK message is removed in favour of only one message that
			// includes all of the event info (including participants)

			session.Write(session.NewMessage().EventCreated(event))
			log.Printf("< (%v) CREATE EVENT OK (eventId: %v, Num.Participants: %v)\n", session.UserId, event.EventId, len(event.Participants))

		} else {
			// Keep this code for clients that uses v0 and v1
			session.Write(session.NewMessage().Ok(packet_type))
			log.Printf("< (%v) CREATE EVENT OK (eventId: %v, Num.Participants: %v)\n", session.UserId, event.EventId, len(event.Participants))

			event_created_msg := session.NewMessage().EventCreated(event.GetEventWithoutParticipants())
			status_msg := session.NewMessage().AttendanceStatus(event.EventId, server.createParticipantListFromMap(event.Participants))
			session.Write(event_created_msg, status_msg)
			log.Printf("< (%v) SEND NEW EVENT %v\n", session.UserId, event.EventId)
		}

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

func onChangeEventPicture(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ModifyEvent)
	log.Printf("> (%v) CHANGE EVENT PICTURE %v (%v bytes)\n", session.UserId, msg.EventId, len(msg.Picture))

	checkAuthenticated(session)

	eventDAO := server.NewEventDAO()
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check event exists
	checkAtLeastOneEventOrPanic(events)

	// Check author
	event := events[0]
	author_id := session.UserId
	checkEventAuthorOrPanic(author_id, event)

	// Event can be modified
	checkEventWritableOrPanic(event)

	// Create picture object
	picture := &core.Picture{
		RawData: msg.Picture,
	}

	// Set or remove picture

	if len(msg.Picture) > 0 {

		// Compute digest for event picture

		digest := sha256.Sum256(msg.Picture)
		picture.Digest = digest[:]

		// Save picture

		err = server.saveEventPicture(event.EventId, picture)
		checkNoErrorOrPanic(err)

	} else {

		// Remove picture

		picture.Digest = nil
		picture.RawData = make([]byte, 0)
		err = server.removeEventPicture(event.EventId, picture)
		checkNoErrorOrPanic(err)
	}

	event.PictureDigest = picture.Digest

	// Send ACK to caller
	session.Write(session.NewMessage().Ok(packet_type))
	log.Printf("< (%v) EVENT PICTURE CHANGED\n", session.UserId)

	// Notify change to participants
	server.task_executor.Submit(&NotifyEventChange{
		Event:  event,
		Target: core.GetParticipantsIdSlice(event.Participants),
	})
}

func onCancelEvent(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CancelEvent)
	log.Printf("> (%v) CANCEL EVENT %v\n", session.UserId, msg)

	checkAuthenticated(session)

	eventDAO := server.NewEventDAO()
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	checkNoErrorOrPanic(err)

	// Event exist
	checkAtLeastOneEventOrPanic(events)

	// Author math
	event := events[0]
	author_id := session.UserId
	checkEventAuthorOrPanic(author_id, event)

	// Event can be modified
	checkEventWritableOrPanic(event)

	// Change event state and position in time line
	new_inbox_position := core.GetCurrentTimeMillis()
	err = eventDAO.SetEventStateAndInboxPosition(event.EventId, core.EventState_CANCELLED, new_inbox_position)
	checkNoErrorOrPanic(err)

	// Moreover, change event position inside user inbox so that next time client request recent events
	// this one is ignored.
	/*for _, participant := range event.Participants {
		err := eventDAO.SetUserEventInboxPosition(participant, event, new_inbox_position)
		if err != nil {
			log.Println("onCancelEvent Error:", err) // FIXME: Add retry logic
		}
	}*/

	// Update event object
	event.InboxPosition = new_inbox_position
	event.State = core.EventState_CANCELLED

	// FIXME: Could send directly the event canceled message, and ignore author from
	// NotifyEventCancelled task
	log.Printf("< (%v) CANCEL EVENT OK\n", session.UserId)
	session.Write(session.NewMessage().Ok(packet_type))

	// Notify participants
	server.task_executor.Submit(&NotifyEventCancelled{
		CancelledBy: session.UserId,
		Event:       event,
	})
}

func onInviteUsers(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.InviteUsers)
	log.Printf("> (%v) INVITE USERS %v\n", session.UserId, msg)

	checkAuthenticated(session)

	// First of all, check participants
	if len(msg.Participants) == 0 {
		session.Write(session.NewMessage().Error(packet_type, proto.E_EVENT_PARTICIPANTS_REQUIRED))
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) PARTICIPANTS REQUIRED\n", session.UserId, msg.EventId)
		return
	}

	eventDAO := server.NewEventDAO()
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	checkNoErrorOrPanic(err)

	// Event does not exist
	checkAtLeastOneEventOrPanic(events)

	// Author mismath
	event := events[0]
	author_id := session.UserId
	checkEventAuthorOrPanic(author_id, event)

	// Event can be modified
	checkEventWritableOrPanic(event)

	// Load user participants
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

		_, err = eventDAO.CompareAndSetNumGuests(event.EventId, len(event.Participants))
		if err != nil {
			log.Println("Invite Users: Update Num. guestss Error:", err)
		}

		event.NumGuests = int32(len(event.Participants))

		session.Write(session.NewMessage().Ok(packet_type))
		log.Printf("< (%v) INVITE USERS OK (event_id=%v, invitations_send=%v, total=%v)\n", session.UserId, msg.EventId, succeedCounter, len(msg.Participants))

		// Notify previous participants of the new ones added
		server.task_executor.Submit(&NotifyParticipantChange{
			Event:               event,
			ParticipantsChanged: new_participants,
			Target:              old_participants,
			NumGuests:           event.NumGuests, // Include also total NumGuests because it's changed
		})

		server.task_executor.Submit(&NotifyEventInvitation{
			Event:  event,
			Target: userParticipants,
		})

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

	// Preconditions: User must have received the invitation, so user must be in the event participant list
	// and user has the event in his inbox
	participant, err := event_dao.LoadParticipant(msg.EventId, session.UserId)
	checkNoErrorOrPanic(err)

	// Event can be modified
	current_time := core.GetCurrentTimeMillis()

	if participant.StartDate < current_time || participant.EventState == core.EventState_CANCELLED {
		log.Printf("< (%v) CONFIRM ATTENDANCE %v EVENT CANNOT BE MODIFIED\n", session.UserId, msg.EventId)
		session.Write(session.NewMessage().Error(packet_type, proto.E_EVENT_CANNOT_BE_MODIFIED))
		return
	}

	// If the stored response is the same as the provided, send OK response inmediately
	if participant.Response == msg.ActionCode {
		session.Write(session.NewMessage().Ok(packet_type))
		return
	}

	if err := event_dao.SetParticipantResponse(participant, msg.ActionCode); err != nil {
		session.Write(session.NewMessage().Error(packet_type, proto.E_OPERATION_FAILED))
		log.Printf("< (%v) CONFIRM ATTENDANCE %v ERROR %v\n", session.UserId, msg.EventId, err)
		return
	}

	// Send OK Response
	participant.Response = msg.ActionCode
	session.Write(session.NewMessage().Ok(packet_type))
	log.Printf("< (%v) CONFIRM ATTENDANCE %v OK\n", session.UserId, msg.EventId)

	// Notify participants
	if event, err := event_dao.LoadEventAndParticipants(msg.EventId); err == nil && len(event) == 1 {

		server.task_executor.Submit(&NotifyParticipantChange{
			Event:               event[0],
			ParticipantsChanged: []uint64{session.UserId},
			Target:              core.GetParticipantsIdSlice(event[0].Participants),
		})

	} else {
		log.Println("onConfirmAttendance:", err)
	}
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
	log.Println("> OK")
	checkAuthenticated(session)
}

func onRequestPrivateEvents(packet_type proto.PacketType, message proto.Message, session *AyiSession) {

	log.Printf("> (%v) REQUEST PRIVATE EVENTS\n", session.UserId) // Message does not has payload
	checkAuthenticated(session)

	server := session.Server
	eventDAO := server.NewEventDAO()

	current_time := core.GetCurrentTimeMillis()
	events, err := eventDAO.LoadUserEventsAndParticipants(session.UserId, current_time)

	if err != nil {
		if err ==  dao.ErrEmptyInbox {
			log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session.UserId, 0)
			session.Write(session.NewMessage().EventsList(nil))
			return
		} else {
			panic(err)
		}
	}

	// Filter event's participant list
	for _, event := range events {

		if len(event.Participants) == 0 {
			log.Printf("WARNING: Event %v has zero participants\n", event.EventId)
			continue
		}

		event.Participants = session.Server.filterEventParticipants(session.UserId, event.Participants)
	}

	// Send event list to user
	log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session.UserId, len(events))
	session.Write(session.NewMessage().EventsList(events))

	// Update delivery status
	for _, event := range events {

		ownParticipant, ok := event.Participants[session.UserId]

		if ok && ownParticipant.Delivered != core.MessageStatus_CLIENT_DELIVERED {

			ownParticipant.Delivered = core.MessageStatus_CLIENT_DELIVERED
			eventDAO.SetParticipantStatus(session.UserId, event.EventId, ownParticipant.Delivered)

			// Notify change in participant status to the other participants
			task := &NotifyParticipantChange{
				Event:               event,
				ParticipantsChanged: []uint64{session.UserId},
				Target:              core.GetParticipantsIdSlice(event.Participants),
			}

			// I'm also sending notification to the author. Could avoid this because author already knows
			// that the event has been send to him
			server.task_executor.Submit(task)
		}
	}
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
