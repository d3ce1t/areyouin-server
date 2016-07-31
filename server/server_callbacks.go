package main

import (
	"log"
	"peeple/areyouin/api"
	fb "peeple/areyouin/facebook"
	"peeple/areyouin/model"
	proto "peeple/areyouin/protocol"
	"peeple/areyouin/protocol/core"
)

func onCreateAccount(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateUserAccount)
	log.Printf("> (%v) USER CREATE ACCOUNT (email: %v, fbId: %v)\n", session, msg.Email, msg.Fbid)

	checkUnauthenticated(session)

	// Create new user account

	userAccount, err := server.Model.Accounts.CreateUserAccount(msg.Name, msg.Email, msg.Password, msg.Phone, msg.Fbid, msg.Fbtoken)
	if err != nil {
		errorCode := getNetErrorCode(err, proto.E_OPERATION_FAILED)
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), errorCode))
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

	if userAccount.HasFacebook() {

		// Import Facebook friends that uses AreYouIN if needed

		server.task_executor.Submit(&ImportFacebookFriends{
			TargetUser: userAccount,
			Fbtoken:    userAccount.FbToken(),
		})
	}

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().UserAccessGranted(userAccount.Id(), userAccount.AuthToken()))
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

		authCred, err := server.Model.Accounts.NewAuthCredentialByEmailAndPassword(msg.Pass1, msg.Pass2)

		if err == nil {
			reply = session.NewMessage().UserAccessGranted(authCred.UserID(), authCred.Token())
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

		authCred, err := server.Model.Accounts.NewAuthCredentialByFacebook(msg.Pass1, msg.Pass2)

		if err == nil {
			reply = session.NewMessage().UserAccessGranted(authCred.UserID(), authCred.Token())
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

	iidToken, err := server.Model.Accounts.GetPushToken(msg.UserId)
	checkNoErrorOrPanic(err)

	session.IsAuth = isAuthenticated
	session.UserId = msg.UserId
	session.IIDToken = iidToken
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) AUTH OK\n", session)
	server.registerSession(session)

	server.refreshSessionActivity(session)
}

func onNewAccessToken(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) REQUEST NEW ACCESS TOKEN\n", session)

	checkAuthenticated(session)

	accessToken, err := server.Model.Accounts.NewImageAccessToken(session.UserId)
	if err != nil {
		reply := session.NewMessage().Error(request.Type(), proto.E_OPERATION_FAILED)
		log.Printf("< (%v) REQUEST NEW ACCESS TOKEN ERROR: %v\n", session, err)
		session.WriteResponse(request.Header.GetToken(), reply)
		return
	}

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().NewAccessToken(accessToken.UserID(), accessToken.Token()))
	log.Printf("< (%v) ACCESS TOKEN: %v\n", session, accessToken.Token())
}

func onIIDTokenReceived(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.InstanceIDToken)
	log.Printf("> (%v) IID TOKEN %v\n", session, msg)

	checkAuthenticated(session)

	iidToken := model.NewIIDToken(msg.Token, int(session.ProtocolVersion))
	if err := server.Model.Accounts.SetPushToken(session.UserId, iidToken); err != nil {
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

	user, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().UserAccount(convUser2Net(user)))
	log.Printf("< (%v) SEND USER ACCOUNT INFO\n", session)
}

func onChangeProfilePicture(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*core.UserAccount)
	log.Printf("> (%v) CHANGE PROFILE PICTURE (%v bytes)\n", session, len(msg.Picture))

	checkAuthenticated(session)

	// Load account
	user, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	// Add or remove profile picture
	err = server.Model.Accounts.ChangeProfilePicture(user, msg.Picture)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) PROFILE PICTURE CHANGED\n", session)
	session.Write(session.NewMessage().UserAccount(convUser2Net(user)))
	log.Printf("< (%v) SEND USER ACCOUNT INFO\n", session)
}

func onSyncGroups(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.SyncGroups)
	log.Printf("> (%v) SYNC GROUPS %v\n", session, msg)

	checkAuthenticated(session)

	// Convert received groups to model.Group objects
	clientGroups := make([]*model.Group, 0, 4)
	builder := model.NewGroupBuilder()

	for _, g := range msg.Groups {
		builder.SetId(g.Id)
		builder.SetName(g.Name)
		for _, friendID := range g.Members {
			builder.AddMember(friendID)
		}
		clientGroups = append(clientGroups, builder.Build())
	}

	// Add groups
	err := server.Model.Friends.SyncGroups(session.UserId, clientGroups)
	checkNoErrorOrPanic(err)

	// Write response back
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) SYNC GROUPS OK\n", session)
}

func onGetGroups(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	log.Printf("> (%v) GET GROUPS\n", session)

	checkAuthenticated(session)

	groups, err := server.Model.Friends.GetAllGroups(session.UserId)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().GroupsList(convGroupList2Net(groups)))
	log.Printf("< (%v) GROUPS LIST (num.groups: %v)\n", session, len(groups))
}

func onCreateEvent(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateEvent)
	log.Printf("> (%v) CREATE EVENT (message: %v, start: %v, end: %v, invitations: %v, picture: %v bytes)\n",
		session, msg.Message, msg.StartDate, msg.EndDate, len(msg.Participants), len(msg.Picture))

	checkAuthenticated(session)

	author, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	event, err := server.Model.Events.CreateNewEvent(author, msg.CreatedDate, msg.StartDate, msg.EndDate, msg.Message)
	checkNoErrorOrPanic(err)

	userParticipants, err := server.Model.Events.CreateParticipantsList(author.Id(), msg.Participants)
	checkNoErrorOrPanic(err)

	if err = server.Model.Events.PublishEvent(event, userParticipants); err == nil {

		// Event Published: set event picture if received

		if len(msg.Picture) != 0 {
			if err = server.Model.Events.ChangeEventPicture(event, msg.Picture); err != nil {
				// Only log error but do nothing. Event has already been published.
				log.Printf("* (%v) Error saving picture for event %v (%v)\n", session, event.Id(), err)
			}
		}

		session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventCreated(convEvent2Net(event)))
		log.Printf("< (%v) CREATE EVENT OK (eventId: %v, Num.Participants: %v, Remaining.Participants: %v)\n",
			session, event.Id(), event.NumGuests(), len(msg.Participants)-event.NumGuests())

		notification := &NotifyEventInvitation{
			Event:  event,
			Target: userParticipants,
		}
		server.task_executor.Submit(notification)

	} else if err == model.ErrParticipantsRequired {

		// Participants required

		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), proto.E_EVENT_PARTICIPANTS_REQUIRED))
		log.Printf("< (%v) CREATE EVENT ERROR: THERE ARE NO PARTICIPANTS\n", session)

	} else {

		// Error

		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), getNetErrorCode(err, proto.E_OPERATION_FAILED)))
		log.Printf("< (%v) CREATE EVENT ERROR: %v\n", session, err)
	}
}

func onChangeEventPicture(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ModifyEvent)
	log.Printf("> (%v) CHANGE EVENT PICTURE %v (%v bytes)\n", session, msg.EventId, len(msg.Picture))

	checkAuthenticated(session)

	// Load event
	event, err := server.Model.Events.GetEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check author
	authorID := session.UserId
	checkEventAuthorOrPanic(authorID, event)

	// Change event picture
	err = server.Model.Events.ChangeEventPicture(event, msg.Picture)
	checkNoErrorOrPanic(err)

	// Send ACK to caller
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) EVENT PICTURE CHANGED\n", session)

	// Notify change to participants
	server.task_executor.Submit(&NotifyEventChange{
		Event:  event,
		Target: event.ParticipantIds(),
	})
}

func onCancelEvent(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CancelEvent)
	log.Printf("> (%v) CANCEL EVENT %v\n", session, msg)

	checkAuthenticated(session)

	event, err := server.Model.Events.GetEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check author
	authorID := session.UserId
	checkEventAuthorOrPanic(authorID, event)

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

	event, err := server.Model.Events.GetEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check author
	authorID := session.UserId
	checkEventAuthorOrPanic(authorID, event)

	// Event can be modified
	checkEventWritableOrPanic(event)

	// Load user participants
	newParticipants, err := server.Model.Events.CreateParticipantsList(authorID, msg.Participants)
	checkNoErrorOrPanic(err)

	// Get current participants for later use
	oldParticipants := event.ParticipantIds()

	// Invite participants
	usersInvited, err := server.Model.Events.InviteUsers(event, newParticipants)
	checkNoErrorOrPanic(err)

	// Write response back
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) INVITE USERS OK (event_id: %v, invitations: %v/%v, total: %v)\n",
		session, event.Id(), len(usersInvited), len(newParticipants), len(msg.Participants))

	// Notify previous participants of the new ones added
	server.task_executor.Submit(&NotifyParticipantChange{
		Event:               event,
		ParticipantsChanged: model.GetUserKeys(usersInvited),
		Target:              oldParticipants,
		NumGuests:           int32(event.NumGuests()), // Include also total NumGuests because it's changed
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

	event, err := server.Model.Events.GetEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Change response
	changed, err := server.Model.Events.ChangeParticipantResponse(session.UserId, api.AttendanceResponse(msg.ActionCode), event)
	checkNoErrorOrPanic(err)

	// Send OK Response
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) CONFIRM ATTENDANCE %v OK\n", session, msg.EventId)

	if changed {
		// Notify participants
		server.task_executor.Submit(&NotifyParticipantChange{
			Event:               event,
			ParticipantsChanged: []int64{session.UserId},
			Target:              event.ParticipantIds(),
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
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventsList(convEventList2Net(filteredEvents)))

	// Update delivery status
	for _, event := range events {

		// TODO: I should receive an ACK before try to change state.
		// Moreover, err is ignored. What should server do in that case?

		changed, _ := server.Model.Events.ChangeDeliveryState(event, session.UserId, api.InvitationStatus_CLIENT_DELIVERED)

		if changed {

			// Notify change in participant status to the other participants
			task := &NotifyParticipantChange{
				Event:               event,
				ParticipantsChanged: []int64{session.UserId},
				Target:              event.ParticipantIds(),
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

	var startWindow int64
	var endWindow int64

	if msg.StartWindow < msg.EndWindow {
		startWindow = events[0].StartDate()
		endWindow = events[len(events)-1].StartDate()
	} else {
		startWindow = events[len(events)-1].StartDate()
		endWindow = events[0].StartDate()
	}

	filteredEvents := server.Model.Events.FilterEvents(events, session.UserId)

	// Send event list to user
	log.Printf("< (%v) SEND EVENTS HISTORY (num.events: %v)", session, len(events))
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventsHistoryList(convEventList2Net(filteredEvents), startWindow, endWindow))
}

func onGetUserFriends(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server

	log.Printf("> (%v) GET USER FRIENDS\n", session) // Message does not has payload
	checkAuthenticated(session)

	friends, err := server.Model.Friends.GetAllFriends(session.UserId)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().FriendsList(convFriendList2Net(friends)))
	log.Printf("< (%v) SEND USER FRIENDS (num.friends: %v)\n", session, len(friends))
}

func onFriendRequest(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CreateFriendRequest)

	log.Printf("> (%v) CREATE FRIEND REQUEST: %v\n", session, msg)
	checkAuthenticated(session)

	// Load session.UserId account
	userAccount, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	// Exist user with provided email
	friendAccount, err := server.Model.Accounts.GetUserAccountByEmail(msg.Email)
	if err != nil {
		if err == model.ErrNotFound {
			panic(ErrFriendNotFound)
		} else {
			panic(err)
		}
	}

	// Send friend request
	friendRequest, err := server.Model.Friends.SendFriendRequest(userAccount, friendAccount)
	checkNoErrorOrPanic(err)

	server.task_executor.Submit(&NotifyFriendRequest{
		UserId:        friendAccount.Id(),
		FriendRequest: friendRequest,
	})

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) CREATE FRIEND REQUEST OK\n", session)
}

func onListFriendRequests(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server

	log.Printf("> (%v) GET FRIEND REQUESTS\n", session)
	checkAuthenticated(session)

	requests, err := server.Model.Friends.GetAllFriendRequests(session.UserId)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().FriendRequestsList(convFriendRequestList2Net(requests)))
	log.Printf("< (%v) SEND FRIEND REQUESTS (num.requests: %v)\n", session, len(requests))
}

func onConfirmFriendRequest(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ConfirmFriendRequest)

	log.Printf("> (%v) CONFIRM FRIEND REQUEST: %v\n", session, msg)
	checkAuthenticated(session)

	// Load current user
	currentUser, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	// Load friend
	friend, err := server.Model.Accounts.GetUserAccount(msg.FriendId)
	checkNoErrorOrPanic(err)

	if msg.Response == proto.ConfirmFriendRequest_CONFIRM {

		err = server.Model.Friends.ConfirmFriendRequest(friend, currentUser, true)
		checkNoErrorOrPanic(err)

		// Send OK to user
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
		log.Printf("< (%v) CONFIRM FRIEND REQUEST OK (accepted)\n", session)

		// Send Friend List to both users if connected
		// TODO: In this case, FRIENDS_LIST is sent as a notification. Because of this,
		// it should be only a subset of all friends in server.
		server.task_executor.Submit(&SendUserFriends{UserId: currentUser.Id()})
		server.task_executor.Submit(&SendUserFriends{UserId: friend.Id()})
		//task.sendGcmNotification(friendUser.Id, friendUser.IIDtoken, task.TargetUser)

	} else if msg.Response == proto.ConfirmFriendRequest_CANCEL {

		err = server.Model.Friends.ConfirmFriendRequest(friend, currentUser, false)
		checkNoErrorOrPanic(err)

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
