// Code generated by protoc-gen-go.
// source: protocol.proto
// DO NOT EDIT!

/*
Package protocol is a generated protocol buffer package.

It is generated from these files:
	protocol.proto

It has these top-level messages:
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
func (*CreateEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// CANCEL EVENT
type CancelEvent struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Reason  string `protobuf:"bytes,2,opt,name=reason" json:"reason,omitempty"`
}

func (m *CancelEvent) Reset()                    { *m = CancelEvent{} }
func (m *CancelEvent) String() string            { return proto.CompactTextString(m) }
func (*CancelEvent) ProtoMessage()               {}
func (*CancelEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// INVITE USERS
type InviteUsers struct {
	EventId      uint64   `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Participants []uint64 `protobuf:"varint,2,rep,name=participants" json:"participants,omitempty"`
}

func (m *InviteUsers) Reset()                    { *m = InviteUsers{} }
func (m *InviteUsers) String() string            { return proto.CompactTextString(m) }
func (*InviteUsers) ProtoMessage()               {}
func (*InviteUsers) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

// CANCEL USERS INVITATION
type CancelUsersInvitation struct {
	EventId      uint64   `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Participants []uint64 `protobuf:"varint,2,rep,name=participants" json:"participants,omitempty"`
}

func (m *CancelUsersInvitation) Reset()                    { *m = CancelUsersInvitation{} }
func (m *CancelUsersInvitation) String() string            { return proto.CompactTextString(m) }
func (*CancelUsersInvitation) ProtoMessage()               {}
func (*CancelUsersInvitation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// CONFIRM ATTENDANCE
type ConfirmAttendance struct {
	EventId    uint64                    `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ActionCode common.AttendanceResponse `protobuf:"varint,2,opt,name=action_code,enum=common.AttendanceResponse" json:"action_code,omitempty"`
}

func (m *ConfirmAttendance) Reset()                    { *m = ConfirmAttendance{} }
func (m *ConfirmAttendance) String() string            { return proto.CompactTextString(m) }
func (*ConfirmAttendance) ProtoMessage()               {}
func (*ConfirmAttendance) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

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
func (*ModifyEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// VOTE CHANGE
type VoteChange struct {
	EventId      uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId     uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
	AcceptChange bool   `protobuf:"varint,3,opt,name=accept_change" json:"accept_change,omitempty"`
}

func (m *VoteChange) Reset()                    { *m = VoteChange{} }
func (m *VoteChange) String() string            { return proto.CompactTextString(m) }
func (*VoteChange) ProtoMessage()               {}
func (*VoteChange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

// USER POSITION
type UserPosition struct {
	GlobalCoordinates *common.Location `protobuf:"bytes,1,opt,name=global_coordinates" json:"global_coordinates,omitempty"`
	EstimationError   float32          `protobuf:"fixed32,2,opt,name=estimation_error" json:"estimation_error,omitempty"`
}

func (m *UserPosition) Reset()                    { *m = UserPosition{} }
func (m *UserPosition) String() string            { return proto.CompactTextString(m) }
func (*UserPosition) ProtoMessage()               {}
func (*UserPosition) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

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
func (*UserPositionRange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

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
func (*CreateUserAccount) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type NewAuthToken struct {
	Pass1 string   `protobuf:"bytes,1,opt,name=pass1" json:"pass1,omitempty"`
	Pass2 string   `protobuf:"bytes,2,opt,name=pass2" json:"pass2,omitempty"`
	Type  AuthType `protobuf:"varint,3,opt,name=type,enum=protocol.AuthType" json:"type,omitempty"`
}

func (m *NewAuthToken) Reset()                    { *m = NewAuthToken{} }
func (m *NewAuthToken) String() string            { return proto.CompactTextString(m) }
func (*NewAuthToken) ProtoMessage()               {}
func (*NewAuthToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

// USER AUTH
type UserAuthentication struct {
	UserId    uint64 `protobuf:"varint,1,opt,name=user_id" json:"user_id,omitempty"`
	AuthToken string `protobuf:"bytes,2,opt,name=auth_token" json:"auth_token,omitempty"`
}

func (m *UserAuthentication) Reset()                    { *m = UserAuthentication{} }
func (m *UserAuthentication) String() string            { return proto.CompactTextString(m) }
func (*UserAuthentication) ProtoMessage()               {}
func (*UserAuthentication) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

// EVENT CANCELLED
type EventCancelled struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	Reason  string `protobuf:"bytes,2,opt,name=reason" json:"reason,omitempty"`
}

func (m *EventCancelled) Reset()                    { *m = EventCancelled{} }
func (m *EventCancelled) String() string            { return proto.CompactTextString(m) }
func (*EventCancelled) ProtoMessage()               {}
func (*EventCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

// EVENT EXPIRED
type EventExpired struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
}

func (m *EventExpired) Reset()                    { *m = EventExpired{} }
func (m *EventExpired) String() string            { return proto.CompactTextString(m) }
func (*EventExpired) ProtoMessage()               {}
func (*EventExpired) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

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
func (*EventModified) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

// INVITATION CANCELLED
type InvitationCancelled struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
}

func (m *InvitationCancelled) Reset()                    { *m = InvitationCancelled{} }
func (m *InvitationCancelled) String() string            { return proto.CompactTextString(m) }
func (*InvitationCancelled) ProtoMessage()               {}
func (*InvitationCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

// ATTENDANCE STATUS
type AttendanceStatus struct {
	EventId          uint64                     `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	AttendanceStatus []*common.EventParticipant `protobuf:"bytes,2,rep,name=attendance_status" json:"attendance_status,omitempty"`
}

func (m *AttendanceStatus) Reset()                    { *m = AttendanceStatus{} }
func (m *AttendanceStatus) String() string            { return proto.CompactTextString(m) }
func (*AttendanceStatus) ProtoMessage()               {}
func (*AttendanceStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

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
func (*EventChangeProposed) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

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
func (*VotingStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

// CHANGE ACCEPTED
type ChangeAccepted struct {
	EventId  uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
}

func (m *ChangeAccepted) Reset()                    { *m = ChangeAccepted{} }
func (m *ChangeAccepted) String() string            { return proto.CompactTextString(m) }
func (*ChangeAccepted) ProtoMessage()               {}
func (*ChangeAccepted) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

// CHANGE DISCARDED
type ChangeDiscarded struct {
	EventId  uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
	ChangeId uint32 `protobuf:"varint,2,opt,name=change_id" json:"change_id,omitempty"`
}

func (m *ChangeDiscarded) Reset()                    { *m = ChangeDiscarded{} }
func (m *ChangeDiscarded) String() string            { return proto.CompactTextString(m) }
func (*ChangeDiscarded) ProtoMessage()               {}
func (*ChangeDiscarded) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

// ACCESS GRANTED
type AccessGranted struct {
	UserId    uint64 `protobuf:"varint,1,opt,name=user_id" json:"user_id,omitempty"`
	AuthToken string `protobuf:"bytes,2,opt,name=auth_token" json:"auth_token,omitempty"`
}

func (m *AccessGranted) Reset()                    { *m = AccessGranted{} }
func (m *AccessGranted) String() string            { return proto.CompactTextString(m) }
func (*AccessGranted) ProtoMessage()               {}
func (*AccessGranted) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

// OK
type Ok struct {
	Type int32 `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
}

func (m *Ok) Reset()                    { *m = Ok{} }
func (m *Ok) String() string            { return proto.CompactTextString(m) }
func (*Ok) ProtoMessage()               {}
func (*Ok) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{22} }

// ERROR
type Error struct {
	Type  int32 `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Error int32 `protobuf:"varint,2,opt,name=error" json:"error,omitempty"`
}

func (m *Error) Reset()                    { *m = Error{} }
func (m *Error) String() string            { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()               {}
func (*Error) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{23} }

// PING/PONG/CLOCK_RESPONSE
type TimeInfo struct {
	CurrentTime int64 `protobuf:"varint,1,opt,name=current_time" json:"current_time,omitempty"`
}

func (m *TimeInfo) Reset()                    { *m = TimeInfo{} }
func (m *TimeInfo) String() string            { return proto.CompactTextString(m) }
func (*TimeInfo) ProtoMessage()               {}
func (*TimeInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{24} }

// READ EVENT
type ReadEvent struct {
	EventId uint64 `protobuf:"varint,1,opt,name=event_id" json:"event_id,omitempty"`
}

func (m *ReadEvent) Reset()                    { *m = ReadEvent{} }
func (m *ReadEvent) String() string            { return proto.CompactTextString(m) }
func (*ReadEvent) ProtoMessage()               {}
func (*ReadEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{25} }

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
func (*ListCursor) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{26} }

type ListPublicEvents struct {
	UserCoordinates *common.Location `protobuf:"bytes,1,opt,name=user_coordinates" json:"user_coordinates,omitempty"`
	RangeInMeters   uint32           `protobuf:"varint,2,opt,name=range_in_meters" json:"range_in_meters,omitempty"`
	Cursor          *ListCursor      `protobuf:"bytes,3,opt,name=cursor" json:"cursor,omitempty"`
}

func (m *ListPublicEvents) Reset()                    { *m = ListPublicEvents{} }
func (m *ListPublicEvents) String() string            { return proto.CompactTextString(m) }
func (*ListPublicEvents) ProtoMessage()               {}
func (*ListPublicEvents) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{27} }

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
func (*EventsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{28} }

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
func (*FriendsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{29} }

func (m *FriendsList) GetFriends() []*common.Friend {
	if m != nil {
		return m.Friends
	}
	return nil
}

func init() {
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
	proto.RegisterEnum("protocol.AuthType", AuthType_name, AuthType_value)
}

var fileDescriptor0 = []byte{
	// 926 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x55, 0x6d, 0x6f, 0x1b, 0x45,
	0x10, 0xc6, 0x6f, 0xa9, 0x3d, 0x7e, 0x89, 0x7d, 0x69, 0xc1, 0x2a, 0xad, 0x88, 0x4e, 0x95, 0x88,
	0xa2, 0xca, 0x08, 0x97, 0x22, 0xf1, 0x09, 0xb9, 0xc6, 0x85, 0x88, 0x92, 0x44, 0x21, 0x44, 0xf0,
	0xe9, 0xb4, 0xde, 0x5b, 0x27, 0xab, 0x9e, 0x77, 0x8f, 0xdd, 0x75, 0x4a, 0x3e, 0xf0, 0x97, 0xf8,
	0x8d, 0xcc, 0xce, 0xfa, 0x2d, 0x70, 0x41, 0x16, 0x5f, 0xac, 0xdb, 0xd9, 0x99, 0x79, 0x66, 0x9e,
	0x9d, 0x67, 0x0c, 0x9d, 0xdc, 0x68, 0xa7, 0xb9, 0xce, 0x06, 0xf4, 0x11, 0xd5, 0x57, 0xe7, 0xa7,
	0xc0, 0xb5, 0x11, 0xc1, 0x1a, 0xe7, 0xd0, 0x1c, 0x1b, 0xc1, 0x9c, 0x98, 0xdc, 0x0a, 0xe5, 0xa2,
	0x7d, 0x78, 0x34, 0x17, 0xd6, 0xb2, 0x6b, 0xd1, 0x2f, 0x1d, 0x96, 0x8e, 0x1a, 0xd1, 0x63, 0x68,
	0x71, 0xba, 0x4f, 0x93, 0x14, 0x7f, 0xfb, 0x65, 0xb4, 0x56, 0xa2, 0x08, 0xc0, 0x3a, 0x66, 0x5c,
	0xb0, 0x55, 0xc8, 0xd6, 0x85, 0xba, 0x50, 0x4b, 0xaf, 0x2a, 0x59, 0x30, 0x36, 0x47, 0x27, 0xc9,
	0x65, 0xce, 0x94, 0xb3, 0xfd, 0xda, 0x61, 0xe5, 0xa8, 0x1a, 0x7f, 0x81, 0x88, 0x4c, 0x71, 0x91,
	0x05, 0x44, 0x1f, 0xe6, 0x3f, 0x12, 0x99, 0x12, 0x64, 0x35, 0xea, 0xc0, 0x1e, 0x22, 0x5a, 0xad,
	0x08, 0xac, 0x11, 0xbf, 0x86, 0xe6, 0x89, 0xba, 0x95, 0x4e, 0xfc, 0x62, 0x85, 0xb1, 0x05, 0x01,
	0xff, 0xc4, 0x29, 0x13, 0xce, 0xb7, 0xf0, 0x24, 0xe0, 0x50, 0x18, 0x65, 0x60, 0x4e, 0x6a, 0xb5,
	0x73, 0x82, 0x2b, 0xe8, 0x8d, 0xb5, 0x9a, 0x49, 0x33, 0x1f, 0x39, 0x87, 0xad, 0xf9, 0x64, 0x05,
	0xc1, 0xd8, 0x0f, 0xe3, 0x3e, 0x71, 0xc2, 0x75, 0x1a, 0x08, 0xea, 0x0c, 0x9f, 0x0e, 0xb8, 0x9e,
	0xcf, 0xb5, 0x1a, 0x6c, 0x42, 0x2f, 0x84, 0xcd, 0xb5, 0xb2, 0x02, 0xf3, 0x36, 0x7f, 0xd2, 0xa9,
	0x9c, 0xdd, 0x3d, 0x44, 0xc0, 0xd6, 0x23, 0x10, 0x03, 0xbb, 0xd1, 0x1d, 0xff, 0x00, 0x70, 0xa5,
	0x9d, 0x18, 0xdf, 0x30, 0x75, 0x5d, 0x54, 0x68, 0x0f, 0x1a, 0x9c, 0xee, 0xbc, 0xc9, 0x27, 0x6e,
	0x47, 0x4f, 0xa0, 0xcd, 0x38, 0x17, 0xb9, 0x4b, 0xc2, 0x0d, 0xe5, 0xae, 0x63, 0x85, 0x2d, 0x4f,
	0xda, 0xb9, 0xb6, 0x92, 0x18, 0x7b, 0x09, 0xd1, 0x75, 0xa6, 0xa7, 0x2c, 0xc3, 0x16, 0xb5, 0x49,
	0xa5, 0x42, 0x50, 0x4b, 0x59, 0x9b, 0xc3, 0xee, 0xaa, 0xd3, 0x77, 0x9a, 0x07, 0x7e, 0xfb, 0xd0,
	0x15, 0xd6, 0xc9, 0x39, 0x9d, 0x12, 0x61, 0x8c, 0x36, 0x04, 0x57, 0x8e, 0x5f, 0x42, 0x6f, 0x3b,
	0xef, 0x05, 0x15, 0xfa, 0x09, 0xec, 0x9b, 0x50, 0x95, 0x4a, 0xe6, 0xc2, 0xe1, 0x5b, 0x51, 0xe6,
	0x32, 0x8e, 0x66, 0x2f, 0x8c, 0xa6, 0x8f, 0x19, 0x71, 0xae, 0x17, 0xc8, 0x56, 0x0b, 0xaa, 0x8a,
	0xcd, 0x57, 0xd3, 0xd9, 0x86, 0x9a, 0x98, 0x33, 0x99, 0x2d, 0x79, 0xc2, 0x9e, 0x73, 0x66, 0xed,
	0x07, 0xac, 0x91, 0x3a, 0x21, 0x87, 0xfc, 0x46, 0xab, 0x40, 0x51, 0xc3, 0x47, 0xcf, 0xa6, 0xd8,
	0x7d, 0x8d, 0x4e, 0xc8, 0xf3, 0x6c, 0xea, 0xf4, 0x7b, 0xa1, 0xfa, 0x7b, 0x34, 0x69, 0xa7, 0xd0,
	0x3a, 0x15, 0x1f, 0x46, 0x0b, 0x77, 0x73, 0xe9, 0xad, 0x14, 0x8d, 0xf9, 0xbe, 0xdc, 0xa0, 0xf9,
	0xe3, 0x70, 0x89, 0x76, 0x08, 0x55, 0x77, 0x97, 0x07, 0xce, 0x3a, 0xc3, 0x68, 0xb0, 0xd6, 0x1b,
	0x25, 0xc0, 0x9b, 0xf8, 0x1b, 0x88, 0xa8, 0x76, 0x3c, 0xe3, 0x4b, 0xc8, 0x25, 0x3f, 0x08, 0xbb,
	0x40, 0xeb, 0xe6, 0x61, 0xf0, 0x79, 0x19, 0xba, 0x24, 0xa1, 0x94, 0x30, 0xf4, 0x43, 0xe8, 0xd0,
	0x78, 0x84, 0x11, 0xce, 0x44, 0xba, 0x83, 0x50, 0x0e, 0xa1, 0x45, 0x31, 0x93, 0x3f, 0x72, 0x69,
	0x8a, 0x22, 0xe2, 0x5f, 0xa1, 0x4d, 0x1e, 0x34, 0x7f, 0xb2, 0x30, 0xe9, 0xfd, 0x59, 0x2b, 0xff,
	0x6b, 0xd6, 0xc2, 0xf4, 0x6d, 0x8d, 0x28, 0x31, 0x1b, 0x7f, 0x0e, 0x07, 0x1b, 0x89, 0xfd, 0x47,
	0xd1, 0xf1, 0x6f, 0xd0, 0xdd, 0x68, 0xe2, 0x67, 0xf4, 0x5f, 0x14, 0x49, 0xfa, 0x15, 0xf4, 0xd8,
	0xda, 0x2b, 0xb1, 0xe4, 0x46, 0xb2, 0x6c, 0x0e, 0xfb, 0xab, 0x81, 0xa3, 0x4e, 0xce, 0x37, 0xba,
	0x8d, 0x7f, 0x87, 0x83, 0xc0, 0x19, 0xcd, 0xf2, 0xb9, 0xd1, 0xb9, 0xb6, 0x85, 0x3d, 0x16, 0x28,
	0x61, 0xb7, 0x8d, 0xb6, 0xd5, 0x36, 0x8d, 0x50, 0xfc, 0x57, 0x09, 0x5a, 0x28, 0x3a, 0xa9, 0xae,
	0x1f, 0x6c, 0xe5, 0x7f, 0x83, 0xe1, 0x56, 0x12, 0x19, 0xcb, 0xb1, 0x85, 0x04, 0xd5, 0x14, 0x10,
	0x2b, 0xd1, 0xc7, 0xd0, 0xb9, 0x45, 0x95, 0xdb, 0xc4, 0x08, 0x2e, 0xe4, 0xad, 0x48, 0x69, 0x76,
	0xdb, 0xd1, 0x01, 0x34, 0x83, 0xdd, 0x69, 0xc7, 0xb2, 0xfe, 0x23, 0x32, 0x62, 0xd2, 0x99, 0x54,
	0xd2, 0xde, 0xa0, 0x5b, 0x9d, 0xa4, 0xfd, 0x1a, 0x3a, 0x81, 0x9e, 0x11, 0xe9, 0x7e, 0x47, 0x7a,
	0xe2, 0xaf, 0x61, 0x3f, 0x84, 0x7d, 0x27, 0x2d, 0x67, 0x26, 0xdd, 0x35, 0xee, 0x2b, 0x68, 0x7b,
	0x20, 0x6b, 0xbf, 0x47, 0x8d, 0x7b, 0xb4, 0x9d, 0x86, 0x3f, 0x82, 0xf2, 0xd9, 0x7b, 0x2f, 0x56,
	0xd2, 0x97, 0xf7, 0xab, 0xc5, 0x2f, 0xa0, 0x36, 0xf1, 0xab, 0xe4, 0xbe, 0x99, 0x36, 0xc0, 0x7a,
	0xc3, 0xd4, 0x50, 0x02, 0xf5, 0x4b, 0xe4, 0xea, 0x44, 0xcd, 0x34, 0xfd, 0x75, 0x2d, 0x8c, 0xf1,
	0x25, 0x12, 0x7f, 0x25, 0xda, 0x92, 0xcf, 0xa1, 0x71, 0x21, 0x58, 0xfa, 0xc0, 0xee, 0x8d, 0x9f,
	0x01, 0xbc, 0x93, 0xd6, 0x8d, 0x17, 0xc6, 0x22, 0x16, 0x2a, 0x8c, 0xd3, 0x17, 0xdd, 0xb6, 0xe3,
	0x3f, 0xa1, 0xeb, 0x6f, 0xcf, 0x17, 0xd3, 0x4c, 0x72, 0x4a, 0x61, 0xa3, 0x63, 0xe8, 0x52, 0x47,
	0xbb, 0xac, 0xc6, 0x82, 0x5d, 0x17, 0x26, 0xe2, 0xc5, 0x1a, 0xa8, 0x42, 0xa1, 0x8f, 0x37, 0xdb,
	0x64, 0x53, 0x4e, 0x7c, 0x0c, 0x10, 0x40, 0xbd, 0x2d, 0x7a, 0x86, 0xad, 0xfb, 0x13, 0xa2, 0x79,
	0x5d, 0xb4, 0xef, 0xe9, 0x22, 0x1e, 0x40, 0xf3, 0xad, 0x91, 0x38, 0x52, 0xc1, 0xf9, 0x33, 0xdc,
	0x75, 0xe1, 0xb8, 0x74, 0xef, 0xac, 0xdc, 0x83, 0xd7, 0xf1, 0x11, 0xd4, 0x57, 0x7b, 0x0b, 0x29,
	0xae, 0x8f, 0x92, 0xd3, 0xd1, 0xe5, 0xc9, 0xd5, 0xa4, 0xfb, 0x11, 0x92, 0x00, 0xa3, 0xe4, 0xed,
	0x68, 0x3c, 0x79, 0x73, 0x76, 0xf6, 0x63, 0xb7, 0xf4, 0xe6, 0x39, 0x7c, 0x2a, 0xec, 0x20, 0x17,
	0x22, 0xcf, 0xc4, 0x80, 0x19, 0x71, 0xa7, 0x17, 0x52, 0xad, 0xeb, 0x9d, 0xee, 0xd1, 0xd7, 0xab,
	0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0x04, 0x2a, 0x22, 0x94, 0x80, 0x08, 0x00, 0x00,
}
