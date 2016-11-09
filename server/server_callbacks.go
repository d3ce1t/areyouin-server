package main

import (
	"log"
	"peeple/areyouin/api"
	fb "peeple/areyouin/facebook"
	"peeple/areyouin/model"
	proto "peeple/areyouin/protocol"
	"peeple/areyouin/protocol/core"
	"peeple/areyouin/utils"
	"time"
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

		// Import Facebook friends that already have AreYouIN

		go func() {
			addedFriends, err := server.Model.Friends.ImportFacebookFriends(userAccount, true)
			if err != nil {
				log.Printf("* (%v) IMPORT FACEBOOK FRIENDS ERROR: %v", session, err)
				return
			}
			log.Printf("* (%v) IMPORT FACEBOOK FRIENDS SUCCESS (added: %v)", session, len(addedFriends))
		}()
	}

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().UserAccessGranted(userAccount.Id(), userAccount.AuthToken()))
	log.Printf("< (%v) CREATE ACCOUNT OK\n", session)
}

func onLinkAccount(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.LinkAccount)
	log.Printf("> (%v) LINK ACCOUNT: %v\n", session, msg)

	checkAuthenticated(session)

	user, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	err = server.Model.Accounts.LinkToFacebook(user, msg.AccountId, msg.AccountToken)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) LINK ACCOUNT OK\n", session)

	session.Write(session.NewMessage().UserAccount(convUser2Net(user)))
	log.Printf("< (%v) SEND USER ACCOUNT INFO\n", session)
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
			reply = session.NewMessage().Error(request.Type(), proto.E_FB_INVALID_ACCESS_TOKEN)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID FB ACCESS: %v\n", session, fb.GetErrorMessage(err))
		} else if err == model.ErrInvalidUserOrPassword {
			reply = session.NewMessage().Error(request.Type(), proto.E_INVALID_USER_OR_PASSWORD)
			log.Printf("< (%v) USER NEW AUTH TOKEN INVALID USER OR PASSWORD", session)
		} else {
			reply = session.NewMessage().Error(request.Type(), proto.E_OPERATION_FAILED)
			log.Printf("< (%v) USER NEW AUTH TOKEN ERROR: %v\n", session, err)
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

	iidToken := model.NewIIDToken(msg.Token, int(session.ProtocolVersion), session.Platform)
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
		if g.Size == -1 && len(g.Members) == 0 {
			// Special case
			if g.Name == "" {
				// Group is marked for removal. So remove it from server
				server.Model.Friends.DeleteGroup(session.UserId, g.Id)
			} else {
				// Only Rename group
				server.Model.Friends.RenameGroup(session.UserId, g.Id, g.Name)
			}
		} else {
			// Update case
			builder.SetId(g.Id)
			builder.SetName(g.Name)
			for _, friendID := range g.Members {
				builder.AddMember(friendID)
			}
			clientGroups = append(clientGroups, builder.Build())
		}
	}

	// Add groups
	if len(clientGroups) > 0 {
		err := server.Model.Friends.AddGroups(session.UserId, clientGroups)
		checkNoErrorOrPanic(err)
	}

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
	createdDate := utils.MillisToTimeUTC(msg.CreatedDate)
	startDate := utils.MillisToTimeUTC(msg.StartDate)
	endDate := utils.MillisToTimeUTC(msg.EndDate)

	log.Printf("> (%v) CREATE EVENT (start: %v, end: %v, invitations: %v, picture: %v bytes)\n",
		session, startDate, endDate, len(msg.Participants), len(msg.Picture))

	checkAuthenticated(session)

	// Get author
	author, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	// New event
	event, err := server.Model.Events.NewEvent(author, createdDate, startDate, endDate, msg.Message, msg.Participants)
	checkNoErrorOrPanic(err)

	// Publish event
	err = server.Model.Events.SaveEvent(event)
	checkNoErrorOrPanic(err)

	// Event Published: set event picture if received

	if len(msg.Picture) != 0 {
		if err = server.Model.Events.ChangeEventPicture(event, msg.Picture); err != nil {
			// Only log error but do nothing. Event has already been published.
			log.Printf("* (%v) Error saving picture for event %v (%v)\n", session, event.Id(), err)
		}
	}

	// Send event with InvitationStatus_CLIENT_DELIVERED
	coreEvent := convEvent2Net(event)
	coreEvent.Participants[session.UserId].Delivered = core.InvitationStatus_CLIENT_DELIVERED
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventCreated(coreEvent))
	log.Printf("< (%v) CREATE EVENT OK (eventId: %v, Num.Participants: %v, Remaining.Participants: %v)\n",
		session, event.Id(), event.NumGuests(), 1+len(msg.Participants)-event.NumGuests())

	// Change invitation status
	_, err = server.Model.Events.ChangeDeliveryState(session.UserId, api.InvitationStatus_CLIENT_DELIVERED, event)
	if err != nil {
		log.Printf("* (%v) CREATE EVENT WARNING Changing delivery state: %err", err)
	}

}

// Modify existing event. If a field isn't set that means it isn't modified.
func onModifyEvent(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server

	msg := message.(*proto.ModifyEvent)
	modificationDate := utils.MillisToTimeUTC(msg.ModifyDate)
	startDate := utils.MillisToTimeUTC(msg.StartDate)
	endDate := utils.MillisToTimeUTC(msg.EndDate)
	log.Printf("> (%v) MODIFY EVENT %v (message: %v, start: %v, end: %v, modify: %v, invitations: %v, picture: %v, remove: %v)\n",
		session, msg.EventId, msg.Message != "", startDate, endDate, modificationDate,
		len(msg.Participants), len(msg.Picture) > 0, msg.RemovePicture)

	checkAuthenticated(session)

	// Load event
	event, err := server.Model.Events.LoadEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check author
	authorID := session.UserId
	checkAccessOrPanic(authorID, event)

	// Modify event
	b := server.Model.Events.NewEventModifier(event, authorID)
	b.SetModifiedDate(modificationDate)
	eventInfoChanged := false

	if msg.Message != "" && msg.Message != event.Description() {
		b.SetDescription(msg.Message)
		eventInfoChanged = true
	}

	if !startDate.IsZero() && !startDate.Equal(event.StartDate()) {
		b.SetStartDate(startDate)
		eventInfoChanged = true
	}

	if !endDate.IsZero() && !endDate.Equal(event.EndDate()) {
		b.SetEndDate(endDate)
		eventInfoChanged = true
	}

	for _, pID := range msg.Participants {
		b.ParticipantAdder().AddUserID(pID)
	}

	modifiedEvent, err := b.Build()
	checkNoErrorOrPanic(err)

	// Persist event
	err = server.Model.Events.SaveEvent(modifiedEvent)
	checkNoErrorOrPanic(err)

	if len(msg.Picture) > 0 || msg.RemovePicture {

		// Set Event Picture

		if err := server.Model.Events.ChangeEventPicture(modifiedEvent, msg.Picture); err != nil {
			log.Printf("* (%v) Error saving picture for event %v (%v)\n", session, modifiedEvent.Id(), err)
		} else {
			eventInfoChanged = true
		}
	}

	// Send ACK to caller
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) MODIFY EVENT OK (eventId: %v)\n", session, event.Id())

	// Send event changed
	if eventInfoChanged {
		// NOTE: Send liteEvent because full event causes a crash in iOS version lower than 1.0.11
		// FIXME: If two users modify event at the same time, each one will receive a different view of the event
		netEvent := convEvent2Net(modifiedEvent.CloneWithEmptyParticipants())
		session.Write(session.NewMessage().EventModified(netEvent))
		log.Printf("< (%v) EVENT %v CHANGED\n", session.UserId, modifiedEvent.Id())
	}

	// Send participants by other means
	newParticipants := server.Model.Events.ExtractNewParticipants(modifiedEvent, event)

	if len(newParticipants) > 0 {
		netParticipants := convParticipantList2Net(newParticipants)
		packet := session.NewMessage().AttendanceStatusWithNumGuests(event.Id(), netParticipants, modifiedEvent.NumGuests())
		session.Write(packet)
		log.Printf("< (%v) EVENT %v ATTENDANCE STATUS CHANGED (%v participants changed)\n",
			session.UserId, modifiedEvent.Id(), len(newParticipants))
	}
}

func onChangeEventPicture(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.ModifyEvent)
	log.Printf("> (%v) CHANGE EVENT PICTURE %v (%v bytes)\n", session, msg.EventId, len(msg.Picture))

	checkAuthenticated(session)

	// Load event
	event, err := server.Model.Events.LoadEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check author
	authorID := session.UserId
	checkAccessOrPanic(authorID, event)

	// Change event picture
	err = server.Model.Events.ChangeEventPicture(event, msg.Picture)
	checkNoErrorOrPanic(err)

	// Send ACK to caller
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) EVENT PICTURE CHANGED\n", session)

	// Send event changed
	liteEvent := event.CloneWithEmptyParticipants()
	session.Write(session.NewMessage().EventModified(convEvent2Net(liteEvent)))
	log.Printf("< (%v) EVENT %v CHANGED\n", session.UserId, event.Id())
}

func onCancelEvent(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.CancelEvent)
	log.Printf("> (%v) CANCEL EVENT %v\n", session, msg)

	checkAuthenticated(session)

	event, err := server.Model.Events.LoadEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check author
	// TODO: Should user permissions be part of server or the model?
	authorID := session.UserId
	checkAccessOrPanic(authorID, event)

	// Cancel event
	cancelledEvent, err := server.Model.Events.NewEventModifier(event, authorID).SetCancelled(true).Build()
	checkNoErrorOrPanic(err)

	// Persist event
	err = server.Model.Events.SaveEvent(cancelledEvent)
	checkNoErrorOrPanic(err)

	// FIXME: Could send directly the event canceled message, and ignore author from
	// NotifyEventCancelled task
	log.Printf("< (%v) CANCEL EVENT OK\n", session)
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))

	// Send event changed
	liteEvent := cancelledEvent.CloneWithEmptyParticipants()
	session.Write(session.NewMessage().EventCancelled(session.UserId, convEvent2Net(liteEvent)))
	log.Printf("< (%v) EVENT %v CHANGED\n", session.UserId, cancelledEvent.Id())
}

func onInviteUsers(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.InviteUsers)
	log.Printf("> (%v) INVITE USERS %v\n", session, msg)

	checkAuthenticated(session)

	// Fail early
	if len(msg.Participants) == 0 {
		session.WriteResponse(request.Header.GetToken(), session.NewMessage().Error(request.Type(), proto.E_EVENT_PARTICIPANTS_REQUIRED))
		log.Printf("< (%v) INVITE USERS ERROR (event_id=%v) PARTICIPANTS REQUIRED\n", session, msg.EventId)
		return
	}

	// Read event
	event, err := server.Model.Events.LoadEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Check author
	checkAccessOrPanic(session.UserId, event)

	// Build event
	b := server.Model.Events.NewEventModifier(event, session.UserId)
	for _, pID := range msg.Participants {
		b.ParticipantAdder().AddUserID(pID)
	}
	modifiedEvent, err := b.Build()
	checkNoErrorOrPanic(err)

	// Save it
	err = server.Model.Events.SaveEvent(modifiedEvent)
	checkNoErrorOrPanic(err)

	newParticipants := server.Model.Events.ExtractNewParticipants(modifiedEvent, event)

	// Write response back
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) INVITE USERS OK (eventID: %v, newParticipants: %v/%v, total: %v)\n",
		session, event.Id(), len(newParticipants), len(msg.Participants), modifiedEvent.NumGuests())

	// Send new participants
	netParticipants := convParticipantList2Net(newParticipants)
	packet := session.NewMessage().AttendanceStatusWithNumGuests(event.Id(), netParticipants, modifiedEvent.NumGuests())
	session.Write(packet)
	log.Printf("< (%v) EVENT %v ATTENDANCE STATUS CHANGED (%v participants changed)\n", session.UserId, modifiedEvent.Id(), len(netParticipants))
}

// When a ConfirmAttendance message is received, the attendance response of the participant
// in the participant event list is changed and notified to the other participants. It is
// important to note that num_attendees is not changed server-side till the event has started.
// Clients are cool counting attendees :)
func onConfirmAttendance(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	msg := message.(*proto.ConfirmAttendance)
	log.Printf("> (%v) CONFIRM ATTENDANCE %v\n", session, msg)

	checkAuthenticated(session)

	server := session.Server
	event, err := server.Model.Events.LoadEvent(msg.EventId)
	checkNoErrorOrPanic(err)

	// Change response
	participant, err := server.Model.Events.ChangeParticipantResponse(session.UserId,
		api.AttendanceResponse(msg.ActionCode), event)
	checkNoErrorOrPanic(err)

	// Send OK Response
	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) CONFIRM ATTENDANCE %v OK\n", session, msg.EventId)

	// Send new AttendanceStatus
	participantList := make(map[int64]*model.Participant)
	participantList[participant.Id()] = participant
	netParticipants := convParticipantList2Net(participantList)
	session.Write(session.NewMessage().AttendanceStatus(event.Id(), netParticipants))
	log.Printf("< (%v) EVENT %v ATTENDANCE STATUS CHANGED (%v participants changed)\n", session.UserId, event.Id(), len(netParticipants))
}

func onReadEvent(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	msg := message.(*proto.ReadEvent)
	log.Printf("> (%v) READ EVENT %v\n", session, msg.EventId)

	checkAuthenticated(session)

	server := session.Server

	// Load recent events
	event, err := server.Model.Events.GetEventForUser(session.UserId, msg.EventId)
	checkNoErrorOrPanic(err)

	netEvent := convEvent2Net(event)

	// Delivery status
	if event.Status() == api.EventState_NOT_STARTED {

		if participant, ok := netEvent.Participants[session.UserId]; ok {
			participant.Delivered = core.InvitationStatus_CLIENT_DELIVERED
		}
	}

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Event(netEvent))
	log.Printf("< (%v) SEND EVENT %v\n", session.UserId, event.Id())

	// WORKAROUND: Read EventManager
	if event.IsCancelled() {
		server.Model.Events.RemoveFromInbox(session.UserId, event.Id())
	}

	if event.Status() == api.EventState_NOT_STARTED {
		// Update delivery status
		// TODO: I should receive an ACK before try to change state.
		if participant, _ := event.Participants.Get(session.UserId); participant != nil {
			if participant.InvitationStatus() != api.InvitationStatus_CLIENT_DELIVERED {
				_, err := server.Model.Events.ChangeDeliveryState(session.UserId, api.InvitationStatus_CLIENT_DELIVERED, event)
				if err != nil {
					log.Printf("* (%v) READ EVENT UPDATE DELIVERY STATUS ERROR (eventID: %v): %v)", session, event.Id(), err)
				}
			}
		}
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
	log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session, len(events))

	// Delivered to own participant
	eventList := convEventList2Net(events)
	for _, event := range eventList {
		if participant, ok := event.Participants[session.UserId]; ok {
			participant.Delivered = core.InvitationStatus_CLIENT_DELIVERED
		}
	}

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().EventsList(eventList))

	// Update delivery status
	for _, event := range events {

		// WORKAROUND: Read EventManager
		if event.IsCancelled() {
			server.Model.Events.RemoveFromInbox(session.UserId, event.Id())
		}

		// TODO: I should receive an ACK before try to change state.
		if participant, _ := event.Participants.Get(session.UserId); participant != nil {
			if participant.InvitationStatus() != api.InvitationStatus_CLIENT_DELIVERED {
				_, err := server.Model.Events.ChangeDeliveryState(session.UserId, api.InvitationStatus_CLIENT_DELIVERED, event)
				if err != nil {
					log.Printf("* (%v) SEND PRIVATE EVENTS UPDATE DELIVERY STATUS ERROR (eventID: %v): %v)", session, event.Id(), err)
				}
			}
		}
	}
}

func onListEventsHistory(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*proto.EventListRequest)

	log.Printf("> (%v) REQUEST EVENTS HISTORY (start: %v, end: %v)\n",
		session, utils.MillisToTimeUTC(msg.StartWindow), utils.MillisToTimeUTC(msg.EndWindow)) // Message does not has payload

	checkAuthenticated(session)

	reqStartWindow := utils.MillisToTimeUTC(msg.StartWindow)
	reqEndWindow := utils.MillisToTimeUTC(msg.EndWindow)
	events, err := server.Model.Events.GetEventsHistory(session.UserId, reqStartWindow, reqEndWindow)
	checkNoErrorOrPanic(err)

	var firstEvent, lastEvent *model.Event
	var startWindow, endWindow time.Time

	if msg.StartWindow < msg.EndWindow {
		firstEvent, lastEvent = events[0], events[len(events)-1]
	} else {
		firstEvent, lastEvent = events[len(events)-1], events[0]
	}

	if firstEvent.IsCancelled() {
		startWindow = firstEvent.InboxPosition()
	} else {
		startWindow = firstEvent.EndDate()
	}

	if lastEvent.IsCancelled() {
		endWindow = lastEvent.InboxPosition()
	} else {
		endWindow = lastEvent.EndDate()
	}

	// Send event list to user
	log.Printf("< (%v) SEND EVENTS HISTORY (num.events: %v, startWindow: %v, endWindow: %v)",
		session, len(events), startWindow, endWindow)

	session.WriteResponse(request.Header.GetToken(),
		session.NewMessage().EventsHistoryList(convEventList2Net(events),
			utils.TimeToMillis(startWindow), utils.TimeToMillis(endWindow)))
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

// Returns Facebook Friends that are AreYouIN registered users but they aren't in user's friends list
func onGetFacebookFriends(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server

	log.Printf("> (%v) GET FACEBOOK FRIENDS\n", session) // Message does not has payload
	checkAuthenticated(session)

	// Get account
	account, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	// Get Facebook Friends that have AreYouIN but are not friends of session.UserID in AreYouIN
	newFBFriends, err := server.Model.Friends.GetNewFacebookFriends(account)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(),
		session.NewMessage().FacebookFriendsList(convUserList2FriendNet(newFBFriends)))
	log.Printf("< (%v) SEND FACEBOOK FRIENDS (num.friends: %v)\n", session, len(newFBFriends))
}

func onImportFacebookFriends(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server

	log.Printf("> (%v) IMPORT FACEBOOK FRIENDS\n", session) // Message does not has payload
	checkAuthenticated(session)

	// Get account
	account, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	// Import Facebook friends
	addedFriends, err := server.Model.Friends.ImportFacebookFriends(account, false)
	checkNoErrorOrPanic(err)

	// Send list of imported users
	session.WriteResponse(request.Header.GetToken(),
		session.NewMessage().FacebookFriendsList(convUserList2FriendNet(addedFriends)))
	log.Printf("< (%v) IMPORT FACEBOOK FRIENDS OK (added: %v)\n", session, len(addedFriends))
}

func onSetFacebookAccessToken(request *proto.AyiPacket, message proto.Message, session *AyiSession) {

	server := session.Server
	msg := message.(*core.FacebookAccessToken)

	log.Printf("> (%v) SET FACEBOOK ACCESS TOKEN (%v)\n", session, msg)
	checkAuthenticated(session)

	// Get account
	account, err := server.Model.Accounts.GetUserAccount(session.UserId)
	checkNoErrorOrPanic(err)

	// Update Facebook token
	err = server.Model.Accounts.SetFacebookAccessToken(account, msg.AccessToken)
	checkNoErrorOrPanic(err)

	session.WriteResponse(request.Header.GetToken(), session.NewMessage().Ok(request.Type()))
	log.Printf("< (%v) SET FACEBOOK ACCESS TOKEN OK\n", session)
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
	_, err = server.Model.Friends.CreateFriendRequest(userAccount, friendAccount)
	checkNoErrorOrPanic(err)

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
