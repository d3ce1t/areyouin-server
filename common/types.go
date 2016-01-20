package common

import (
	"github.com/twinj/uuid"
)

type EventDAO interface {
	InsertEventCAS(event *Event) (ok bool, err error)
	InsertEvent(event *Event) error
	InsertEventAndParticipants(event *Event) error
	LoadEvent(event_ids ...uint64) (events []*Event, err error)
	LoadEventAndParticipants(event_ids ...uint64) (events []*Event, err error)
	LoadParticipant(event_id uint64, user_id uint64) (*EventParticipant, error)
	LoadAllParticipants(event_id uint64) ([]*EventParticipant, error)
	LoadUserInbox(user_id uint64, fromDate int64) ([]uint64, error)
	LoadUserEvents(user_id uint64, fromDate int64) (events []*Event, err error)
	LoadUserEventsAndParticipants(user_id uint64, fromDate int64) ([]*Event, error)
	AddOrUpdateParticipants(event_id uint64, participantList []*EventParticipant) error
	AddEventToUserInbox(user_id uint64, event *Event, response AttendanceResponse) error
	CompareAndSetNumGuests(event_id uint64, num_guests int32) (bool, error)
	SetNumGuests(event_id uint64, num_guests int32) error
	CompareAndSetNumAttendees(event_id uint64, num_attendees int) (bool, error)
	SetNumAttendees(event_id uint64, num_attendees int) error
	SetParticipantStatus(user_id uint64, event_id uint64, status MessageStatus) error
	SetParticipantResponse(user_id uint64, event_id uint64, response AttendanceResponse) error
}

type UserDAO interface {
	CheckEmailCredentials(email string, password string) (uint64, error)
	CheckAuthToken(user_id uint64, auth_token uuid.UUID) (bool, error)
	ExistWithSanity(user *UserAccount) (bool, error)
	SetAuthToken(user_id uint64, auth_token uuid.UUID) error
	SetLastConnection(user_id uint64, time int64) error
	SetFacebookAccessToken(user_id uint64, fb_id string, fb_token string) error
	SetAuthTokenAndFBToken(user_id uint64, auth_token uuid.UUID, fb_id string, fb_token string) error
	GetIDByEmail(email string) (uint64, error)
	GetIDByFacebookID(fb_id string) (uint64, error)
	Exists(user_id uint64) (bool, error)
	Load(user_id uint64) (*UserAccount, error)
	LoadByEmail(email string) (*UserAccount, error)
	Insert(user *UserAccount) error
	Delete(user *UserAccount) error
	DeleteEmailCredentials(email string) error
	DeleteFacebookCredentials(fb_id string) error
	MakeFriends(user1 *Friend, user2 *Friend) error
	AddFriend(user_id uint64, friend *Friend, group_id int32) error
	DeleteFriendsGroup(user_id uint64, group_id int32) error
	LoadFriends(user_id uint64, group_id int32) ([]*Friend, error)
	LoadFriendsIndex(user_id uint64, group_id int32) (map[uint64]*Friend, error)
	AreFriends(user_id uint64, other_user_id uint64) (bool, error)
}
