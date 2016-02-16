package common

import (
	"github.com/twinj/uuid"
)

type EventInbox struct {
	UserId     uint64
	EventId    uint64
	AuthorId   uint64
	AuthorName string
	StartDate  int64
	Message    string
	Response   AttendanceResponse
}

type UserFriend interface {
	GetName() string
	GetUserId() uint64
}

type EventDAO interface {
	InsertEventCAS(event *Event) (ok bool, err error)
	InsertEvent(event *Event) error
	InsertEventAndParticipants(event *Event) error
	LoadEvent(event_ids ...uint64) (events []*Event, err error)
	LoadEventAndParticipants(event_ids ...uint64) (events []*Event, err error)
	LoadParticipant(event_id uint64, user_id uint64) (*Participant, error)
	LoadUserInbox(user_id uint64, fromDate int64, toDate int64) ([]*EventInbox, error)
	LoadUserEvents(user_id uint64, fromDate int64) (events []*Event, err error)
	LoadUserEventsAndParticipants(user_id uint64, fromDate int64) ([]*Event, error)
	AddOrUpdateParticipants(event_id uint64, participantList map[uint64]*EventParticipant) error
	InsertEventToUserInbox(participant *EventParticipant, event *Event) error
	AddOrUpdateEventToUserInbox(participant *EventParticipant, event *Event) error
	CompareAndSetNumGuests(event_id uint64, num_guests int32) (bool, error)
	SetNumGuests(event_id uint64, num_guests int32) error
	CompareAndSetNumAttendees(event_id uint64, num_attendees int) (bool, error)
	SetNumAttendees(event_id uint64, num_attendees int) error
	SetParticipantStatus(user_id uint64, event_id uint64, status MessageStatus) error
	SetParticipantResponse(participant *Participant, response AttendanceResponse) error
}

type UserDAO interface {
	CheckValidAccount(user_id uint64, check_credentials bool) (bool, error)
	CheckValidCredentials(user_id uint64, email string, fb_id string) (bool, error)
	CheckEmailCredentials(email string, password string) (uint64, error)
	CheckAuthToken(user_id uint64, auth_token uuid.UUID) (bool, error)
	ExistEmail(email string) (bool, error)
	//ExistsUserAccount(user_id uint64) (bool, error)
	SetAuthToken(user_id uint64, auth_token uuid.UUID) error
	SetLastConnection(user_id uint64, time int64) error
	SetFacebookAccessToken(user_id uint64, fb_id string, fb_token string) error
	SetAuthTokenAndFBToken(user_id uint64, auth_token uuid.UUID, fb_id string, fb_token string) error
	SetIIDToken(user_id uint64, iid_token string) error
	GetIDByEmail(email string) (uint64, error)
	GetIDByFacebookID(fb_id string) (uint64, error)
	MakeFriends(user1 UserFriend, user2 UserFriend) error
	IsFriend(user_id uint64, other_user_id uint64) (bool, error)
	AreFriends(user_id uint64, other_user_id uint64) (bool, error)
	Insert(user *UserAccount) error
	SaveProfilePicture(user_id uint64, picture []byte) error
	//AddFriend(user_id uint64, friend *Friend, group_id int32) error
	LoadWithPicture(user_id uint64) (*UserAccount, error)
	Load(user_id uint64) (*UserAccount, error)
	LoadByEmail(email string) (*UserAccount, error)
	LoadAllUsers() ([]*UserAccount, error)
	LoadFriends(user_id uint64, group_id int32) ([]*Friend, error)
	LoadFriendsIndex(user_id uint64, group_id int32) (map[uint64]*Friend, error)
	LoadEmailCredential(email string) (credent *EmailCredential, err error)
	LoadFacebookCredential(fbid string) (credent *FacebookCredential, err error)
	Delete(user *UserAccount) error
	DeleteUserAccount(user_id uint64) error
	DeleteEmailCredentials(email string) error
	DeleteFacebookCredentials(fb_id string) error
	DeleteFriendsGroup(user_id uint64, group_id int32) error
}
