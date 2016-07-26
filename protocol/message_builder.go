package protocol

import (
	core "peeple/areyouin/protocol/core"
)

// Interface
type MessageBuilder interface {
	NewAccessToken(user_id int64, auth_token string) *AyiPacket
	EventCreated(event *core.Event) *AyiPacket
	EventCancelled(who_id int64, event *core.Event) *AyiPacket
	EventExpired(event_id int64) *AyiPacket
	EventModified(event *core.Event) *AyiPacket
	InvitationReceived(event *core.Event) *AyiPacket
	AttendanceStatus(event_id int64, participants map[int64]*core.EventParticipant) *AyiPacket
	AttendanceStatusWithNumGuests(event_id int64, status map[int64]*core.EventParticipant, num_guests int32) *AyiPacket
	UserAccessGranted(user_id int64, auth_token string) *AyiPacket
	Ok(msg_type PacketType) *AyiPacket
	Error(msg_type PacketType, error_code int32) *AyiPacket
	Ping() *AyiPacket
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
