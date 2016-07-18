package protocol

import (
	core "peeple/areyouin/common"
)

type PacketBuilder struct {
	message *AyiPacket
}

// Modifiers
/*func (mb *PacketBuilder) CreateEvent(message string, start_date int64, end_date int64, participants []int64) *AyiPacket {
	mb.message.Header.SetType(M_CREATE_EVENT)
	mb.message.SetMessage(&CreateEvent{Message: message, StartDate: start_date, EndDate: end_date, Participants: participants})
	return mb.message
}

func (mb *PacketBuilder) CancelEvent(event_id int64, reason string) *AyiPacket {
	mb.message.Header.SetType(M_CANCEL_EVENT)
	mb.message.SetMessage(&CancelEvent{EventId: event_id, Reason: reason})
	return mb.message
}

func (mb *PacketBuilder) InviteUsers(event_id int64, participants []int64) *AyiPacket {
	mb.message.Header.SetType(M_INVITE_USERS)
	mb.message.SetMessage(&InviteUsers{EventId: event_id, Participants: participants})
	return mb.message
}

func (mb *PacketBuilder) CancelUsersInvitation(event_id int64, participants []int64) *AyiPacket {
	mb.message.Header.SetType(M_CANCEL_USERS_INVITATION)
	mb.message.SetMessage(&CancelUsersInvitation{EventId: event_id, Participants: participants})
	return mb.message
}

func (mb *PacketBuilder) ConfirmAttendance(event_id int64, action_code core.AttendanceResponse) *AyiPacket {
	mb.message.Header.SetType(M_CONFIRM_ATTENDANCE)
	mb.message.SetMessage(&ConfirmAttendance{EventId: event_id, ActionCode: action_code})
	return mb.message
}

func (mb *PacketBuilder) ModifyEventDate(event_id int64, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.SetType(M_MODIFY_EVENT_DATE)
	mb.message.SetMessage(&ModifyEvent{EventId: event_id, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *PacketBuilder) ModifyEventMessage(event_id int64, message string) *AyiPacket {
	mb.message.Header.SetType(M_MODIFY_EVENT_MESSAGE)
	mb.message.SetMessage(&ModifyEvent{EventId: event_id, Message: message})
	return mb.message
}

func (mb *PacketBuilder) ModifyEvent(event_id int64, message string, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.SetType(M_MODIFY_EVENT)
	mb.message.SetMessage(&ModifyEvent{EventId: event_id, Message: message, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *PacketBuilder) VoteChange(event_id int64, change_id int32, accept_change bool) *AyiPacket {
	mb.message.Header.SetType(M_VOTE_CHANGE)
	mb.message.SetMessage(&VoteChange{EventId: event_id, ChangeId: change_id, AcceptChange: accept_change})
	return mb.message
}

func (mb *PacketBuilder) UserPosition(latitude float32, longitude float32, estimation_error float32) *AyiPacket {
	mb.message.Header.SetType(M_USER_POSITION)
	mb.message.SetMessage(&UserPosition{GlobalCoordinates: &core.Location{Latitude: latitude, Longitude: longitude}, EstimationError: estimation_error})
	return mb.message
}

func (mb *PacketBuilder) UserPositionRange(range_in_meters float32) *AyiPacket {
	mb.message.Header.SetType(M_USER_POSITION_RANGE)
	mb.message.SetMessage(&UserPositionRange{RangeInMeters: range_in_meters})
	return mb.message
}

func (mb *PacketBuilder) CreateUserAccount(name string, email string, password string, phone string, fbid string, fbtoken string) *AyiPacket {
	mb.message.Header.SetType(M_USER_CREATE_ACCOUNT)
	mb.message.SetMessage(&CreateUserAccount{Name: name, Email: email, Password: password, Phone: phone, Fbid: fbid, Fbtoken: fbtoken})
	return mb.message
}

func (mb *PacketBuilder) NewAuthTokenByEmail(email string, password string) *AyiPacket {
	mb.message.Header.SetType(M_USER_NEW_AUTH_TOKEN)
	mb.message.SetMessage(&NewAuthToken{Pass1: email, Pass2: password, Type: AuthType_A_NATIVE})
	return mb.message
}

func (mb *PacketBuilder) NewAuthTokenByFacebook(fbid string, fbtoken string) *AyiPacket {
	mb.message.Header.SetType(M_USER_NEW_AUTH_TOKEN)
	mb.message.SetMessage(&NewAuthToken{Pass1: fbid, Pass2: fbtoken, Type: AuthType_A_FACEBOOK})
	return mb.message
}

func (mb *PacketBuilder) UserAuthencation(user_id int64, auth_token uuid.UUID) *AyiPacket {
	mb.message.Header.SetType(M_USER_AUTH)
	mb.message.SetMessage(&AccessToken{UserId: user_id, AuthToken: auth_token.String()})
	return mb.message
}*/

func (mb *PacketBuilder) NewAccessToken(user_id int64, auth_token string) *AyiPacket {
	mb.message.Header.SetType(M_ACCESS_TOKEN)
	mb.message.SetMessage(&AccessToken{UserId: user_id, AuthToken: auth_token})
	return mb.message
}

/*func (mb *PacketBuilder) IIDToken(token string) *AyiPacket {
	mb.message.Header.SetType(M_IID_TOKEN)
	mb.message.SetMessage(&InstanceIDToken{Token: token})
	return mb.message
}

func (mb *PacketBuilder) ChangeProfilePicture(picture []byte) *AyiPacket {
	mb.message.Header.SetType(M_CHANGE_PROFILE_PICTURE)
	mb.message.SetMessage(&UserAccount{Picture: picture})
	return mb.message
}*/

// Notifications
func (mb *PacketBuilder) EventCreated(event *core.Event) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_CREATED)
	mb.message.SetMessage(event)
	return mb.message
}

func (mb *PacketBuilder) EventCancelled(who_id int64, event *core.Event) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_CANCELLED)
	mb.message.SetMessage(&EventCancelled{WhoId: who_id, EventId: event.EventId, Event: event})
	return mb.message
}

func (mb *PacketBuilder) EventExpired(event_id int64) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_EXPIRED)
	mb.message.SetMessage(&EventExpired{EventId: event_id})
	return mb.message
}

/*func (mb *PacketBuilder) EventDateModified(event_id int64, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_DATE_MODIFIED)
	mb.message.SetMessage(&EventModified{EventId: event_id, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *PacketBuilder) EventMessageModified(event_id int64, message string) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_MESSAGE_MODIFIED)
	mb.message.SetMessage(&EventModified{EventId: event_id, Message: message})
	return mb.message
}*/

func (mb *PacketBuilder) EventModified(event *core.Event) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_MODIFIED)
	mb.message.SetMessage(event)
	return mb.message
}

func (mb *PacketBuilder) InvitationReceived(event *core.Event) *AyiPacket {
	mb.message.Header.SetType(M_INVITATION_RECEIVED)
	mb.message.SetMessage(event)
	return mb.message
}

/*func (mb *PacketBuilder) InvitationCancelled(event_id int64) *AyiPacket {
	mb.message.Header.SetType(M_INVITATION_CANCELLED)
	mb.message.SetMessage(&InvitationCancelled{EventId: event_id})
	return mb.message
}*/

func (mb *PacketBuilder) AttendanceStatus(event_id int64, status []*core.EventParticipant) *AyiPacket {
	mb.message.Header.SetType(M_ATTENDANCE_STATUS)
	mb.message.SetMessage(&AttendanceStatus{EventId: event_id, AttendanceStatus: status})
	return mb.message
}

func (mb *PacketBuilder) AttendanceStatusWithNumGuests(event_id int64, status []*core.EventParticipant, num_guests int32) *AyiPacket {
	mb.message.Header.SetType(M_ATTENDANCE_STATUS)
	mb.message.SetMessage(&AttendanceStatus{EventId: event_id, AttendanceStatus: status, NumGuests: num_guests})
	return mb.message
}

/*func (mb *PacketBuilder) EventChangeDateProposed(event_id int64, change_id int32, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_CHANGE_DATE_PROPOSED)
	mb.message.SetMessage(&EventChangeProposed{EventId: event_id, ChangeId: change_id, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *PacketBuilder) EventChangeMessageProposed(event_id int64, change_id int32, message string) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_CHANGE_MESSAGE_PROPOSED)
	mb.message.SetMessage(&EventChangeProposed{EventId: event_id, ChangeId: change_id, Message: message})
	return mb.message
}

func (mb *PacketBuilder) EventChangeProposed(event_id int64, change_id int32, message string, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_CHANGE_PROPOSED)
	mb.message.SetMessage(&EventChangeProposed{EventId: event_id, ChangeId: change_id, Message: message,
		StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *PacketBuilder) VotingStatus(event_id int64, change_id int32, start_date int64,
	end_date int64, votes_received uint32, votes_total uint32) *AyiPacket {
	mb.message.Header.SetType(M_VOTING_STATUS)
	mb.message.SetMessage(&VotingStatus{EventId: event_id, ChangeId: change_id,
		StartDate: start_date, EndDate: end_date, ElapsedTime: end_date - start_date,
		VotesReceived: votes_received, VotesTotal: votes_total})
	return mb.message
}

func (mb *PacketBuilder) VotingFinished(event_id int64, change_id int32) *AyiPacket {
	mb.message.Header.SetType(M_VOTING_FINISHED)
	mb.message.SetMessage(&VotingStatus{EventId: event_id, ChangeId: change_id, Finished: true})
	return mb.message
}

func (mb *PacketBuilder) ChangeAccepted(event_id int64, change_id int32) *AyiPacket {
	mb.message.Header.SetType(M_CHANGE_ACCEPTED)
	mb.message.SetMessage(&ChangeAccepted{EventId: event_id, ChangeId: change_id})
	return mb.message
}

func (mb *PacketBuilder) ChangeDiscarded(event_id int64, change_id int32) *AyiPacket {
	mb.message.Header.SetType(M_CHANGE_DISCARDED)
	mb.message.SetMessage(&ChangeDiscarded{EventId: event_id, ChangeId: change_id})
	return mb.message
}*/

func (mb *PacketBuilder) UserAccessGranted(user_id int64, auth_token string) *AyiPacket {
	mb.message.Header.SetType(M_ACCESS_GRANTED)
	mb.message.SetMessage(&AccessToken{UserId: user_id, AuthToken: auth_token})
	return mb.message
}

func (mb *PacketBuilder) Ok(msg_type PacketType) *AyiPacket {
	mb.message.Header.SetType(M_OK)
	mb.message.SetMessage(&Ok{Type: uint32(msg_type)})
	return mb.message
}

func (mb *PacketBuilder) Error(msg_type PacketType, error_code int32) *AyiPacket {
	mb.message.Header.SetType(M_ERROR)
	mb.message.SetMessage(&Error{Type: uint32(msg_type), Error: error_code})
	return mb.message
}

// Requests
func (mb *PacketBuilder) Ping() *AyiPacket {
	mb.message.Header.SetType(M_PING)
	mb.message.SetMessage(&TimeInfo{core.GetCurrentTimeMillis()})
	return mb.message
}

/*func (mb *PacketBuilder) ReadEvent(event_id int64) *AyiPacket {
	mb.message.Header.SetType(M_READ_EVENT)
	mb.message.SetMessage(&ReadEvent{EventId: event_id})
	return mb.message
}*/

/*func (mb *PacketBuilder) ListAuthoredEvents(cursor uint32) *AyiPacket {
	mb.message.Header.SetType(M_LIST_AUTHORED_EVENTS)
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}*/

/*func (mb *PacketBuilder) ListPrivateEvents(cursor uint32) *AyiPacket {
	mb.message.Header.SetType(M_LIST_PRIVATE_EVENTS)
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}*/

/*func (mb *PacketBuilder) ListPublicEvents(latitude float32, longitude float32, range_in_meters uint32, cursor uint32) *AyiPacket {
	mb.message.Header.SetType(M_LIST_PUBLIC_EVENTS)
	mb.message.SetMessage(&ListPublicEvents{UserCoordinates: &core.Location{Latitude: latitude, Longitude: longitude},
		RangeInMeters: range_in_meters, Cursor: &ListCursor{cursor}})
	return mb.message
}*/

/*func (mb *PacketBuilder) HistoryAuthoredEvents(cursor uint32) *AyiPacket {
	mb.message.Header.SetType(M_HISTORY_AUTHORED_EVENTS)
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}*/

/*func (mb *PacketBuilder) HistoryPrivateEvents(cursor uint32) *AyiPacket {
	mb.message.Header.SetType(M_HISTORY_PRIVATE_EVENTS)
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}*/

/*func (mb *PacketBuilder) HistoryPublicEvents(cursor uint32) *AyiPacket {
	mb.message.Header.SetType(M_HISTORY_PUBLIC_EVENTS)
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}*/

/*func (mb *PacketBuilder) UserFriends() *AyiPacket {
	mb.message.Header.SetType(M_GET_USER_FRIENDS)
	return mb.message
}*/

// Responses
func (mb *PacketBuilder) Pong() *AyiPacket {
	mb.message.Header.SetType(M_PONG)
	mb.message.SetMessage(&TimeInfo{core.GetCurrentTimeMillis()})
	return mb.message
}

/* Create an Event message with all its information, including participants information */
func (mb *PacketBuilder) EventInfo(event *core.Event) *AyiPacket {
	mb.message.Header.SetType(M_EVENT_INFO)
	mb.message.SetMessage(event)
	return mb.message
}

/* Create a List of events */
func (mb *PacketBuilder) EventsList(events_list []*core.Event) *AyiPacket {
	mb.message.Header.SetType(M_EVENTS_LIST)
	mb.message.SetMessage(&EventsList{Event: events_list})
	return mb.message
}

func (mb *PacketBuilder) EventsHistoryList(events_list []*core.Event, startWindow int64, endWindow int64) *AyiPacket {
	mb.message.Header.SetType(M_EVENTS_HISTORY_LIST)
	mb.message.SetMessage(&EventsList{Event: events_list, StartWindow: startWindow, EndWindow: endWindow})
	return mb.message
}

func (mb *PacketBuilder) FriendsList(friends_list []*core.Friend) *AyiPacket {
	mb.message.Header.SetType(M_FRIENDS_LIST)
	mb.message.SetMessage(&FriendsList{Friends: friends_list})
	return mb.message
}

func (mb *PacketBuilder) ClockResponse() *AyiPacket {
	mb.message.Header.SetType(M_CLOCK_RESPONSE)
	mb.message.SetMessage(&TimeInfo{core.GetCurrentTimeMillis()})
	return mb.message
}

func (mb *PacketBuilder) UserAccount(user *core.UserAccount) *AyiPacket {
	mb.message.Header.SetType(M_USER_ACCOUNT)
	mb.message.SetMessage(&UserAccount{
		Name:          user.Name,
		Email:         user.Email,
		PictureDigest: user.PictureDigest,
		FbId:					 user.Fbid})
	return mb.message
}

func (mb *PacketBuilder) GroupsList(groups_list []*core.Group) *AyiPacket {
	mb.message.Header.SetType(M_GROUPS_LIST)
	mb.message.SetMessage(&GroupsList{Groups: groups_list})
	return mb.message
}

func (mb *PacketBuilder) FriendRequestReceived(request *core.FriendRequest) *AyiPacket {
	mb.message.Header.SetType(M_FRIEND_REQUEST_RECEIVED)
	mb.message.SetMessage(request)
	return mb.message
}

func (mb *PacketBuilder) FriendRequestsList(requests_list []*core.FriendRequest) *AyiPacket {
	mb.message.Header.SetType(M_FRIEND_REQUESTS_LIST)
	mb.message.SetMessage(&FriendRequestsList{FriendRequests: requests_list})
	return mb.message
}
