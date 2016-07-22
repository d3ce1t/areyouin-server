package main

import (
	"log"
	core "peeple/areyouin/common"
	"peeple/areyouin/model"
	"peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
	proto "peeple/areyouin/protocol"
	"github.com/twinj/uuid"
)

func onCreateAccount(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateUserAccount)
	log.Printf("> (%v) USER CREATE ACCOUNT (email: %v, fbId: %v)\n", session, msg.Email, msg.Fbid)

	checkUnauthenticated(session)

	// Create new user account

	userAccount, err := server.Model.Accounts.CreateUserAccount(msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)
	if err != nil {
		error_code := getNetErrorCode(err, proto.E_OPERATION_FAILED)
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), error_code))
		log.Printf("< (%v) CREATE ACCOUNT ERROR: %v\n", session, err)
		return
	}

	if len(msg.Picture) != 0 {
		if err := server.Model.Accounts.ChangeProfilePicture(userAccount, msg.Picture); err != nil {
			log.Printf("* (%v) CREATE ACCOUNT: SET PROFILE PICTURE ERROR: %v\n", session, err)
		} else {
			log.Printf("* (%v) CREATE ACCOUNT: PROFILE PICTURE SET\n", session)
		}
	}

	if userAccount.HasFacebookCredentials() {

		// Import Facebook friends that uses AreYouIN if needed

		server.task_executor.Submit(&ImportFacebookFriends{
			TargetUser: userAccount,
			Fbtoken:    userAccount.Fbtoken,
		})
	}

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().UserAccessGranted(userAccount.Id, userAccount.AuthToken))
	log.Printf("< (%v) CREATE ACCOUNT OK\n", session)
}

func onUserNewAuthToken(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.NewAuthToken)
	log.Printf("> (%v) USER NEW AUTH TOKEN %v\n", session, msg)

	checkUnauthenticated(session)

	var reply *proto.AyiPacket

	if msg.Type != proto.AuthType_A_NATIVE && msg.Type != proto.AuthType_A_FACEBOOK {
		reply = session.NewMessage().Error(request.Type(), proto.E_MALFORMED_MESSAGE)
		session.WriteResponse(request.Header.GetToken(), reply)
		log.Printf("< (%v) USER NEW AUTH TOKEN MALFORMED MESSAGE\n", session)
		return
	}

	if msg.Type == proto.AuthType_A_NATIVE {

		// Get new token by e-mail and password

		auth_cred, err := server.Model.Accounts.NewAuthCredentialByEmailAndPassword(msg.Pass1, msg.Pass2)

		if err == nil {
			reply = session.NewMessage().UserAccessGranted(auth_cred.UserId, auth_cred.Token)
			log.Printf("< (%v) USER NEW AUTH TOKEN ACCESS GRANTED\n", session)
		} else if err == model.ErrInvalidUserOrPassword {
			reply = session.NewMessage().Error(request.Type(), proto.E_INVALID_USER_OR_PASSWORD)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD\n", session)
		} else {
			reply = session.NewMessage().Error(request.Type(), proto.E_OPERATION_FAILED)
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
		}

	} else if msg.Type == proto.AuthType_A_FACEBOOK {

		// Get new token by Facebook User ID and Facebook Access Token
		// In this context, E_INVALID_USER_OR_PASSWORD means that account
		// does not exist or it is an invalid account.

		auth_cred, err := server.Model.Accounts.NewAuthCredentialByFacebook(msg.Pass1, msg.Pass2)

		if err == nil {
			reply = session.NewMessage().UserAccessGranted(auth_cred.UserId, auth_cred.Token)
			log.Printf("< (%v) USER NEW AUTH TOKEN ACCESS GRANTED\n", session)
		} else if err == fb.ErrFacebookAccessForbidden {
			reply = session.NewMessage().Error(request.Type(), proto.E_FB_INVALID_ACCESS)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID FB ACCESS %v\n", session, fb.GetErrorMessage(err))
		} else if err == model.ErrInvalidUserOrPassword {
			reply = session.NewMessage().Error(request.Type(), proto.E_INVALID_USER_OR_PASSWORD)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD", session)
		} else {
			reply = session.NewMessage().Error(request.Type(), proto.E_OPERATION_FAILED)
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR %v\n", session, err)
		}

	}

	session.WriteResponse(request.Header.GetToken(), reply)
}

func onUserAuthentication(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.AccessToken)
	log.Printf("> (%v) USER AUTH %v\n", session, msg)

	checkUnauthenticated(session)

	isAuthenticated, err := server.Model.Accounts.AuthenticateUser(msg.UserId, msg.AuthToken)
	if err != nil {
		errorCode := getNetErrorCode(err, proto.E_OPERATION_FAILED)
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), errorCode))
		log.Printf("< (%v) AUTH ERROR %v\n", session, err)
		return
	}

	userDAO := dao.NewUserDAO(server.DbSession)
	iidToken, err := userDAO.GetIIDToken(msg.UserId)
	checkNoErrorOrPanic(err)

	session.IsAuth = isAuthenticated
	session.UserId = msg.UserId
	session.IIDToken = iidToken
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) AUTH OK\n", session)
	server.RegisterSession(session)
	userDAO.SetLastConnection(session.UserId, core.GetCurrentTimeMillis())

	if session.ProtocolVersion < 2 {
		sendPrivateEvents(session)
	}
}

func onNewAccessToken(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) REQUEST NEW ACCESS TOKEN\n", session)

	checkAuthenticated(session)

	accessTokenDAO := dao.NewAccessTokenDAO(server.DbSession)
	new_access_token := uuid.NewV4().String()

	// Overwrites previous one if exists
	err := accessTokenDAO.Insert(session.UserId, new_access_token)
	if err != nil {
		reply := session.NewMessage().Error(request.Type(), proto.E_OPERATION_FAILED)
		log.Printf("< (%v) REQUEST NEW ACCESS TOKEN ERROR: %v\n", session, err)
		session.WriteResponse(request.Header.GetToken(), reply)
		return
	}

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().NewAccessToken(session.UserId, new_access_token))
	log.Printf("< (%v) ACCESS TOKEN: %v\n", session, new_access_token)
}

func onIIDTokenReceived(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.InstanceIDToken)
	log.Printf("> (%v) IID TOKEN %v\n", session, msg)

	checkAuthenticated(session)

	if msg.Token == "" || len(msg.Token) < 10 {
		reply := session.NewMessage().Error(request.Type(), proto.E_INVALID_INPUT)
		log.Printf("< (%v) IID TOKEN ERROR: INVALID INPUT\n", session)
		session.WriteResponse(request.Header.GetToken(), reply)
		return
	}

	userDAO := dao.NewUserDAO(server.DbSession)
	iidToken := &core.IIDToken{Token: msg.Token, Version: int(session.ProtocolVersion)}
	if err := userDAO.SetIIDToken(session.UserId, iidToken); err != nil {
		reply := session.NewMessage().Error(request.Type(), proto.E_OPERATION_FAILED)
		log.Printf("< (%v) IID TOKEN ERROR: %v\n", session, err)
		session.WriteResponse(request.Header.GetToken(), reply)
		return
	}

	session.IIDToken = iidToken
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) IID TOKEN OK\n", session)
}

func onGetUserAccount(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) GET USER ACCOUNT\n", session)

	checkAuthenticated(session)

	userDAO := dao.NewUserDAO(server.DbSession)
	user, err := userDAO.Load(session.UserId) // Do not include picture
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().UserAccount(user))
	log.Printf("< (%v) SEND USER ACCOUNT INFO\n", session)
}

func onChangeProfilePicture(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.UserAccount)
	log.Printf("> (%v) CHANGE PROFILE PICTURE (%v bytes)\n", session, len(msg.Picture))

	checkAuthenticated(session)

	// Load account
	userDAO := dao.NewUserDAO(server.DbSession)
	user, err := userDAO.Load(session.UserId)
	checkNoErrorOrPanic(err)

	// Add or remove profile picture
	err = server.Model.Accounts.ChangeProfilePicture(user, msg.Picture)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) PROFILE PICTURE CHANGED\n", session)
	session.Write(session.NewMessage().UserAccount(user))
	log.Printf("< (%v) SEND USER ACCOUNT INFO\n", session)
}

func onSyncGroups(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.SyncGroups)
	log.Printf("> (%v) SYNC GROUPS %v\n", session, msg)

	checkAuthenticated(session)

	// Load server groups

	friendDAO := dao.NewFriendDAO(server.DbSession)

	// FIXME: All groups are always loaded. However, a subset could be loaded when sync
	// behaviour is not TRUNCATE.
	serverGroups, err := friendDAO.LoadGroupsAndMembers(session.UserId)
	checkNoErrorOrPanic(err)

	// Sync
	server.syncFriendGroups(msg.Owner, serverGroups, msg.Groups, msg.SyncBehaviour)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) SYNC GROUPS OK\n", session)
}

func onGetGroups(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) GET GROUPS\n", session)

	checkAuthenticated(session)

	friendDAO := dao.NewFriendDAO(server.DbSession)
	groups, err := friendDAO.LoadGroupsAndMembers(session.UserId)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().GroupsList(groups))
	log.Printf("< (%v) GROUPS LIST (num.groups: %v)\n", session, len(groups))
}

func onCreateEvent(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateEvent)
	log.Printf("> (%v) CREATE EVENT (message: %v, start: %v, end: %v, invitations: %v, picture: %v bytes)\n",
		session, msg.Message, msg.StartDate, msg.EndDate, len(msg.Participants), len(msg.Picture))

	checkAuthenticated(session)

	userDAO := dao.NewUserDAO(server.DbSession)
	author, err := userDAO.Load(session.UserId)
	checkNoErrorOrPanic(err)

	event, err := server.Model.Events.CreateNewEvent(author, msg.CreatedDate, msg.StartDate, msg.EndDate, msg.Message)
	checkNoErrorOrPanic(err)

	// Load users participants
	userParticipants, err := server.Model.Events.CreateParticipantsList(author.Id, msg.Participants)
	checkNoErrorOrPanic(err)

	if err = server.Model.Events.PublishEvent(event, userParticipants); err == nil {

		// Event Published: set event picture if received

		if len(msg.Picture) != 0 {
			if err = server.Model.Events.ChangeEventPicture(event, msg.Picture); err != nil {
				// Only log error but do nothing. Event has already been published.
				log.Printf("* (%v) Error saving picture for event %v (%v)\n", session, event.EventId, err)
			}
		}

		// From protocol v2 onward, OK message is removed in favour of only one message that
		// includes all of the event info (including participants)

		session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventCreated(event))
		log.Printf("< (%v) CREATE EVENT OK (eventId: %v, Num.Participants: %v, Remaining.Participants: %v)\n",
			session, event.EventId, len(event.Participants), 1 + len(msg.Participants) - len(event.Participants))

		notification := &NotifyEventInvitation {
			Event:  event,
			Target: userParticipants,
		}
		server.task_executor.Submit(notification)

	}	else if err == model.ErrParticipantsRequired {

		// Participants required

		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), proto.E_EVENT_PARTICIPANTS_REQUIRED))
		log.Printf("< (%v) CREATE EVENT ERROR: THERE ARE NO PARTICIPANTS\n", session)

	} else {

		// Error

		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), getNetErrorCode(err, proto.E_OPERATION_FAILED)))
		log.Printf("< (%v) CREATE EVENT ERROR %v\n", session, err)
	}

}

func onChangeEventPicture(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ModifyEvent)
	log.Printf("> (%v) CHANGE EVENT PICTURE %v (%v bytes)\n", session, msg.EventId, len(msg.Picture))

	checkAuthenticated(session)

	eventDAO := dao.NewEventDAO(server.DbSession)
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check event exists
	checkAtLeastOneEventOrPanic(events)

	// Check author
	event := events[0]
	author_id := session.UserId
	checkEventAuthorOrPanic(author_id, event)

	// Actually change event picture
	err = server.Model.Events.ChangeEventPicture(event, msg.Picture)
	checkNoErrorOrPanic(err)

	// Send ACK to caller
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) EVENT PICTURE CHANGED\n", session)

	// Notify change to participants
	server.task_executor.Submit(&NotifyEventChange{
		Event:  event,
		Target: core.GetParticipantKeys(event.Participants),
	})
}

func onCancelEvent(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CancelEvent)
	log.Printf("> (%v) CANCEL EVENT %v\n", session, msg)

	checkAuthenticated(session)

	eventDAO := dao.NewEventDAO(server.DbSession)
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	checkNoErrorOrPanic(err)

	// Event exist
	checkAtLeastOneEventOrPanic(events)

	// Author math
	event := events[0]
	author_id := session.UserId
	checkEventAuthorOrPanic(author_id, event)

	// Cancel event
	err = server.Model.Events.CancelEvent(event)
	checkNoErrorOrPanic(err)

	// FIXME: Could send directly the event canceled message, and ignore author from
	// NotifyEventCancelled task
	log.Printf("< (%v) CANCEL EVENT OK\n", session)
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))

	// Notify participants
	server.task_executor.Submit(&NotifyEventCancelled{
		CancelledBy: session.UserId,
		Event:       event,
	})
}

func onInviteUsers(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.InviteUsers)
	log.Printf("> (%v) INVITE USERS %v\n", session, msg)

	checkAuthenticated(session)

	// First of all, check participants
	if len(msg.Participants) == 0 {
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), proto.E_EVENT_PARTICIPANTS_REQUIRED))
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) PARTICIPANTS REQUIRED\n", session, msg.EventId)
		return
	}

	eventDAO := dao.NewEventDAO(server.DbSession)
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	checkNoErrorOrPanic(err)

	// Event does exist
	checkAtLeastOneEventOrPanic(events)

	// Author mismath
	event := events[0]
	author_id := session.UserId
	checkEventAuthorOrPanic(author_id, event)

	// Event can be modified
	checkEventWritableOrPanic(event)

	// Load user participants
	participants_id := core.GetNewParticipants(msg.Participants, event)
	newParticipants, err := server.Model.Events.CreateParticipantsList(author_id, participants_id)
	checkNoErrorOrPanic(err)

	// Get current participants for later use
	oldParticipants := core.GetParticipantKeys(event.Participants)

	// Invite participants
	usersInvited, err := server.Model.Events.InviteUsers(event, newParticipants)
	checkNoErrorOrPanic(err)

	// Write response back
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) INVITE USERS OK (event_id: %v, invitations: %v/%v, total: %v)\n",
		session, event.EventId, len(usersInvited), len(newParticipants), len(msg.Participants))

	// Notify previous participants of the new ones added
	server.task_executor.Submit(&NotifyParticipantChange{
		Event:               event,
		ParticipantsChanged: core.GetUserKeys(usersInvited),
		Target:              oldParticipants,
		NumGuests:           event.NumGuests, // Include also total NumGuests because it's changed
	})

	server.task_executor.Submit(&NotifyEventInvitation{
		Event:  event,
		Target: usersInvited,
	})
}

// When a ConfirmAttendance message is received, the attendance response of the participant
// in the participant event list is changed and notified to the other participants. It is
// important to note that num_attendees is not changed server-side till the event has started.
// Clients are cool counting attendees :)
func onConfirmAttendance(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ConfirmAttendance)
	log.Printf("> (%v) CONFIRM ATTENDANCE %v\n", session, msg)

	checkAuthenticated(session)

	eventDAO := dao.NewEventDAO(server.DbSession)
	events, err := eventDAO.LoadEventAndParticipants(msg.EventId)
	checkNoErrorOrPanic(err)

	// Event does exist
	checkAtLeastOneEventOrPanic(events)

	event := events[0]

	// Change response
	changed, err := server.Model.Events.ChangeParticipantResponse(session.UserId, msg.ActionCode, event)
	checkNoErrorOrPanic(err)

	// Send OK Response
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) CONFIRM ATTENDANCE %v OK\n", session, msg.EventId)

	if changed {
		// Notify participants
		server.task_executor.Submit(&NotifyParticipantChange{
			Event:               event,
			ParticipantsChanged: []int64{session.UserId},
			Target:              core.GetParticipantKeys(event.Participants),
		})
	}
}

func onListPrivateEvents(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	log.Printf("> (%v) REQUEST PRIVATE EVENTS\n", session) // Message does not has payload
	checkAuthenticated(session)

	server := session.Server
	events, err := server.Model.Events.GetRecentEvents(session.UserId)

	if err == model.ErrEmptyInbox {
		log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session, 0)
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventsList(nil))
		return
	} else if err != nil {
		panic(err)
	}

	// Send event list to user
	filteredEvents := server.Model.Events.FilterEvents(events, session.UserId)
	log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session, len(filteredEvents))
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventsList(filteredEvents))

	// Update delivery status
	for _, event := range events {

		// TODO: I should receive an ACK before try to change state.
		// Moreover, err is ignored. What should server do in that case?

		changed, _ := server.Model.Events.ChangeDeliveryState(event, session.UserId, core.MessageStatus_CLIENT_DELIVERED)

		if changed {

			// Notify change in participant status to the other participants
			task := &NotifyParticipantChange{
				Event:               event,
				ParticipantsChanged: []int64{session.UserId},
				Target:              core.GetParticipantKeys(event.Participants),
			}

			// I'm also sending notification to the author. Could avoid this
			// because author already knows that the event has been send
			// to him
			server.task_executor.Submit(task)
		}

	}
}

func onListEventsHistory(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.EventListRequest)

	log.Printf("> (%v) REQUEST EVENTS HISTORY (start: %v, end: %v)\n", session, msg.StartWindow, msg.EndWindow) // Message does not has payload
	checkAuthenticated(session)

	events, err := server.Model.Events.GetEventsHistory(session.UserId, msg.StartWindow, msg.EndWindow)
	checkNoErrorOrPanic(err)

	// Check event exists
	checkAtLeastOneEventOrPanic(events)

	var startWindow int64 = 0
	var endWindow int64 = 0

	if msg.StartWindow < msg.EndWindow {
		startWindow = events[0].StartDate
		endWindow = events[len(events)-1].StartDate
	} else {
		startWindow = events[len(events)-1].StartDate
		endWindow = events[0].StartDate
	}

	filteredEvents := server.Model.Events.FilterEvents(events, session.UserId)

	// Send event list to user
	log.Printf("< (%v) SEND EVENTS HISTORY (num.events: %v)", session, len(events))
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventsHistoryList(filteredEvents, startWindow, endWindow))
}

func onGetUserFriends(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server

	log.Printf("> (%v) GET USER FRIENDS\n", session) // Message does not has payload
	checkAuthenticated(session)

	friend_dao := dao.NewFriendDAO(server.DbSession)
	friends, err := friend_dao.LoadFriends(session.UserId, 0) // TODO: Fix this
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().FriendsList(friends))
	log.Printf("< (%v) SEND USER FRIENDS (num.friends: %v)\n", session, len(friends))
}

func onFriendRequest(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateFriendRequest)

	log.Printf("> (%v) CREATE FRIEND REQUEST: %v\n", session, msg)
	checkAuthenticated(session)

	user_dao := dao.NewUserDAO(server.DbSession)
	friend_dao := dao.NewFriendDAO(server.DbSession)

	// Load session.UserId account
	userAccount, err := user_dao.Load(session.UserId)
	checkNoErrorOrPanic(err)

	// Exist user with provided email
	friendAccount, err := user_dao.LoadByEmail(msg.Email)
	if err != nil {
		if err == dao.ErrNotFound {
			panic(ErrFriendNotFound)
		} else {
			panic(err)
		}
	}

	// Not friends
	areFriends, err := server.Model.Accounts.AreFriends(session.UserId, friendAccount.Id)
	checkNoErrorOrPanic(err)

	if areFriends {
		panic(ErrSendRequest_AlreadyFriends)
	}

	// Not exist previous request
	exist_previous_request, err := friend_dao.ExistFriendRequest(session.UserId, friendAccount.Id)
	checkNoErrorOrPanic(err)
	if exist_previous_request {
		panic(ErrSendRequest_AlreadySent)
	}

	// Preconditions satisfied

	created_date := core.GetCurrentTimeMillis()
	err = friend_dao.InsertFriendRequest(friendAccount.Id, userAccount.Id, userAccount.Name,
																					userAccount.Email, created_date)
	checkNoErrorOrPanic(err)

	friendRequest := &core.FriendRequest {
		FriendId: userAccount.Id,
		Name: userAccount.Name,
		Email: userAccount.Email,
		CreatedDate: created_date,
	}

	server.task_executor.Submit(&NotifyFriendRequest{
		UserId:  friendAccount.Id,
		FriendRequest: friendRequest,
	})

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) CREATE FRIEND REQUEST OK\n", session)
}

func onListFriendRequests(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server

	log.Printf("> (%v) GET FRIEND REQUESTS\n", session)
	checkAuthenticated(session)

	friend_dao := dao.NewFriendDAO(server.DbSession)
	requests, err := friend_dao.LoadFriendRequests(session.UserId)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().FriendRequestsList(requests))
	log.Printf("< (%v) SEND FRIEND REQUESTS (num.requests: %v)\n", session, len(requests))
}

func onConfirmFriendRequest(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ConfirmFriendRequest)

	log.Printf("> (%v) CONFIRM FRIEND REQUEST: %v\n", session, msg)
	checkAuthenticated(session)

	friend_dao := dao.NewFriendDAO(server.DbSession)

	// Load request from database in order to check that it exists
	friendRequest, err := friend_dao.LoadFriendRequest(session.UserId, msg.FriendId)
	checkNoErrorOrPanic(err)

	if msg.Response == proto.ConfirmFriendRequest_CONFIRM {

		// Friend Request has been accepted. Make friends.

		user_dao := dao.NewUserDAO(server.DbSession)

		// Load current user
		currentUser, err := user_dao.Load(session.UserId)
		checkNoErrorOrPanic(err)

		// Load friend
		friend, err := user_dao.Load(msg.FriendId)
		checkNoErrorOrPanic(err)

		// Make both friends
		err = friend_dao.MakeFriends(currentUser, friend)
		checkNoErrorOrPanic(err)
		log.Printf("* (%v) User %v and User %v are now friends\n", session, currentUser.Id, friend.Id)

		// Delete Friend Request
		err = friend_dao.DeleteFriendRequest(session.UserId, msg.FriendId, friendRequest.CreatedDate)
		checkNoErrorOrPanic(err)

		// Send OK to user
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
		log.Printf("< (%v) CONFIRM FRIEND REQUEST OK (accepted)\n", session)

		// Send Friend List to both users if connected
		// TODO: In this case, FRIENDS_LIST is sent as a notification. Because of this,
		// it should be only a subset of all friends in server.
		server.task_executor.Submit(&SendUserFriends{UserId: currentUser.Id})
		server.task_executor.Submit(&SendUserFriends{UserId: friend.Id})
		//task.sendGcmNotification(friendUser.Id, friendUser.IIDtoken, task.TargetUser)

	} else if msg.Response == proto.ConfirmFriendRequest_CANCEL {

		// Friend Request has been cancelled. Remove it.

		friend_dao.DeleteFriendRequest(session.UserId, msg.FriendId, friendRequest.CreatedDate)
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
		log.Printf("< (%v) CONFIRM FRIEND REQUEST OK (cancelled)\n", session)

	} else {
		panic(ErrOperationFailed)
	}

}

func onOk(request *proto.AyiPacket, message proto.Message, session *AyiSession) {
	log.Println("> OK")
	checkAuthenticated(session)
}

func onClockRequest(request *proto.AyiPacket, message proto.Message, session *AyiSession) {
	log.Printf("> (%v) CLOCK REQUEST\n", session)
	checkAuthenticated(session)
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().ClockResponse())
	log.Printf("< (%v) CLOCK RESPONSE\n", session)
}

func onPing(request *proto.AyiPacket, message proto.Message, session *AyiSession) {
	msg := message.(*proto.TimeInfo)
	log.Printf("> (%v) PING %v\n", session, msg.CurrentTime)
	checkAuthenticated(session)
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Pong())
	log.Printf("< (%v) PONG\n", session)
}
