// Code generated by protoc-gen-go.
// source: protocol.proto
// DO NOT EDIT!

/*
Package protocol is a generated protocol buffer package.

It is generated from these files:
	protocol.proto

It has these top-level messages:
	AyiHeaderV2
	CreateEvent
	CancelEvent
	InviteUsers
	CancelUsersInvitation
	ConfirmAttendance
	ModifyEvent
	VoteChange
	UserPosition
	UserPositionRange
	CreateUserAccount
	NewAuthToken
	UserAuthentication
	InstanceIDToken
	EventCancelled
	EventExpired
	EventModified
	InvitationCancelled
	AttendanceStatus
	EventChangeProposed
	VotingStatus
	ChangeAccepted
	ChangeDiscarded
	AccessGranted
	Ok
	Error
	TimeInfo
	ReadEvent
	ListCursor
	ListPublicEvents
	EventsList
	FriendsList
	UserAccount
*/
package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import common "peeple/areyouin/common"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// NEW AUTH TOKEN
type AuthType int32

const (
	AuthType_A_NATIVE   AuthType = 0
	AuthType_A_FACEBOOK AuthType = 1
)

var AuthType_name = map[int32]string{
	0: "A_NATIVE",
	1: "A_FACEBOOK",
}
var AuthType_value = map[string]int32{
	"A_NATIVE":   0,
	"A_FACEBOOK": 1,
}

func (x AuthType) String() string {
	return proto.EnumName(AuthType_name, int32(x))
}
func (AuthType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// Header
type AyiHeaderV2 struct {
	Version     uint32 `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Token       uint32 `protobuf:"varint,2,opt,name=token" json:"token,omitempty"`
	Type        uint32 `protobuf:"varint,3,opt,name=type" json:"type,omitempty"`
	PayloadSize uint32 `protobuf:"varint,4,opt,name=payloadSize" json:"payloadSize,omitempty"`
}

func (m *AyiHeaderV2) Reset()                    { *m = AyiHeaderV2{} }
func (m *AyiHeaderV2) String() string            { return proto.CompactTextString(m) }
func (*AyiHeaderV2) ProtoMessage()               {}
func (*AyiHeaderV2) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// CREATE EVENT
type CreateEvent struct {
	Message      string   `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
	CreatedDate  int64    `protobuf:"varint,2,opt,name=created_date" json:"created_date,omitempty"`
	StartDate    int64    `protobuf:"varint,3,opt,name=start_date" json:"start_date,omitempty"`
	EndDate      int64    `protobuf:"varint,4,opt,name=end_date" json:"end_date,omitempty"`
	Participants []uint64 `protobuf:"varint,5,rep,name=participants" json:"participants,omitempty"`
}

func (m *CreateEvent) Reset()                    { *m = CreateEvent{} }
func (m *CreateEvent) String() string            { return proto.CompactTextString(m) }
func (*CreateEvent) ProtoMessage()               {}
func (*CreateEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// CANCEL EVENT
type CancelEvent struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Reason  string `protobuf:"bytes,2,opt,name=reason" json:"reason,omitempty"`
}

func (m *CancelEvent) Reset()                    { *m = CancelEvent{} }
func (m *CancelEvent) String() string            { return proto.CompactTextString(m) }
func (*CancelEvent) ProtoMessage()               {}
func (*CancelEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

// INVITE USERS
type InviteUsers struct {
	EventId      uint64   `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Participants []uint64 `protobuf:"varint,2,rep,name=participants" json:"participants,omitempty"`
}

func (m *InviteUsers) Reset()                    { *m = InviteUsers{} }
func (m *InviteUsers) String() string            { return proto.CompactTextString(m) }
func (*InviteUsers) ProtoMessage()               {}
func (*InviteUsers) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// CANCEL USERS INVITATION
type CancelUsersInvitation struct {
	EventId      uint64   `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Participants []uint64 `protobuf:"varint,2,rep,name=participants" json:"participants,omitempty"`
}

func (m *CancelUsersInvitation) Reset()                    { *m = CancelUsersInvitation{} }
func (m *CancelUsersInvitation) String() string            { return proto.CompactTextString(m) }
func (*CancelUsersInvitation) ProtoMessage()               {}
func (*CancelUsersInvitation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

// CONFIRM ATTENDANCE
type ConfirmAttendance struct {
	EventId    uint64                    `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ActionCode common.AttendanceResponse `protobuf:"varint,2,opt,name=action_code,enum=common.AttendanceResponse" json:"action_code,omitempty"`
}

func (m *ConfirmAttendance) Reset()                    { *m = ConfirmAttendance{} }
func (m *ConfirmAttendance) String() string            { return proto.CompactTextString(m) }
func (*ConfirmAttendance) ProtoMessage()               {}
func (*ConfirmAttendance) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// MODIFY EVENT DATE
// MODIFY EVENT MESSAGE
// MODIFY EVENT
type ModifyEvent struct {
	EventId   uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Message   string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	StartDate int64  `protobuf:"varint,3,opt,name=start_date" json:"start_date,omitempty"`
	EndDate   int64  `protobuf:"varint,4,opt,name=end_date" json:"end_date,omitempty"`
}

func (m *ModifyEvent) Reset()                    { *m = ModifyEvent{} }
func (m *ModifyEvent) String() string            { return proto.CompactTextString(m) }
func (*ModifyEvent) ProtoMessage()               {}
func (*ModifyEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

// VOTE CHANGE
type VoteChange struct {
	EventId      uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId     uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
	AcceptChange bool   `protobuf:"varint,3,opt,name=accept_change" json:"accept_change,omitempty"`
}

func (m *VoteChange) Reset()                    { *m = VoteChange{} }
func (m *VoteChange) String() string            { return proto.CompactTextString(m) }
func (*VoteChange) ProtoMessage()               {}
func (*VoteChange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

// USER POSITION
type UserPosition struct {
	GlobalCoordinates *common.Location `protobuf:"bytes,1,opt,name=global_coordinates" json:"global_coordinates,omitempty"`
	EstimationError   float32          `protobuf:"fixed32,2,opt,name=estimation_error" json:"estimation_error,omitempty"`
}

func (m *UserPosition) Reset()                    { *m = UserPosition{} }
func (m *UserPosition) String() string            { return proto.CompactTextString(m) }
func (*UserPosition) ProtoMessage()               {}
func (*UserPosition) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *UserPosition) GetGlobalCoordinates() *common.Location {
	if m != nil {
		return m.GlobalCoordinates
	}
	return nil
}

// USER POSITION RANGE
type UserPositionRange struct {
	RangeInMeters float32 `protobuf:"fixed32,1,opt,name=range_in_meters" json:"range_in_meters,omitempty"`
}

func (m *UserPositionRange) Reset()                    { *m = UserPositionRange{} }
func (m *UserPositionRange) String() string            { return proto.CompactTextString(m) }
func (*UserPositionRange) ProtoMessage()               {}
func (*UserPositionRange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

// CREATE USER ACCOUNT
type CreateUserAccount struct {
	Name     string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Email    string `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	Password string `protobuf:"bytes,3,opt,name=password" json:"password,omitempty"`
	Phone    string `protobuf:"bytes,4,opt,name=phone" json:"phone,omitempty"`
	Fbid     string `protobuf:"bytes,5,opt,name=fbid" json:"fbid,omitempty"`
	Fbtoken  string `protobuf:"bytes,6,opt,name=fbtoken" json:"fbtoken,omitempty"`
}

func (m *CreateUserAccount) Reset()                    { *m = CreateUserAccount{} }
func (m *CreateUserAccount) String() string            { return proto.CompactTextString(m) }
func (*CreateUserAccount) ProtoMessage()               {}
func (*CreateUserAccount) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

type NewAuthToken struct {
	Pass1 string   `protobuf:"bytes,1,opt,name=pass1" json:"pass1,omitempty"`
	Pass2 string   `protobuf:"bytes,2,opt,name=pass2" json:"pass2,omitempty"`
	Type  AuthType `protobuf:"varint,3,opt,name=type,enum=protocol.AuthType" json:"type,omitempty"`
}

func (m *NewAuthToken) Reset()                    { *m = NewAuthToken{} }
func (m *NewAuthToken) String() string            { return proto.CompactTextString(m) }
func (*NewAuthToken) ProtoMessage()               {}
func (*NewAuthToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

// USER AUTH
type UserAuthentication struct {
	UserId    uint64 `protobuf:"varint,1,opt,name=user_id" json:"user_id,omitempty"`
	AuthToken string `protobuf:"bytes,2,opt,name=auth_token" json:"auth_token,omitempty"`
}

func (m *UserAuthentication) Reset()                    { *m = UserAuthentication{} }
func (m *UserAuthentication) String() string            { return proto.CompactTextString(m) }
func (*UserAuthentication) ProtoMessage()               {}
func (*UserAuthentication) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

// INSTANCE ID TOKEN
type InstanceIDToken struct {
	Token string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
}

func (m *InstanceIDToken) Reset()                    { *m = InstanceIDToken{} }
func (m *InstanceIDToken) String() string            { return proto.CompactTextString(m) }
func (*InstanceIDToken) ProtoMessage()               {}
func (*InstanceIDToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

// EVENT CANCELLED
type EventCancelled struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Reason  string `protobuf:"bytes,2,opt,name=reason" json:"reason,omitempty"`
}

func (m *EventCancelled) Reset()                    { *m = EventCancelled{} }
func (m *EventCancelled) String() string            { return proto.CompactTextString(m) }
func (*EventCancelled) ProtoMessage()               {}
func (*EventCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

// EVENT EXPIRED
type EventExpired struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
}

func (m *EventExpired) Reset()                    { *m = EventExpired{} }
func (m *EventExpired) String() string            { return proto.CompactTextString(m) }
func (*EventExpired) ProtoMessage()               {}
func (*EventExpired) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

// EVENT DATE MODIFIED
// EVENT MESSAGE MODIFIED
// EVENT MODIFIED
type EventModified struct {
	EventId   uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	StartDate int64  `protobuf:"varint,2,opt,name=start_date" json:"start_date,omitempty"`
	EndDate   int64  `protobuf:"varint,3,opt,name=end_date" json:"end_date,omitempty"`
	Message   string `protobuf:"bytes,4,opt,name=message" json:"message,omitempty"`
}

func (m *EventModified) Reset()                    { *m = EventModified{} }
func (m *EventModified) String() string            { return proto.CompactTextString(m) }
func (*EventModified) ProtoMessage()               {}
func (*EventModified) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

// INVITATION CANCELLED
type InvitationCancelled struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
}

func (m *InvitationCancelled) Reset()                    { *m = InvitationCancelled{} }
func (m *InvitationCancelled) String() string            { return proto.CompactTextString(m) }
func (*InvitationCancelled) ProtoMessage()               {}
func (*InvitationCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

// ATTENDANCE STATUS
type AttendanceStatus struct {
	EventId          uint64                     `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	AttendanceStatus []*common.EventParticipant `protobuf:"bytes,2,rep,name=attendance_status" json:"attendance_status,omitempty"`
}

func (m *AttendanceStatus) Reset()                    { *m = AttendanceStatus{} }
func (m *AttendanceStatus) String() string            { return proto.CompactTextString(m) }
func (*AttendanceStatus) ProtoMessage()               {}
func (*AttendanceStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

func (m *AttendanceStatus) GetAttendanceStatus() []*common.EventParticipant {
	if m != nil {
		return m.AttendanceStatus
	}
	return nil
}

// EVENT CHANGE DATE PROPOSED
// EVENT CHANGE MESSAGE PROPOSED
// EVENT CHANGE PROPOSED
type EventChangeProposed struct {
	EventId   uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId  uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
	StartDate int64  `protobuf:"varint,3,opt,name=start_date" json:"start_date,omitempty"`
	EndDate   int64  `protobuf:"varint,4,opt,name=end_date" json:"end_date,omitempty"`
	Message   string `protobuf:"bytes,5,opt,name=message" json:"message,omitempty"`
}

func (m *EventChangeProposed) Reset()                    { *m = EventChangeProposed{} }
func (m *EventChangeProposed) String() string            { return proto.CompactTextString(m) }
func (*EventChangeProposed) ProtoMessage()               {}
func (*EventChangeProposed) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

// VOTING STATUS
// VOTING FINISHED
type VotingStatus struct {
	EventId       uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId      uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
	StartDate     int64  `protobuf:"varint,3,opt,name=start_date" json:"start_date,omitempty"`
	EndDate       int64  `protobuf:"varint,4,opt,name=end_date" json:"end_date,omitempty"`
	ElapsedTime   int64  `protobuf:"varint,5,opt,name=elapsed_time" json:"elapsed_time,omitempty"`
	VotesReceived uint32 `protobuf:"varint,6,opt,name=votes_received" json:"votes_received,omitempty"`
	VotesTotal    uint32 `protobuf:"varint,7,opt,name=votes_total" json:"votes_total,omitempty"`
	Finished      bool   `protobuf:"varint,8,opt,name=finished" json:"finished,omitempty"`
}

func (m *VotingStatus) Reset()                    { *m = VotingStatus{} }
func (m *VotingStatus) String() string            { return proto.CompactTextString(m) }
func (*VotingStatus) ProtoMessage()               {}
func (*VotingStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

// CHANGE ACCEPTED
type ChangeAccepted struct {
	EventId  uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
}

func (m *ChangeAccepted) Reset()                    { *m = ChangeAccepted{} }
func (m *ChangeAccepted) String() string            { return proto.CompactTextString(m) }
func (*ChangeAccepted) ProtoMessage()               {}
func (*ChangeAccepted) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

// CHANGE DISCARDED
type ChangeDiscarded struct {
	EventId  uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
}

func (m *ChangeDiscarded) Reset()                    { *m = ChangeDiscarded{} }
func (m *ChangeDiscarded) String() string            { return proto.CompactTextString(m) }
func (*ChangeDiscarded) ProtoMessage()               {}
func (*ChangeDiscarded) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{22} }

// ACCESS GRANTED
type AccessGranted struct {
	UserId    uint64 `protobuf:"varint,1,opt,name=user_id" json:"user_id,omitempty"`
	AuthToken string `protobuf:"bytes,2,opt,name=auth_token" json:"auth_token,omitempty"`
}

func (m *AccessGranted) Reset()                    { *m = AccessGranted{} }
func (m *AccessGranted) String() string            { return proto.CompactTextString(m) }
func (*AccessGranted) ProtoMessage()               {}
func (*AccessGranted) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{23} }

// OK
type Ok struct {
	Type    int32  `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Payload []byte `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *Ok) Reset()                    { *m = Ok{} }
func (m *Ok) String() string            { return proto.CompactTextString(m) }
func (*Ok) ProtoMessage()               {}
func (*Ok) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{24} }

// ERROR
type Error struct {
	Type  int32 `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Error int32 `protobuf:"varint,2,opt,name=error" json:"error,omitempty"`
}

func (m *Error) Reset()                    { *m = Error{} }
func (m *Error) String() string            { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()               {}
func (*Error) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{25} }

// PING/PONG/CLOCK_RESPONSE
type TimeInfo struct {
	CurrentTime int64 `protobuf:"varint,1,opt,name=current_time" json:"current_time,omitempty"`
}

func (m *TimeInfo) Reset()                    { *m = TimeInfo{} }
func (m *TimeInfo) String() string            { return proto.CompactTextString(m) }
func (*TimeInfo) ProtoMessage()               {}
func (*TimeInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{26} }

// READ EVENT
type ReadEvent struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
}

func (m *ReadEvent) Reset()                    { *m = ReadEvent{} }
func (m *ReadEvent) String() string            { return proto.CompactTextString(m) }
func (*ReadEvent) ProtoMessage()               {}
func (*ReadEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{27} }

// LIST AUTHORED EVENTS
// LIST PRIVATE EVENTS
// LIST PUBLIC EVENTS
// HISTORY AUTHORED EVENTS
// HISTORY PRIVATE EVENTS
// HISTORY PUBLIC EVENTS
type ListCursor struct {
	Cursor uint32 `protobuf:"varint,1,opt,name=cursor" json:"cursor,omitempty"`
}

func (m *ListCursor) Reset()                    { *m = ListCursor{} }
func (m *ListCursor) String() string            { return proto.CompactTextString(m) }
func (*ListCursor) ProtoMessage()               {}
func (*ListCursor) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{28} }

type ListPublicEvents struct {
	UserCoordinates *common.Location `protobuf:"bytes,1,opt,name=user_coordinates" json:"user_coordinates,omitempty"`
	RangeInMeters   uint32           `protobuf:"varint,2,opt,name=range_in_meters" json:"range_in_meters,omitempty"`
	Cursor          *ListCursor      `protobuf:"bytes,3,opt,name=cursor" json:"cursor,omitempty"`
}

func (m *ListPublicEvents) Reset()                    { *m = ListPublicEvents{} }
func (m *ListPublicEvents) String() string            { return proto.CompactTextString(m) }
func (*ListPublicEvents) ProtoMessage()               {}
func (*ListPublicEvents) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{29} }

func (m *ListPublicEvents) GetUserCoordinates() *common.Location {
	if m != nil {
		return m.UserCoordinates
	}
	return nil
}

func (m *ListPublicEvents) GetCursor() *ListCursor {
	if m != nil {
		return m.Cursor
	}
	return nil
}

// EVENTS LIST
type EventsList struct {
	Event []*common.Event `protobuf:"bytes,1,rep,name=event" json:"event,omitempty"`
}

func (m *EventsList) Reset()                    { *m = EventsList{} }
func (m *EventsList) String() string            { return proto.CompactTextString(m) }
func (*EventsList) ProtoMessage()               {}
func (*EventsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{30} }

func (m *EventsList) GetEvent() []*common.Event {
	if m != nil {
		return m.Event
	}
	return nil
}

// FRIENDS LIST
type FriendsList struct {
	Friends []*common.Friend `protobuf:"bytes,1,rep,name=friends" json:"friends,omitempty"`
}

func (m *FriendsList) Reset()                    { *m = FriendsList{} }
func (m *FriendsList) String() string            { return proto.CompactTextString(m) }
func (*FriendsList) ProtoMessage()               {}
func (*FriendsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{31} }

func (m *FriendsList) GetFriends() []*common.Friend {
	if m != nil {
		return m.Friends
	}
	return nil
}

// USER ACCOUNT
type UserAccount struct {
	Name          string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Email         string `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	Picture       []byte `protobuf:"bytes,3,opt,name=picture,proto3" json:"picture,omitempty"`
	PictureDigest []byte `protobuf:"bytes,4,opt,name=picture_digest,proto3" json:"picture_digest,omitempty"`
}

func (m *UserAccount) Reset()                    { *m = UserAccount{} }
func (m *UserAccount) String() string            { return proto.CompactTextString(m) }
func (*UserAccount) ProtoMessage()               {}
func (*UserAccount) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{32} }

func init() {
	proto.RegisterType((*AyiHeaderV2)(nil), "protocol.AyiHeaderV2")
	proto.RegisterType((*CreateEvent)(nil), "protocol.CreateEvent")
	proto.RegisterType((*CancelEvent)(nil), "protocol.CancelEvent")
	proto.RegisterType((*InviteUsers)(nil), "protocol.InviteUsers")
	proto.RegisterType((*CancelUsersInvitation)(nil), "protocol.CancelUsersInvitation")
	proto.RegisterType((*ConfirmAttendance)(nil), "protocol.ConfirmAttendance")
	proto.RegisterType((*ModifyEvent)(nil), "protocol.ModifyEvent")
	proto.RegisterType((*VoteChange)(nil), "protocol.VoteChange")
	proto.RegisterType((*UserPosition)(nil), "protocol.UserPosition")
	proto.RegisterType((*UserPositionRange)(nil), "protocol.UserPositionRange")
	proto.RegisterType((*CreateUserAccount)(nil), "protocol.CreateUserAccount")
	proto.RegisterType((*NewAuthToken)(nil), "protocol.NewAuthToken")
	proto.RegisterType((*UserAuthentication)(nil), "protocol.UserAuthentication")
	proto.RegisterType((*InstanceIDToken)(nil), "protocol.InstanceIDToken")
	proto.RegisterType((*EventCancelled)(nil), "protocol.EventCancelled")
	proto.RegisterType((*EventExpired)(nil), "protocol.EventExpired")
	proto.RegisterType((*EventModified)(nil), "protocol.EventModified")
	proto.RegisterType((*InvitationCancelled)(nil), "protocol.InvitationCancelled")
	proto.RegisterType((*AttendanceStatus)(nil), "protocol.AttendanceStatus")
	proto.RegisterType((*EventChangeProposed)(nil), "protocol.EventChangeProposed")
	proto.RegisterType((*VotingStatus)(nil), "protocol.VotingStatus")
	proto.RegisterType((*ChangeAccepted)(nil), "protocol.ChangeAccepted")
	proto.RegisterType((*ChangeDiscarded)(nil), "protocol.ChangeDiscarded")
	proto.RegisterType((*AccessGranted)(nil), "protocol.AccessGranted")
	proto.RegisterType((*Ok)(nil), "protocol.Ok")
	proto.RegisterType((*Error)(nil), "protocol.Error")
	proto.RegisterType((*TimeInfo)(nil), "protocol.TimeInfo")
	proto.RegisterType((*ReadEvent)(nil), "protocol.ReadEvent")
	proto.RegisterType((*ListCursor)(nil), "protocol.ListCursor")
	proto.RegisterType((*ListPublicEvents)(nil), "protocol.ListPublicEvents")
	proto.RegisterType((*EventsList)(nil), "protocol.EventsList")
	proto.RegisterType((*FriendsList)(nil), "protocol.FriendsList")
	proto.RegisterType((*UserAccount)(nil), "protocol.UserAccount")
	proto.RegisterEnum("protocol.AuthType", AuthType_name, AuthType_value)
}

var fileDescriptor0 = []byte{
	// 1019 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x55, 0x6d, 0x6f, 0x1b, 0x45,
	0x10, 0x26, 0x76, 0xdc, 0xda, 0xe3, 0x97, 0x38, 0x97, 0x16, 0xac, 0xd2, 0x8a, 0x68, 0xa9, 0x44,
	0x14, 0x55, 0x46, 0xb8, 0x14, 0x89, 0x4f, 0xc8, 0x4d, 0x5d, 0x6a, 0x51, 0x12, 0x2b, 0x09, 0x11,
	0x7c, 0x3a, 0xad, 0xef, 0xd6, 0xce, 0xaa, 0xe7, 0xdd, 0x63, 0x77, 0xed, 0x62, 0x24, 0xfe, 0x12,
	0xbf, 0x91, 0xd9, 0x59, 0xbf, 0xb5, 0x38, 0xc8, 0xe2, 0x8b, 0xe5, 0x9d, 0x9b, 0x99, 0x67, 0xe6,
	0xd9, 0x67, 0x66, 0xa1, 0x91, 0x1b, 0xed, 0x74, 0xa2, 0xb3, 0x36, 0xfd, 0x89, 0xca, 0xcb, 0xf3,
	0x23, 0x48, 0xb4, 0x11, 0xc1, 0xca, 0x06, 0x50, 0xed, 0xce, 0xe5, 0x1b, 0xc1, 0x53, 0x61, 0x6e,
	0x3a, 0xd1, 0x01, 0xdc, 0x9f, 0x09, 0x63, 0xa5, 0x56, 0xad, 0xbd, 0xe3, 0xbd, 0x93, 0x7a, 0x54,
	0x87, 0x92, 0xd3, 0xef, 0x84, 0x6a, 0x15, 0xe8, 0x58, 0x83, 0x7d, 0x37, 0xcf, 0x45, 0xab, 0x48,
	0xa7, 0x23, 0xa8, 0xe6, 0x7c, 0x9e, 0x69, 0x9e, 0x5e, 0xc9, 0x3f, 0x45, 0x6b, 0xdf, 0x1b, 0x59,
	0x0e, 0xd5, 0x33, 0x23, 0xb8, 0x13, 0xbd, 0x99, 0x50, 0xce, 0x67, 0x9c, 0x08, 0x6b, 0xf9, 0x58,
	0x50, 0xc6, 0x4a, 0xf4, 0x00, 0x6a, 0x09, 0x7d, 0x4f, 0xe3, 0x14, 0x7f, 0x29, 0x71, 0x31, 0x8a,
	0x00, 0xac, 0xe3, 0xc6, 0x05, 0x5b, 0x91, 0x6c, 0x4d, 0x28, 0x0b, 0xb5, 0xf0, 0xda, 0x27, 0x0b,
	0xc6, 0xe6, 0xe8, 0x24, 0x13, 0x99, 0x73, 0xe5, 0x6c, 0xab, 0x74, 0x5c, 0x3c, 0xd9, 0x67, 0x5f,
	0x23, 0x22, 0x57, 0x89, 0xc8, 0x02, 0xa2, 0x0f, 0xf3, 0x7f, 0x62, 0x99, 0x12, 0xe4, 0x7e, 0xd4,
	0x80, 0x7b, 0x88, 0x68, 0x75, 0xe8, 0xa2, 0xc2, 0x5e, 0x40, 0xb5, 0xaf, 0x66, 0xd2, 0x89, 0x5f,
	0x2c, 0x36, 0xbb, 0x25, 0xe0, 0x63, 0x9c, 0x02, 0xe1, 0xfc, 0x00, 0x0f, 0x03, 0x0e, 0x85, 0x51,
	0x06, 0xee, 0x90, 0xaa, 0x9d, 0x13, 0xdc, 0xc0, 0xe1, 0x99, 0x56, 0x23, 0x69, 0x26, 0x5d, 0xe7,
	0xb0, 0x35, 0x9f, 0x6c, 0x4b, 0x30, 0xf6, 0xc3, 0x13, 0x9f, 0x38, 0x4e, 0x74, 0x1a, 0x08, 0x6a,
	0x74, 0x1e, 0xb5, 0x13, 0x3d, 0x99, 0x68, 0xd5, 0x5e, 0x87, 0x5e, 0x0a, 0x9b, 0x6b, 0x65, 0x05,
	0xe6, 0xad, 0xfe, 0xac, 0x53, 0x39, 0x9a, 0xdf, 0x45, 0xc0, 0xc6, 0x25, 0x10, 0x03, 0xbb, 0xd1,
	0xcd, 0xde, 0x00, 0xdc, 0x68, 0x27, 0xce, 0x6e, 0xb9, 0x1a, 0x6f, 0x2b, 0xf4, 0x10, 0x2a, 0x09,
	0x7d, 0xf3, 0xa6, 0x20, 0x90, 0x87, 0x50, 0xe7, 0x49, 0x22, 0x72, 0x17, 0x87, 0x2f, 0x94, 0xbb,
	0x8c, 0x15, 0xd6, 0x3c, 0x69, 0x03, 0x6d, 0x25, 0x31, 0xf6, 0x0c, 0xa2, 0x71, 0xa6, 0x87, 0x3c,
	0xc3, 0x16, 0xb5, 0x49, 0xa5, 0x42, 0x50, 0x4b, 0x59, 0xab, 0x9d, 0xe6, 0xb2, 0xd3, 0xb7, 0x3a,
	0x09, 0xfc, 0xb6, 0xa0, 0x29, 0xac, 0x93, 0x13, 0x3a, 0xc5, 0xc2, 0x18, 0x6d, 0x08, 0xae, 0xc0,
	0x9e, 0xc1, 0xe1, 0x66, 0xde, 0x4b, 0x2a, 0xf4, 0x33, 0x38, 0x30, 0xa1, 0x2a, 0x15, 0x4f, 0x84,
	0xc3, 0xbb, 0xa2, 0xcc, 0x05, 0x94, 0xe6, 0x61, 0x90, 0xa6, 0x8f, 0xe9, 0x26, 0x89, 0x9e, 0x22,
	0x5b, 0x28, 0x69, 0xc5, 0x27, 0x4b, 0x75, 0xa2, 0xde, 0xc5, 0x84, 0xcb, 0x6c, 0xc1, 0x13, 0xf6,
	0x9c, 0x73, 0x6b, 0xdf, 0x63, 0x8d, 0xd4, 0x09, 0x39, 0xe4, 0xb7, 0x5a, 0x05, 0x8a, 0x2a, 0x3e,
	0x7a, 0x34, 0xc4, 0xee, 0x4b, 0x74, 0x42, 0x9e, 0x47, 0xc3, 0x30, 0x2f, 0xf7, 0x48, 0x69, 0xe7,
	0x50, 0x3b, 0x17, 0xef, 0xbb, 0x53, 0x77, 0x7b, 0xed, 0xad, 0x14, 0x8d, 0xf9, 0xbe, 0x59, 0xa3,
	0xf9, 0x63, 0x67, 0x81, 0x76, 0xbc, 0x31, 0x5d, 0x8d, 0x4e, 0xd4, 0x5e, 0x4d, 0x30, 0x25, 0xc0,
	0x2f, 0xec, 0x7b, 0x88, 0xa8, 0x76, 0x3c, 0xe3, 0x4d, 0xc8, 0x05, 0x3f, 0x08, 0x3b, 0x45, 0xeb,
	0xfa, 0x62, 0xf0, 0x7a, 0x39, 0xba, 0xc4, 0xeb, 0xd1, 0xad, 0xb0, 0x63, 0x38, 0xe8, 0x2b, 0xbc,
	0x74, 0x14, 0x4e, 0xff, 0xd5, 0xaa, 0x9a, 0xe0, 0x41, 0xd5, 0xb0, 0x0e, 0x34, 0x48, 0x40, 0x41,
	0xe4, 0x99, 0x48, 0x77, 0x18, 0xa5, 0x63, 0xa8, 0x51, 0x4c, 0xef, 0x8f, 0x5c, 0x9a, 0x6d, 0x11,
	0xec, 0x57, 0xa8, 0x93, 0x07, 0x29, 0x54, 0x6e, 0x4d, 0xfa, 0xa1, 0x1a, 0x0b, 0xff, 0x52, 0x63,
	0xd0, 0xe7, 0x86, 0x88, 0x89, 0x7b, 0xf6, 0x15, 0x1c, 0xad, 0x87, 0xf0, 0x3f, 0x8a, 0x66, 0xbf,
	0x41, 0x73, 0x3d, 0x35, 0x57, 0xe8, 0x3f, 0xdd, 0x36, 0xf4, 0xcf, 0xe1, 0x90, 0xaf, 0xbc, 0x62,
	0x4b, 0x6e, 0x34, 0xb8, 0xd5, 0x4e, 0x6b, 0x29, 0x49, 0xea, 0x64, 0xb0, 0x9e, 0x6c, 0xf6, 0x3b,
	0x1c, 0x05, 0xce, 0x48, 0xed, 0x03, 0xa3, 0x73, 0x6d, 0xb7, 0xf6, 0xb8, 0x65, 0x56, 0x76, 0xdb,
	0x79, 0x1b, 0x6d, 0x93, 0xc8, 0xd8, 0xdf, 0x7b, 0x50, 0xc3, 0xb1, 0x94, 0x6a, 0x7c, 0x67, 0x2b,
	0xff, 0x1b, 0x0c, 0xf7, 0x96, 0xc8, 0x78, 0x8e, 0x2d, 0xc4, 0x38, 0x6f, 0x01, 0xb1, 0x18, 0x7d,
	0x0a, 0x8d, 0x19, 0xee, 0x01, 0x1b, 0x1b, 0x91, 0x08, 0x39, 0x13, 0x29, 0xa9, 0x9b, 0xf6, 0x7f,
	0xb0, 0x3b, 0xed, 0x78, 0xd6, 0xba, 0x4f, 0x46, 0x4c, 0x3a, 0x92, 0x4a, 0xda, 0x5b, 0x74, 0x2b,
	0xd3, 0xf0, 0xbf, 0x80, 0x46, 0xa0, 0xa7, 0x4b, 0x9b, 0x61, 0x47, 0x7a, 0xd8, 0x77, 0x70, 0x10,
	0xc2, 0x5e, 0x49, 0x9b, 0x70, 0x93, 0xee, 0x1a, 0xf7, 0x2d, 0xd4, 0x3d, 0x90, 0xb5, 0x3f, 0xe2,
	0x16, 0xf0, 0x68, 0x3b, 0x8d, 0xc7, 0x97, 0x50, 0xb8, 0x78, 0xb7, 0x7a, 0xdf, 0xbc, 0x5f, 0xc9,
	0x07, 0x2e, 0xde, 0x37, 0x72, 0xaa, 0xb1, 0xa7, 0x50, 0xea, 0xf9, 0xed, 0xf3, 0x91, 0x9f, 0x5f,
	0x1a, 0xab, 0xa5, 0x54, 0xc2, 0x99, 0x28, 0x5f, 0x23, 0x79, 0x7d, 0x35, 0xd2, 0xf4, 0xda, 0x4d,
	0x8d, 0xf1, 0x35, 0x13, 0xa1, 0x7b, 0xb4, 0x58, 0x9f, 0x40, 0xe5, 0x12, 0x9f, 0xdc, 0x3b, 0xd6,
	0x35, 0x7b, 0x0c, 0xf0, 0x56, 0x5a, 0x77, 0x36, 0x35, 0x16, 0xb1, 0x70, 0xe4, 0x12, 0xfa, 0x17,
	0x9e, 0x64, 0xf6, 0x17, 0x34, 0xfd, 0xd7, 0xc1, 0x74, 0x98, 0xc9, 0x84, 0x52, 0xd8, 0xe8, 0x14,
	0x9a, 0xd4, 0xe2, 0x2e, 0xdb, 0x74, 0xcb, 0x7a, 0x0c, 0x12, 0x79, 0xba, 0x02, 0x2a, 0x52, 0xe8,
	0x83, 0xf5, 0x02, 0x5a, 0x97, 0xc3, 0x4e, 0x01, 0x02, 0xa8, 0xb7, 0x45, 0x8f, 0xb1, 0x75, 0x7f,
	0x42, 0x34, 0x3f, 0x28, 0xf5, 0x0f, 0x06, 0x85, 0xb5, 0xa1, 0xfa, 0xda, 0x48, 0xd4, 0x58, 0x70,
	0xfe, 0x02, 0xd7, 0x63, 0x38, 0x2e, 0xdc, 0x1b, 0x4b, 0xf7, 0xe0, 0xc5, 0xae, 0xa0, 0xba, 0xf3,
	0x6a, 0xf6, 0x97, 0x23, 0x13, 0x37, 0x35, 0x41, 0xcd, 0x35, 0xaf, 0xd2, 0x85, 0x21, 0x4e, 0xe5,
	0x18, 0x1f, 0x0c, 0xd2, 0x74, 0xed, 0xf4, 0x04, 0xca, 0xcb, 0xfd, 0x89, 0x19, 0xcb, 0xdd, 0xf8,
	0xbc, 0x7b, 0xdd, 0xbf, 0xe9, 0x35, 0x3f, 0x41, 0x66, 0xa1, 0x1b, 0xbf, 0xee, 0x9e, 0xf5, 0x5e,
	0x5e, 0x5c, 0xfc, 0xd4, 0xdc, 0x7b, 0xf9, 0x04, 0x3e, 0x17, 0xb6, 0x9d, 0x0b, 0x91, 0x67, 0xa2,
	0xcd, 0x8d, 0x98, 0xeb, 0xa9, 0x54, 0x2b, 0x12, 0x86, 0xf7, 0xe8, 0xdf, 0xf3, 0x7f, 0x02, 0x00,
	0x00, 0xff, 0xff, 0x99, 0xf3, 0xe7, 0x13, 0x5a, 0x09, 0x00, 0x00,
}
