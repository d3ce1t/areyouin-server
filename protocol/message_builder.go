package protocol

import (
	"github.com/twinj/uuid"
	core "peeple/areyouin/common"
)

// Interface
type MessageBuilder interface {
	CreateEvent(message string, start_date int64, end_date int64, participants []uint64) *AyiPacket
	CancelEvent(event_id uint64, reason string) *AyiPacket
	InviteUsers(event_id uint64, participants []uint64) *AyiPacket
	CancelUsersInvitation(event_id uint64, participants []uint64) *AyiPacket
	ConfirmAttendance(event_id uint64, action_code core.AttendanceResponse) *AyiPacket
	ModifyEventDate(event_id uint64, start_date int64, end_date int64) *AyiPacket
	ModifyEventMessage(event_id uint64, message string) *AyiPacket
	ModifyEvent(event_id uint64, message string, start_date int64, end_date int64) *AyiPacket
	VoteChange(event_id uint64, change_id uint32, accept_change bool) *AyiPacket
	UserPosition(latitude float32, longitude float32, estimation_error float32) *AyiPacket
	UserPositionRange(range_in_meters float32) *AyiPacket
	CreateUserAccount(name string, email string, password string, phone string, fbid string, fbtoken string) *AyiPacket
	NewAuthTokenByEmail(email string, password string) *AyiPacket
	NewAuthTokenByFacebook(fbid string, fbtoken string) *AyiPacket
	UserAuthencation(user_id uint64, auth_token uuid.UUID) *AyiPacket
	IIDToken(token string) *AyiPacket
	ChangeProfilePicture(picture []byte) *AyiPacket
	EventCreated(event *core.Event) *AyiPacket
	EventCancelled(event_id uint64, reason string) *AyiPacket
	EventExpired(event_id uint64) *AyiPacket
	EventDateModified(event_id uint64, start_date int64, end_date int64) *AyiPacket
	EventMessageModified(event_id uint64, message string) *AyiPacket
	EventModified(event_id uint64, message string, start_date int64, end_date int64) *AyiPacket
	InvitationReceived(event *core.Event) *AyiPacket
	InvitationCancelled(event_id uint64) *AyiPacket
	AttendanceStatus(event_id uint64, status []*core.EventParticipant) *AyiPacket
	EventChangeDateProposed(event_id uint64, change_id uint32, start_date int64, end_date int64) *AyiPacket
	EventChangeMessageProposed(event_id uint64, change_id uint32, message string) *AyiPacket
	EventChangeProposed(event_id uint64, change_id uint32, message string, start_date int64, end_date int64) *AyiPacket
	UserAccessGranted(user_id uint64, auth_token uuid.UUID) *AyiPacket
	Ok(msg_type PacketType) *AyiPacket
	Error(msg_type PacketType, error_code int32) *AyiPacket
	Ping() *AyiPacket
	ReadEvent(event_id uint64) *AyiPacket
	ListAuthoredEvents(cursor uint32) *AyiPacket
	ListPrivateEvents(cursor uint32) *AyiPacket
	ListPublicEvents(latitude float32, longitude float32, range_in_meters uint32, cursor uint32) *AyiPacket
	HistoryAuthoredEvents(cursor uint32) *AyiPacket
	HistoryPrivateEvents(cursor uint32) *AyiPacket
	HistoryPublicEvents(cursor uint32) *AyiPacket
	UserFriends() *AyiPacket
	Pong() *AyiPacket
	EventInfo(event *core.Event) *AyiPacket
	EventsList(events_list []*core.Event) *AyiPacket
	FriendsList(friends_list []*core.Friend) *AyiPacket
	ClockResponse() *AyiPacket
	UserAccount(user *core.UserAccount) *AyiPacket
}
