package api

import "time"

type DbSession interface {
	Connect() error
	IsValid() bool
	Closed() bool
}

type SettingsDAO interface {
	Find(key SettingOption) (string, error)
	Insert(key SettingOption, value string) error
}

type UserDAO interface {
	Load(userId int64) (*UserDTO, error)
	LoadByEmail(email string) (*UserDTO, error)
	LoadByFB(fbId string) (*UserDTO, error)
	LoadProfilePicture(userId int64) (*PictureDTO, error)
	LoadIIDToken(userId int64) (*IIDTokenDTO, error)
	Insert(user *UserDTO) error
	InsertFacebookCredentials(userId int64, fbId string, fbToken string) (ok bool, err error)
	SetPassword(email string, newPassword [32]byte, newSalt [32]byte) (bool, error)
	SaveProfilePicture(userId int64, picture *PictureDTO) error
	SetLastConnection(userId int64, time int64) error
	SetAuthToken(userId int64, auth_token string) error
	SetFacebookCredential(userId int64, fbId string, fbToken string) error
	SetFacebook(userId int64, fbId string, fbToken string) error
	SetIIDToken(userId int64, iidToken *IIDTokenDTO) error
	Delete(user *UserDTO) error
}

type EventTimeLineDAO interface {
	FindAllBackward(date time.Time) ([]*TimeLineEntryDTO, error)
	FindAllForward(date time.Time) ([]*TimeLineEntryDTO, error)
	FindAllBetween(fromDate time.Time, toDate time.Time) ([]*TimeLineEntryDTO, error)
	Insert(item *TimeLineEntryDTO) error
	Delete(item *TimeLineEntryDTO) error
	Replace(oldItem *TimeLineEntryDTO, newItem *TimeLineEntryDTO) error
	DeleteAll() error
}

type EventHistoryDAO interface {
	Insert(userID int64, event *TimeLineEntryDTO) error
	InsertBatch(event *TimeLineEntryDTO, userIDs ...int64) error
	FindAllBackward(userID int64, fromDate time.Time) ([]int64, error)
	FindAllForward(userID int64, fromDate time.Time) ([]int64, error)
	DeleteAll() error
}

type EventDAO interface {
	RangeAll(f func(*EventDTO) error) error
	RangeEvents(f func(*EventDTO) error, event_ids ...int64) error
	LoadEvents(ids ...int64) (events []*EventDTO, err error)
	LoadEventPicture(eventId int64) (*PictureDTO, error)
	LoadParticipant(participantID int64, eventID int64) (*ParticipantDTO, error)
	Insert(event *EventDTO) error
	Replace(oldEvent *EventDTO, newEvent *EventDTO) error
	InsertParticipant(p *ParticipantDTO) error
	SetEventPicture(event_id int64, picture *PictureDTO) error
}

type FriendDAO interface {
	LoadFriends(userId int64, groupId int32) ([]*FriendDTO, error)
	ContainsFriend(userId int64, otherUserId int64) (bool, error)
	LoadGroups(userId int64) ([]*GroupDTO, error)
	LoadGroupsWithMembers(userId int64) ([]*GroupDTO, error)
	SetFriendPictureDigest(userId int64, friendId int64, digest []byte) error
	InsertGroup(userId int64, group *GroupDTO) error
	SetGroupName(user_id int64, groupId int32, name string) error
	AddMembers(userId int64, groupId int32, friendIds ...int64) error
	MakeFriends(user1 *FriendDTO, user2 *FriendDTO) error
	DeleteGroup(userId int64, groupId int32) error
	DeleteMembers(userId int64, groupId int32, friendIds ...int64) error
}

type FriendRequestDAO interface {
	Load(fromUser int64, toUser int64) (*FriendRequestDTO, error)
	LoadAll(user_id int64) ([]*FriendRequestDTO, error)
	Exist(fromUser int64, toUser int64) (bool, error)
	Insert(friendRequest *FriendRequestDTO) error
	Delete(friendRequest *FriendRequestDTO) error
}

type ThumbnailDAO interface {
	Load(id int64, dpi int32) ([]byte, error)
	Insert(id int64, digest []byte, thumbnails map[int32][]byte) error
	Remove(id int64) error
}

type AccessTokenDAO interface {
	Load(userId int64) (*AccessTokenDTO, error)
	Insert(accessToken *AccessTokenDTO) error
	SetLastUsed(user_id int64, time int64) error
	Remove(user_id int64) error
}

type LogDAO interface {
	LogRegisteredUser(userID int64, createdDate int64) error
	LogActiveSession(node int, userIDA int64, lastTime int64) error
	FindActiveSessions(node int, time time.Time) ([]*ActiveSessionInfoDTO, error)
}
