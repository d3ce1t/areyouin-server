package protocol

import (
	"github.com/twinj/uuid"
	"time"
)

type MessageBuilder struct {
	message *AyiPacket
}

// Modifiers
func (mb *MessageBuilder) CreateEvent(message string, start_date int64, end_date int64, participants []uint64) *AyiPacket {
	mb.message.Header.Type = M_CREATE_EVENT
	mb.message.SetMessage(&CreateEvent{Message: message, StartDate: start_date, EndDate: end_date, Participants: participants})
	return mb.message
}

func (mb *MessageBuilder) CancelEvent(event_id uint64, reason string) *AyiPacket {
	mb.message.Header.Type = M_CANCEL_EVENT
	mb.message.SetMessage(&CancelEvent{EventId: event_id, Reason: reason})
	return mb.message
}

func (mb *MessageBuilder) InviteUsers(event_id uint64, participants []uint64) *AyiPacket {
	mb.message.Header.Type = M_INVITE_USERS
	mb.message.SetMessage(&InviteUsers{EventId: event_id, Participants: participants})
	return mb.message
}

func (mb *MessageBuilder) CancelUsersInvitation(event_id uint64, participants []uint64) *AyiPacket {
	mb.message.Header.Type = M_CANCEL_USERS_INVITATION
	mb.message.SetMessage(&CancelUsersInvitation{EventId: event_id, Participants: participants})
	return mb.message
}

func (mb *MessageBuilder) ConfirmAttendance(event_id uint64, action_code AttendanceResponse) *AyiPacket {
	mb.message.Header.Type = M_CONFIRM_ATTENDANCE
	mb.message.SetMessage(&ConfirmAttendance{EventId: event_id, ActionCode: action_code})
	return mb.message
}

func (mb *MessageBuilder) ModifyEventDate(event_id uint64, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.Type = M_MODIFY_EVENT_DATE
	mb.message.SetMessage(&ModifyEvent{EventId: event_id, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *MessageBuilder) ModifyEventMessage(event_id uint64, message string) *AyiPacket {
	mb.message.Header.Type = M_MODIFY_EVENT_MESSAGE
	mb.message.SetMessage(&ModifyEvent{EventId: event_id, Message: message})
	return mb.message
}

func (mb *MessageBuilder) ModifyEvent(event_id uint64, message string, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.Type = M_MODIFY_EVENT
	mb.message.SetMessage(&ModifyEvent{EventId: event_id, Message: message, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *MessageBuilder) VoteChange(event_id uint64, change_id uint32, accept_change bool) *AyiPacket {
	mb.message.Header.Type = M_VOTE_CHANGE
	mb.message.SetMessage(&VoteChange{EventId: event_id, ChangeId: change_id, AcceptChange: accept_change})
	return mb.message
}

func (mb *MessageBuilder) UserPosition(latitude float32, longitude float32, estimation_error float32) *AyiPacket {
	mb.message.Header.Type = M_USER_POSITION
	mb.message.SetMessage(&UserPosition{GlobalCoordinates: &Location{latitude, longitude}, EstimationError: estimation_error})
	return mb.message
}

func (mb *MessageBuilder) UserPositionRange(range_in_meters float32) *AyiPacket {
	mb.message.Header.Type = M_USER_POSITION_RANGE
	mb.message.SetMessage(&UserPositionRange{RangeInMeters: range_in_meters})
	return mb.message
}

func (mb *MessageBuilder) CreateUserAccount(name string, email string, password string, phone string, fbid string, fbtoken string) *AyiPacket {
	mb.message.Header.Type = M_USER_CREATE_ACCOUNT
	mb.message.SetMessage(&CreateUserAccount{Name: name, Email: email, Password: password, Phone: phone, Fbid: fbid, Fbtoken: fbtoken})
	return mb.message
}

func (mb *MessageBuilder) NewAuthTokenByEmail(email string, password string) *AyiPacket {
	mb.message.Header.Type = M_USER_NEW_AUTH_TOKEN
	mb.message.SetMessage(&NewAuthToken{Pass1: email, Pass2: password, Type: AuthType_A_NATIVE})
	return mb.message
}

func (mb *MessageBuilder) NewAuthTokenByFacebook(fbid string, fbtoken string) *AyiPacket {
	mb.message.Header.Type = M_USER_NEW_AUTH_TOKEN
	mb.message.SetMessage(&NewAuthToken{Pass1: fbid, Pass2: fbtoken, Type: AuthType_A_FACEBOOK})
	return mb.message
}

func (mb *MessageBuilder) UserAuthencation(user_id uint64, auth_token uuid.UUID) *AyiPacket {
	mb.message.Header.Type = M_USER_AUTH
	mb.message.SetMessage(&UserAuthentication{UserId: user_id, AuthToken: auth_token.String()})
	return mb.message
}

// Notifications
func (mb *MessageBuilder) EventCreated(event *Event) *AyiPacket {
	mb.message.Header.Type = M_EVENT_CREATED
	mb.message.SetMessage(event)
	return mb.message
}

func (mb *MessageBuilder) EventCancelled(event_id uint64, reason string) *AyiPacket {
	mb.message.Header.Type = M_EVENT_CANCELLED
	mb.message.SetMessage(&EventCancelled{EventId: event_id, Reason: reason})
	return mb.message
}

func (mb *MessageBuilder) EventExpired(event_id uint64) *AyiPacket {
	mb.message.Header.Type = M_EVENT_EXPIRED
	mb.message.SetMessage(&EventExpired{EventId: event_id})
	return mb.message
}

func (mb *MessageBuilder) EventDateModified(event_id uint64, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.Type = M_EVENT_DATE_MODIFIED
	mb.message.SetMessage(&EventModified{EventId: event_id, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *MessageBuilder) EventMessageModified(event_id uint64, message string) *AyiPacket {
	mb.message.Header.Type = M_EVENT_MESSAGE_MODIFIED
	mb.message.SetMessage(&EventModified{EventId: event_id, Message: message})
	return mb.message
}

func (mb *MessageBuilder) EventModified(event_id uint64, message string, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.Type = M_EVENT_MODIFIED
	mb.message.SetMessage(&EventModified{EventId: event_id, Message: message, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *MessageBuilder) InvitationReceived(event *Event) *AyiPacket {
	mb.message.Header.Type = M_INVITATION_RECEIVED
	mb.message.SetMessage(event)
	return mb.message
}

func (mb *MessageBuilder) InvitationCancelled(event_id uint64) *AyiPacket {
	mb.message.Header.Type = M_INVITATION_CANCELLED
	mb.message.SetMessage(&InvitationCancelled{EventId: event_id})
	return mb.message
}

func (mb *MessageBuilder) AttendanceStatus(event_id uint64, status []*EventParticipant) *AyiPacket {
	mb.message.Header.Type = M_ATTENDANCE_STATUS
	mb.message.SetMessage(&AttendanceStatus{EventId: event_id, AttendanceStatus: status})
	return mb.message
}

func (mb *MessageBuilder) EventChangeDateProposed(event_id uint64, change_id uint32, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.Type = M_EVENT_CHANGE_DATE_PROPOSED
	mb.message.SetMessage(&EventChangeProposed{EventId: event_id, ChangeId: change_id, StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *MessageBuilder) EventChangeMessageProposed(event_id uint64, change_id uint32, message string) *AyiPacket {
	mb.message.Header.Type = M_EVENT_CHANGE_MESSAGE_PROPOSED
	mb.message.SetMessage(&EventChangeProposed{EventId: event_id, ChangeId: change_id, Message: message})
	return mb.message
}

func (mb *MessageBuilder) EventChangeProposed(event_id uint64, change_id uint32, message string, start_date int64, end_date int64) *AyiPacket {
	mb.message.Header.Type = M_EVENT_CHANGE_PROPOSED
	mb.message.SetMessage(&EventChangeProposed{EventId: event_id, ChangeId: change_id, Message: message,
		StartDate: start_date, EndDate: end_date})
	return mb.message
}

func (mb *MessageBuilder) VotingStatus(event_id uint64, change_id uint32, start_date int64,
	end_date int64, votes_received uint32, votes_total uint32) *AyiPacket {
	mb.message.Header.Type = M_VOTING_STATUS
	mb.message.SetMessage(&VotingStatus{EventId: event_id, ChangeId: change_id,
		StartDate: start_date, EndDate: end_date, ElapsedTime: end_date - start_date,
		VotesReceived: votes_received, VotesTotal: votes_total})
	return mb.message
}

func (mb *MessageBuilder) VotingFinished(event_id uint64, change_id uint32) *AyiPacket {
	mb.message.Header.Type = M_VOTING_FINISHED
	mb.message.SetMessage(&VotingStatus{EventId: event_id, ChangeId: change_id, Finished: true})
	return mb.message
}

func (mb *MessageBuilder) ChangeAccepted(event_id uint64, change_id uint32) *AyiPacket {
	mb.message.Header.Type = M_CHANGE_ACCEPTED
	mb.message.SetMessage(&ChangeAccepted{EventId: event_id, ChangeId: change_id})
	return mb.message
}

func (mb *MessageBuilder) ChangeDiscarded(event_id uint64, change_id uint32) *AyiPacket {
	mb.message.Header.Type = M_CHANGE_DISCARDED
	mb.message.SetMessage(&ChangeDiscarded{EventId: event_id, ChangeId: change_id})
	return mb.message
}

func (mb *MessageBuilder) UserAccessGranted(user_id uint64, auth_token uuid.UUID) *AyiPacket {
	mb.message.Header.Type = M_ACCESS_GRANTED
	mb.message.SetMessage(&AccessGranted{UserId: user_id, AuthToken: auth_token.String()})
	return mb.message
}

func (mb *MessageBuilder) Ok(msg_type PacketType) *AyiPacket {
	mb.message.Header.Type = M_OK
	mb.message.SetMessage(&Ok{Type: int32(msg_type)})
	return mb.message
}

func (mb *MessageBuilder) Error(msg_type PacketType, error_code int32) *AyiPacket {
	mb.message.Header.Type = M_ERROR
	mb.message.SetMessage(&Error{Type: int32(msg_type), Error: error_code})
	return mb.message
}

// Requests
func (mb *MessageBuilder) Ping() *AyiPacket {
	mb.message.Header.Type = M_PING
	mb.message.SetMessage(&Ping{time.Now().Unix()})
	return mb.message
}

func (mb *MessageBuilder) ReadEvent(event_id uint64) *AyiPacket {
	mb.message.Header.Type = M_READ_EVENT
	mb.message.SetMessage(&ReadEvent{EventId: event_id})
	return mb.message
}

func (mb *MessageBuilder) ListAuthoredEvents(cursor uint32) *AyiPacket {
	mb.message.Header.Type = M_LIST_AUTHORED_EVENTS
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}

func (mb *MessageBuilder) ListPrivateEvents(cursor uint32) *AyiPacket {
	mb.message.Header.Type = M_LIST_PRIVATE_EVENTS
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}

func (mb *MessageBuilder) ListPublicEvents(latitude float32, longitude float32, range_in_meters uint32, cursor uint32) *AyiPacket {
	mb.message.Header.Type = M_LIST_PUBLIC_EVENTS
	mb.message.SetMessage(&ListPublicEvents{UserCoordinates: &Location{latitude, longitude},
		RangeInMeters: range_in_meters, Cursor: &ListCursor{cursor}})
	return mb.message
}

func (mb *MessageBuilder) HistoryAuthoredEvents(cursor uint32) *AyiPacket {
	mb.message.Header.Type = M_HISTORY_AUTHORED_EVENTS
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}

func (mb *MessageBuilder) HistoryPrivateEvents(cursor uint32) *AyiPacket {
	mb.message.Header.Type = M_HISTORY_PRIVATE_EVENTS
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}

func (mb *MessageBuilder) HistoryPublicEvents(cursor uint32) *AyiPacket {
	mb.message.Header.Type = M_HISTORY_PUBLIC_EVENTS
	mb.message.SetMessage(&ListCursor{Cursor: cursor})
	return mb.message
}

func (mb *MessageBuilder) UserFriends() *AyiPacket {
	mb.message.Header.Type = M_USER_FRIENDS
	return mb.message
}

// Responses
func (mb *MessageBuilder) Pong() *AyiPacket {
	mb.message.Header.Type = M_PONG
	mb.message.SetMessage(&Pong{time.Now().Unix()})
	return mb.message
}

/* Create an Event message with all its information, including participants information */
func (mb *MessageBuilder) EventInfo(event *Event) *AyiPacket {
	mb.message.Header.Type = M_EVENT_INFO
	mb.message.SetMessage(event)
	return mb.message
}

/* Create a List of events */
func (mb *MessageBuilder) EventsList(events_list []*Event) *AyiPacket {
	mb.message.Header.Type = M_EVENTS_LIST
	mb.message.SetMessage(&EventsList{Event: events_list})
	return mb.message
}

func (mb *MessageBuilder) FriendsList(friends_list []*Friend) *AyiPacket {
	mb.message.Header.Type = M_FRIENDS_LIST
	mb.message.SetMessage(&FriendsList{Friends: friends_list})
	return mb.message
}
