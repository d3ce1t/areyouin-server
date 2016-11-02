package api

import (
	"bytes"
	"fmt"
	"time"
)

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
	Cancelled     bool
	Timestamp     int64 // Microseconds
	Participants  map[int64]*ParticipantDTO
}

func EqualEventDTO(a *EventDTO, b *EventDTO) bool {

	if a.Timestamp != b.Timestamp || a.Id != b.Id ||
		a.AuthorId != b.AuthorId || a.AuthorName != b.AuthorName || a.Description != b.Description ||
		a.CreatedDate != b.CreatedDate || a.StartDate != b.StartDate || a.EndDate != b.EndDate ||
		a.InboxPosition != b.InboxPosition || a.Cancelled != b.Cancelled ||
		!bytes.Equal(a.PictureDigest, b.PictureDigest) || len(a.Participants) != len(b.Participants) {
		return false
	}

	for pID, aParticipant := range a.Participants {

		bParticipant, ok := b.Participants[pID]
		if !ok {
			return false
		}

		if !EqualParticipantDTO(aParticipant, bParticipant) {
			return false
		}
	}

	return true

}

func (d EventDTO) String() string {
	return fmt.Sprintf("%v", d)
}

func (d *EventDTO) Clone() *EventDTO {

	copy := new(EventDTO)
	*copy = *d
	copy.Participants = make(map[int64]*ParticipantDTO)

	for pID, p := range d.Participants {
		copy.Participants[pID] = p.Clone()
	}

	return copy
}

type TimeLineEntryDTO struct {
	EventID  int64
	Position time.Time
}

type TimeLineByEndDate []*TimeLineEntryDTO

func (a TimeLineByEndDate) Len() int           { return len(a) }
func (a TimeLineByEndDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TimeLineByEndDate) Less(i, j int) bool { return a[i].Position.Before(a[j].Position) }

type ParticipantDTO struct {
	UserID           int64
	EventID          int64
	Name             string
	Response         AttendanceResponse
	InvitationStatus InvitationStatus
	NameTS           int64 // Microseconds
	ResponseTS       int64 // Microseconds
	StatusTS         int64 // Microseconds
}

func EqualParticipantDTO(a *ParticipantDTO, b *ParticipantDTO) bool {
	if a.UserID != b.UserID || a.EventID != b.EventID || a.Name != b.Name ||
		a.Response != b.Response || a.InvitationStatus != b.InvitationStatus ||
		a.NameTS != b.NameTS || a.ResponseTS != b.ResponseTS || a.StatusTS != b.StatusTS {
		return false
	}

	return true
}

func (d *ParticipantDTO) Clone() *ParticipantDTO {
	copy := new(ParticipantDTO)
	*copy = *d
	return copy
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
	Token    string
	Version  int
	Platform string
}

type PictureDTO struct {
	RawData []byte
	Digest  []byte
}

type ActiveSessionInfoDTO struct {
	Node     int
	UserID   int64
	LastTime int64
}
