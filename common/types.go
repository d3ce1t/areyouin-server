package common

import (
	"github.com/twinj/uuid"
)

type EventDAO interface {
	Insert(event *Event) (ok bool, err error)
	//EventHasParticipant(event_id uint64, user_id uint64) bool
	LoadParticipant(event_id uint64, user_id uint64) (*EventParticipant, error)
	LoadAllParticipants(event_id uint64) []*EventParticipant
	AddOrUpdateParticipants(event_id uint64, participantList []*EventParticipant) error
	AddEventToUserInbox(user_id uint64, event *Event, response AttendanceResponse) error
	LoadUserEvents(user_id uint64, fromDate int64) (events []*Event, err error)
	CompareAndSetNumGuests(event_id uint64, num_guests int32) (bool, error)
	SetNumGuests(event_id uint64, num_guests int32) error
	CompareAndSetNumAttendees(event_id uint64, num_attendees int) (bool, error)
	SetNumAttendees(event_id uint64, num_attendees int) error
	SetParticipantStatus(user_id uint64, event_id uint64, status MessageStatus) error
	SetParticipantResponse(user_id uint64, event_id uint64, response AttendanceResponse) error
}

type UserDAO interface {
	CheckEmailCredentials(email string, password string) uint64
	CheckAuthToken(user_id uint64, auth_token uuid.UUID) bool
	ExistWithSanity(user *UserAccount) bool
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
	AddFriend(user_id uint64, friend *Friend, group_id int32) error
	LoadFriends(user_id uint64, group_id int32) []*Friend
	AreFriends(user_id uint64, other_user_id uint64) bool
}
