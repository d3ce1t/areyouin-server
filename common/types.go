package common

import (
	proto "areyouin/protocol"
	"github.com/twinj/uuid"
)

type EventDAO interface {
	Insert(event *proto.Event) (ok bool, err error)
	EventHasParticipant(event_id uint64, user_id uint64) bool
	LoadParticipants(event_id uint64) []*proto.EventParticipant
	AddOrUpdateParticipants(event_id uint64, participantList []*proto.EventParticipant) error
	AddEventToUserInbox(user_id uint64, event *proto.Event, response proto.AttendanceResponse) error
	LoadUserEvents(user_id uint64, fromDate int64) (events []*proto.Event, err error)
	SetNumGuests(event_id uint64, num_guests int32) error
	SetNumAttendees(event_id uint64, num_attendees int) error
	SetParticipantStatus(user_id uint64, event_id uint64, status proto.MessageStatus) error
	SetParticipantResponse(user_id uint64, event_id uint64, response proto.AttendanceResponse) error
}

type UserDAO interface {
	CheckEmailCredentials(email string, password string) uint64
	CheckAuthToken(user_id uint64, auth_token uuid.UUID) bool
	SetAuthToken(user_id uint64, auth_token uuid.UUID) error
	SetLastConnection(user_id uint64, time int64) error
	SetFacebookAccessToken(user_id uint64, fb_id string, fb_token string) error
	SetAuthTokenAndFBToken(user_id uint64, auth_token uuid.UUID, fb_id string, fb_token string) error
	GetIDByEmail(email string) uint64
	GetIDByFacebookID(fb_id string) uint64
	Exists(user_id uint64) bool
	Load(user_id uint64) *UserAccount
	LoadByEmail(email string) *UserAccount
	Insert(user *UserAccount) (ok bool, err error)
	Delete(user *UserAccount) error
	DeleteEmailCredentials(email string) error
	DeleteFacebookCredentials(fb_id string) error
	AddFriend(user_id uint64, friend *proto.Friend, group_id int32) error
	LoadFriends(user_id uint64, group_id int32) []*proto.Friend
	AreFriends(user_id uint64, other_user_id uint64) bool
}
