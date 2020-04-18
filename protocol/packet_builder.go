package protocol

import (
	core "github.com/d3ce1t/areyouin-server/protocol/core"
	"github.com/d3ce1t/areyouin-server/utils"
)

type PacketBuilder struct {
	message *AyiPacket
}

func (mb *PacketBuilder) NewAccessToken(user_id int64, auth_token string) *AyiPacket {
	mb.message.Header.SetType(M_ACCESS_TOKEN)
	mb.message.SetMessage(&AccessToken{UserId: user_id, AuthToken: auth_token})
	return mb.message
}

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

func (mb *PacketBuilder) AttendanceStatus(event_id int64, participants map[int64]*core.EventParticipant) *AyiPacket {
	mb.message.Header.SetType(M_ATTENDANCE_STATUS)
	participantsSlice := make([]*core.EventParticipant, 0, len(participants))
	for _, v := range participants {
		participantsSlice = append(participantsSlice, v)
	}
	mb.message.SetMessage(&AttendanceStatus{EventId: event_id, AttendanceStatus: participantsSlice})
	return mb.message
}

func (mb *PacketBuilder) AttendanceStatusWithNumGuests(event_id int64, participants map[int64]*core.EventParticipant, num_guests int) *AyiPacket {
	mb.message.Header.SetType(M_ATTENDANCE_STATUS)
	participantsSlice := make([]*core.EventParticipant, 0, len(participants))
	for _, v := range participants {
		participantsSlice = append(participantsSlice, v)
	}
	mb.message.SetMessage(&AttendanceStatus{EventId: event_id, AttendanceStatus: participantsSlice, NumGuests: int32(num_guests)})
	return mb.message
}

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
	mb.message.SetMessage(&TimeInfo{utils.GetCurrentTimeMillis()})
	return mb.message
}

// Responses
func (mb *PacketBuilder) Pong() *AyiPacket {
	mb.message.Header.SetType(M_PONG)
	mb.message.SetMessage(&TimeInfo{utils.GetCurrentTimeMillis()})
	return mb.message
}

/* Create an Event message with all its information, including participants information */
func (mb *PacketBuilder) Event(event *core.Event) *AyiPacket {
	mb.message.Header.SetType(M_EVENT)
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

func (mb *PacketBuilder) FacebookFriendsList(friends_list []*core.Friend) *AyiPacket {
	mb.message.Header.SetType(M_FACEBOOK_FRIENDS_LIST)
	mb.message.SetMessage(&FriendsList{Friends: friends_list})
	return mb.message
}

func (mb *PacketBuilder) ClockResponse() *AyiPacket {
	mb.message.Header.SetType(M_CLOCK_RESPONSE)
	mb.message.SetMessage(&TimeInfo{utils.GetCurrentTimeMillis()})
	return mb.message
}

func (mb *PacketBuilder) UserAccount(user *core.UserAccount) *AyiPacket {
	mb.message.Header.SetType(M_USER_ACCOUNT)
	mb.message.SetMessage(user)
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
