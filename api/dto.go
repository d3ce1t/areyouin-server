package api

type UserDTO struct {
	Id            int64
	Name          string
	Email         string
	EmailVerified bool
	PictureDigest []byte
	IidToken      IIDTokenDTO
	AuthToken     string
	Password      [32]byte
	Salt          [32]byte
	FbId          string
	FbToken       string
	LastConn      int64
	CreatedDate   int64
}

type EventDTO struct {
	Id            int64
	AuthorId      int64
	AuthorName    string
	Description   string
	PictureDigest []byte
	CreatedDate   int64
	InboxPosition int64
	StartDate     int64
	EndDate       int64
	NumAttendees  int32
	NumGuests     int32
	Cancelled     bool
	Participants  map[int64]*ParticipantDTO
}

type ParticipantDTO struct {
	UserId           int64
	Name             string
	Response         AttendanceResponse
	InvitationStatus InvitationStatus
}

type AccessTokenDTO struct {
	UserId      int64
	Token       string
	LastUsed    int64
	CreatedDate int64
}

type GroupDTO struct {
	Id      int32
	Name    string
	Size    int32
	Members []int64
}

type FriendDTO struct {
	UserId        int64
	Name          string
	PictureDigest []byte
}

type FriendRequestDTO struct {
	ToUser      int64
	FromUser    int64
	Name        string
	Email       string
	CreatedDate int64
}

type IIDTokenDTO struct {
	Token   string
	Version int
}

type PictureDTO struct {
	RawData []byte
	Digest  []byte
}
