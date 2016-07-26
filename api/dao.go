package api

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
	//SetPassword(userId int64, password [32]byte, salt [32]byte) (ok bool, err error)
	SaveProfilePicture(userId int64, picture *PictureDTO) error
	SetLastConnection(userId int64, time int64) error
	SetAuthToken(userId int64, auth_token string) error
	SetFacebookToken(userId int64, fbId string, fb_token string) error
	SetIIDToken(userId int64, iidToken *IIDTokenDTO) error
	Delete(user *UserDTO) error

	/*CheckValidAccountObject(user_id int64, email string, fb_id string, check_credentials bool) (bool, error)
	CheckValidAccount(user_id int64, check_credentials bool) (bool, error)
	GetIDByEmailAndPassword(email string, password string) (int64, error)
	GetIDByFacebookID(fb_id string) (int64, error)*/

	/*SetAuthTokenAndFBToken(user_id int64, auth_token string, fb_id string, fb_token string) error
	ResetEmailCredentialPassword(user_id int64, email string, password string) (ok bool, err error)*/

	/*DeleteUserAccount(user_id int64) error
	DeleteEmailCredentials(email string) error
	DeleteFacebookCredentials(fb_id string) error*/
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
	LoadAll(user_id int64) ([]*FriendRequestDTO, error)
	Exist(toUser int64, fromUser int64) (bool, error)
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
