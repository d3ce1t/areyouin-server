package common

import (
	"github.com/twinj/uuid"
)

type EventInbox struct {
	UserId     int64
	EventId    int64
	AuthorId   int64
	AuthorName string
	StartDate  int64
	Message    string
	Response   AttendanceResponse
}

type UserFriend interface {
	GetName() string
	GetUserId() int64
	GetPictureDigest() []byte
}

type IIDToken struct {
	Token string
	Version int
}

type Picture struct {
	RawData []byte
	Digest  []byte
}

type EventDAO interface {
	InsertEventAndParticipants(event *Event) error
	LoadEventPicture(event_id int64) ([]byte, error)
	LoadEvent(event_ids ...int64) (events []*Event, err error)
	LoadEventAndParticipants(event_ids ...int64) (events []*Event, err error)
	LoadParticipant(event_id int64, user_id int64) (*Participant, error)

	LoadUserInbox(user_id int64, fromDate int64) ([]*EventInbox, error)
	LoadUserInboxReverse(user_id int64, fromDate int64) ([]*EventInbox, error)
	LoadUserInboxBetween(user_id int64, fromDate int64, toDate int64) ([]*EventInbox, error)

	LoadUserEventsAndParticipants(user_id int64, fromDate int64) ([]*Event, error)
	LoadUserEventsHistoryAndparticipants(user_id int64, fromDate int64, toDate int64) ([]*Event, error)

	InsertEventToUserInbox(participant *EventParticipant, event *Event) error
	AddOrUpdateEventToUserInbox(participant *EventParticipant, event *Event) error
	CompareAndSetNumGuests(event_id int64, num_guests int) (bool, error)
	//SetNumGuests(event_id int64, num_guests int32) error
	CompareAndSetNumAttendees(event_id int64, num_attendees int) (bool, error)
	//SetNumAttendees(event_id int64, num_attendees int) error
	SetParticipantStatus(user_id int64, event_id int64, status MessageStatus) error
	SetParticipantResponse(participant *Participant, response AttendanceResponse) error
	//SetUserEventInboxPosition(participant *EventParticipant, event *Event, new_position int64) error
	SetEventStateAndInboxPosition(event_id int64, new_status EventState, new_position int64) error
	SetEventPicture(event_id int64, picture *Picture) error
}

type UserDAO interface {
	CheckValidAccountObject(user_id int64, email string, fb_id string, check_credentials bool) (bool, error)
	CheckValidAccount(user_id int64, check_credentials bool) (bool, error)
	GetIDByEmailAndPassword(email string, password string) (int64, error)
	GetIDByFacebookID(fb_id string) (int64, error)
	LoadWithPicture(user_id int64) (*UserAccount, error)
	Load(user_id int64) (*UserAccount, error)
	LoadByEmail(email string) (*UserAccount, error)
	LoadAllUsers() ([]*UserAccount, error)
	LoadEmailCredential(email string) (credent *EmailCredential, err error)
	LoadFacebookCredential(fbid string) (credent *FacebookCredential, err error)
	LoadUserPicture(user_id int64) ([]byte, error)
	GetIIDToken(user_id int64) (*IIDToken, error)
	Insert(user *UserAccount) error
	SaveProfilePicture(user_id int64, picture *Picture) error
	SetAuthToken(user_id int64, auth_token uuid.UUID) error
	SetLastConnection(user_id int64, time int64) error
	SetFacebookAccessToken(user_id int64, fb_id string, fb_token string) error
	SetAuthTokenAndFBToken(user_id int64, auth_token uuid.UUID, fb_id string, fb_token string) error
	SetIIDToken(user_id int64, iid_token *IIDToken) error
	ResetEmailCredentialPassword(user_id int64, email string, password string) (ok bool, err error)
	Delete(user *UserAccount) error
	DeleteUserAccount(user_id int64) error
	DeleteEmailCredentials(email string) error
	DeleteFacebookCredentials(fb_id string) error
}

type FriendDAO interface {
	LoadFriends(user_id int64, group_id int32) ([]*Friend, error)
	LoadFriendsMap(user_id int64) (map[int64]*Friend, error)
	IsFriend(user_id int64, other_user_id int64) (bool, error)
	AreFriends(user_id int64, other_user_id int64) (bool, error)
	MakeFriends(user1 UserFriend, user2 UserFriend) error
	SetPictureDigest(user_id int64, friend_id int64, digest []byte) error
	LoadGroups(user_id int64) ([]*Group, error)
	LoadGroupsAndMembers(user_id int64) ([]*Group, error)
	AddGroup(user_id int64, group *Group) error
	SetGroupName(user_id int64, group_id int32, name string) error
	AddMembers(user_id int64, group_id int32, friend_ids ...int64) error
	DeleteMembers(user_id int64, group_id int32, friend_ids ...int64) error
	DeleteGroup(user_id int64, group_id int32) error
	LoadFriendRequest(user_id int64, friend_id int64) (*FriendRequest, error)
	LoadFriendRequests(user_id int64) ([]*FriendRequest, error)
	ExistFriendRequest(user_id int64, friend_id int64) (bool, error)
	InsertFriendRequest(user_id int64, friend_id int64, name string, email string, created_date int64) error
	DeleteFriendRequest(user_id int64, friend_id int64, created_date int64) error
}

type ThumbnailDAO interface {
	Insert(id int64, digest []byte, thumbnails map[int32][]byte) error
	Load(id int64, dpi int32) ([]byte, error)
	Remove(id int64) error
}

type AccessTokenDAO interface {
	Insert(user_id int64, token string) error
	CheckAccessToken(user_id int64, access_token string) (bool, error)
	SetLastUsed(user_id int64, time int64) error
	Remove(user_id int64) error
}
