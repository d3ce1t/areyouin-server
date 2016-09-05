package api

import "time"

type DbSession interface {
	Connect() error
	IsValid() bool
	Closed() bool
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

type EventDAO interface {
	LoadEvents(ids ...int64) (events []*EventDTO, err error)
	LoadRecentEventsFromUser(userId int64, fromDate int64) ([]*EventDTO, error)
	LoadEventsHistoryFromUser(userId int64, fromDate int64, toDate int64) ([]*EventDTO, error)
	LoadEventPicture(eventId int64) (*PictureDTO, error)
	Insert(event *EventDTO) error
	AddParticipantToEvent(participant *ParticipantDTO, event *EventDTO) error
	SetEventPicture(event_id int64, picture *PictureDTO) error
	SetEventStateAndInboxPosition(eventId int64, newStatus EventState, newPosition int64) error
	SetParticipantInvitationStatus(userId int64, eventId int64, status InvitationStatus) error
	SetParticipantResponse(participant int64, response AttendanceResponse, event *EventDTO) error
	SetNumGuests(eventId int64, numGuests int) (ok bool, err error)
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
