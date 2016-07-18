package protocol

import (
	core "peeple/areyouin/common"
)

// Interface
type MessageBuilder interface {
	//CreateEvent(message string, start_date int64, end_date int64, participants []int64) *AyiPacket
	//CancelEvent(event_id int64, reason string) *AyiPacket
	//InviteUsers(event_id int64, participants []int64) *AyiPacket
	//CancelUsersInvitation(event_id int64, participants []int64) *AyiPacket
	//ConfirmAttendance(event_id int64, action_code core.AttendanceResponse) *AyiPacket
	//ModifyEventDate(event_id int64, start_date int64, end_date int64) *AyiPacket
	//ModifyEventMessage(event_id int64, message string) *AyiPacket
	//ModifyEvent(event_id int64, message string, start_date int64, end_date int64) *AyiPacket
	//VoteChange(event_id int64, change_id int32, accept_change bool) *AyiPacket
	//UserPosition(latitude float32, longitude float32, estimation_error float32) *AyiPacket
	//UserPositionRange(range_in_meters float32) *AyiPacket
	//CreateUserAccount(name string, email string, password string, phone string, fbid string, fbtoken string) *AyiPacket
	NewAccessToken(user_id int64, auth_token string) *AyiPacket
	//NewAuthTokenByEmail(email string, password string) *AyiPacket
	//NewAuthTokenByFacebook(fbid string, fbtoken string) *AyiPacket
	//UserAuthencation(user_id int64, auth_token uuid.UUID) *AyiPacket
	//IIDToken(token string) *AyiPacket
	//ChangeProfilePicture(picture []byte) *AyiPacket
	EventCreated(event *core.Event) *AyiPacket
	EventCancelled(who_id int64, event *core.Event) *AyiPacket
	EventExpired(event_id int64) *AyiPacket
	//EventDateModified(event_id int64, start_date int64, end_date int64) *AyiPacket
	//EventMessageModified(event_id int64, message string) *AyiPacket
	EventModified(event *core.Event) *AyiPacket
	InvitationReceived(event *core.Event) *AyiPacket
	//InvitationCancelled(event_id int64) *AyiPacket
	AttendanceStatus(event_id int64, status []*core.EventParticipant) *AyiPacket
	AttendanceStatusWithNumGuests(event_id int64, status []*core.EventParticipant, num_guests int32) *AyiPacket
	//EventChangeDateProposed(event_id int64, change_id int32, start_date int64, end_date int64) *AyiPacket
	//EventChangeMessageProposed(event_id int64, change_id int32, message string) *AyiPacket
	//EventChangeProposed(event_id int64, change_id int32, message string, start_date int64, end_date int64) *AyiPacket
	UserAccessGranted(user_id int64, auth_token string) *AyiPacket
	Ok(msg_type PacketType) *AyiPacket
	Error(msg_type PacketType, error_code int32) *AyiPacket
	Ping() *AyiPacket
	//ReadEvent(event_id int64) *AyiPacket
	//ListAuthoredEvents(cursor uint32) *AyiPacket
	//ListPrivateEvents(cursor uint32) *AyiPacket
	//ListPublicEvents(latitude float32, longitude float32, range_in_meters uint32, cursor uint32) *AyiPacket
	//HistoryAuthoredEvents(cursor uint32) *AyiPacket
	//HistoryPrivateEvents(cursor uint32) *AyiPacket
	//HistoryPublicEvents(cursor uint32) *AyiPacket
	//UserFriends() *AyiPacket
	Pong() *AyiPacket
	EventInfo(event *core.Event) *AyiPacket
	EventsList(events_list []*core.Event) *AyiPacket
	EventsHistoryList(events_list []*core.Event, startWindow int64, endWindow int64) *AyiPacket
	FriendsList(friends_list []*core.Friend) *AyiPacket
	ClockResponse() *AyiPacket
	UserAccount(user *core.UserAccount) *AyiPacket
	GroupsList(groups_list []*core.Group) *AyiPacket
	FriendRequestReceived(request *core.FriendRequest) *AyiPacket
	FriendRequestsList(requests_list []*core.FriendRequest) *AyiPacket
}
