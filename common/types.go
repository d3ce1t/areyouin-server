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
	GetPictureDigest() []byte
}

type Picture struct {
	RawData []byte
	Digest  []byte
}

type EventDAO interface {
	InsertEventAndParticipants(event *Event) error
	LoadEventPicture(event_id uint64) ([]byte, error)
	LoadEvent(event_ids ...uint64) (events []*Event, err error)
	LoadEventAndParticipants(event_ids ...uint64) (events []*Event, err error)
	LoadParticipant(event_id uint64, user_id uint64) (*Participant, error)

	LoadUserInbox(user_id uint64, fromDate int64) ([]*EventInbox, error)
	LoadUserInboxReverse(user_id uint64, fromDate int64) ([]*EventInbox, error)
	LoadUserInboxBetween(user_id uint64, fromDate int64, toDate int64) ([]*EventInbox, error)

	LoadUserEventsAndParticipants(user_id uint64, fromDate int64) ([]*Event, error)
	LoadUserEventsHistoryAndparticipants(user_id uint64, fromDate int64, toDate int64) ([]*Event, error)

	InsertEventToUserInbox(participant *EventParticipant, event *Event) error
	AddOrUpdateEventToUserInbox(participant *EventParticipant, event *Event) error
	CompareAndSetNumGuests(event_id uint64, num_guests int) (bool, error)
	//SetNumGuests(event_id uint64, num_guests int32) error
	CompareAndSetNumAttendees(event_id uint64, num_attendees int) (bool, error)
	//SetNumAttendees(event_id uint64, num_attendees int) error
	SetParticipantStatus(user_id uint64, event_id uint64, status MessageStatus) error
	SetParticipantResponse(participant *Participant, response AttendanceResponse) error
	//SetUserEventInboxPosition(participant *EventParticipant, event *Event, new_position int64) error
	SetEventStateAndInboxPosition(event_id uint64, new_status EventState, new_position int64) error
	SetEventPicture(event_id uint64, picture *Picture) error
}

type UserDAO interface {
	CheckValidAccountObject(user_id uint64, email string, fb_id string, check_credentials bool) (bool, error)
	CheckValidAccount(user_id uint64, check_credentials bool) (bool, error)
	GetIDByEmailAndPassword(email string, password string) (uint64, error)
	GetIDByFacebookID(fb_id string) (uint64, error)
	LoadWithPicture(user_id uint64) (*UserAccount, error)
	Load(user_id uint64) (*UserAccount, error)
	LoadByEmail(email string) (*UserAccount, error)
	LoadAllUsers() ([]*UserAccount, error)
	LoadEmailCredential(email string) (credent *EmailCredential, err error)
	LoadFacebookCredential(fbid string) (credent *FacebookCredential, err error)
	LoadUserPicture(user_id uint64) ([]byte, error)
	GetIIDToken(user_id uint64) (string, error)
	Insert(user *UserAccount) error
	SaveProfilePicture(user_id uint64, picture *Picture) error
	SetAuthToken(user_id uint64, auth_token uuid.UUID) error
	SetLastConnection(user_id uint64, time int64) error
	SetFacebookAccessToken(user_id uint64, fb_id string, fb_token string) error
	SetAuthTokenAndFBToken(user_id uint64, auth_token uuid.UUID, fb_id string, fb_token string) error
	SetIIDToken(user_id uint64, iid_token string) error
	ResetEmailCredentialPassword(user_id uint64, email string, password string) (ok bool, err error)
	Delete(user *UserAccount) error
	DeleteUserAccount(user_id uint64) error
	DeleteEmailCredentials(email string) error
	DeleteFacebookCredentials(fb_id string) error
}

type FriendDAO interface {
	LoadFriends(user_id uint64, group_id int32) ([]*Friend, error)
	LoadFriendsMap(user_id uint64) (map[uint64]*Friend, error)
	IsFriend(user_id uint64, other_user_id uint64) (bool, error)
	AreFriends(user_id uint64, other_user_id uint64) (bool, error)
	MakeFriends(user1 UserFriend, user2 UserFriend) error
	SetPictureDigest(user_id uint64, friend_id uint64, digest []byte) error
	LoadGroups(user_id uint64) ([]*Group, error)
	LoadGroupsAndMembers(user_id uint64) ([]*Group, error)
	AddGroup(user_id uint64, group *Group) error
	SetGroupName(user_id uint64, group_id int32, name string) error
	AddMembers(user_id uint64, group_id int32, friend_ids ...uint64) error
	DeleteMembers(user_id uint64, group_id int32, friend_ids ...uint64) error
	DeleteGroup(user_id uint64, group_id int32) error
}

type ThumbnailDAO interface {
	Insert(id uint64, digest []byte, thumbnails map[int32][]byte) error
	Load(id uint64, dpi int32) ([]byte, error)
	Remove(id uint64) error
}

type AccessTokenDAO interface {
	Insert(user_id uint64, token string) error
	CheckAccessToken(user_id uint64, access_token string) (bool, error)
	SetLastUsed(user_id uint64, time int64) error
	Remove(user_id uint64) error
}
