// Code generated by protoc-gen-go.
// source: protocol.proto
// DO NOT EDIT!

/*
Package protocol is a generated protocol buffer package.

It is generated from these files:
	protocol.proto

It has these top-level messages:
	AyiHeaderV2
	Hello
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
	LinkAccount
	NewAuthToken
	AccessToken
	InstanceIDToken
	SyncGroups
	CreateFriendRequest
	ConfirmFriendRequest
	EventCancelled
	EventExpired
	InvitationCancelled
	AttendanceStatus
	EventChangeProposed
	VotingStatus
	ChangeAccepted
	ChangeDiscarded
	Ok
	Error
	TimeInfo
	ReadEvent
	EventListRequest
	EventsList
	FriendsList
	GroupsList
	FriendRequestsList
*/
package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import core "peeple/areyouin/protocol/core"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

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

type ConfirmFriendRequest_FriendRequestResponse int32

const (
	ConfirmFriendRequest_CANCEL  ConfirmFriendRequest_FriendRequestResponse = 0
	ConfirmFriendRequest_CONFIRM ConfirmFriendRequest_FriendRequestResponse = 1
)

var ConfirmFriendRequest_FriendRequestResponse_name = map[int32]string{
	0: "CANCEL",
	1: "CONFIRM",
}
var ConfirmFriendRequest_FriendRequestResponse_value = map[string]int32{
	"CANCEL":  0,
	"CONFIRM": 1,
}

func (x ConfirmFriendRequest_FriendRequestResponse) String() string {
	return proto.EnumName(ConfirmFriendRequest_FriendRequestResponse_name, int32(x))
}
func (ConfirmFriendRequest_FriendRequestResponse) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{18, 0}
}

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

// Hello
type Hello struct {
	ProtocolVersion uint32 `protobuf:"varint,1,opt,name=protocol_version,json=protocolVersion" json:"protocol_version,omitempty"`
	ClientVersion   string `protobuf:"bytes,2,opt,name=client_version,json=clientVersion" json:"client_version,omitempty"`
	Platform        string `protobuf:"bytes,3,opt,name=platform" json:"platform,omitempty"`
	PlatformVersion string `protobuf:"bytes,4,opt,name=platform_version,json=platformVersion" json:"platform_version,omitempty"`
}

func (m *Hello) Reset()                    { *m = Hello{} }
func (m *Hello) String() string            { return proto.CompactTextString(m) }
func (*Hello) ProtoMessage()               {}
func (*Hello) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// CREATE EVENT
type CreateEvent struct {
	Message      string  `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
	CreatedDate  int64   `protobuf:"varint,2,opt,name=created_date,json=createdDate" json:"created_date,omitempty"`
	StartDate    int64   `protobuf:"varint,3,opt,name=start_date,json=startDate" json:"start_date,omitempty"`
	EndDate      int64   `protobuf:"varint,4,opt,name=end_date,json=endDate" json:"end_date,omitempty"`
	Participants []int64 `protobuf:"varint,5,rep,packed,name=participants" json:"participants,omitempty"`
	Picture      []byte  `protobuf:"bytes,6,opt,name=picture,proto3" json:"picture,omitempty"`
}

func (m *CreateEvent) Reset()                    { *m = CreateEvent{} }
func (m *CreateEvent) String() string            { return proto.CompactTextString(m) }
func (*CreateEvent) ProtoMessage()               {}
func (*CreateEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

// CANCEL EVENT
type CancelEvent struct {
	EventId int64  `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	Reason  string `protobuf:"bytes,2,opt,name=reason" json:"reason,omitempty"`
}

func (m *CancelEvent) Reset()                    { *m = CancelEvent{} }
func (m *CancelEvent) String() string            { return proto.CompactTextString(m) }
func (*CancelEvent) ProtoMessage()               {}
func (*CancelEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// INVITE USERS
type InviteUsers struct {
	EventId      int64   `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	Participants []int64 `protobuf:"varint,2,rep,packed,name=participants" json:"participants,omitempty"`
}

func (m *InviteUsers) Reset()                    { *m = InviteUsers{} }
func (m *InviteUsers) String() string            { return proto.CompactTextString(m) }
func (*InviteUsers) ProtoMessage()               {}
func (*InviteUsers) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

// CANCEL USERS INVITATION
type CancelUsersInvitation struct {
	EventId      int64   `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	Participants []int64 `protobuf:"varint,2,rep,packed,name=participants" json:"participants,omitempty"`
}

func (m *CancelUsersInvitation) Reset()                    { *m = CancelUsersInvitation{} }
func (m *CancelUsersInvitation) String() string            { return proto.CompactTextString(m) }
func (*CancelUsersInvitation) ProtoMessage()               {}
func (*CancelUsersInvitation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// CONFIRM ATTENDANCE
type ConfirmAttendance struct {
	EventId    int64                   `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	ActionCode core.AttendanceResponse `protobuf:"varint,2,opt,name=action_code,json=actionCode,enum=core.AttendanceResponse" json:"action_code,omitempty"`
}

func (m *ConfirmAttendance) Reset()                    { *m = ConfirmAttendance{} }
func (m *ConfirmAttendance) String() string            { return proto.CompactTextString(m) }
func (*ConfirmAttendance) ProtoMessage()               {}
func (*ConfirmAttendance) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

// MODIFY EVENT DATE
// MODIFY EVENT MESSAGE
// MODIFY EVENT
type ModifyEvent struct {
	EventId   int64  `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	Message   string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	StartDate int64  `protobuf:"varint,3,opt,name=start_date,json=startDate" json:"start_date,omitempty"`
	EndDate   int64  `protobuf:"varint,4,opt,name=end_date,json=endDate" json:"end_date,omitempty"`
	Picture   []byte `protobuf:"bytes,5,opt,name=picture,proto3" json:"picture,omitempty"`
}

func (m *ModifyEvent) Reset()                    { *m = ModifyEvent{} }
func (m *ModifyEvent) String() string            { return proto.CompactTextString(m) }
func (*ModifyEvent) ProtoMessage()               {}
func (*ModifyEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

// VOTE CHANGE
type VoteChange struct {
	EventId      int64 `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	ChangeId     int32 `protobuf:"varint,2,opt,name=change_id,json=changeId" json:"change_id,omitempty"`
	AcceptChange bool  `protobuf:"varint,3,opt,name=accept_change,json=acceptChange" json:"accept_change,omitempty"`
}

func (m *VoteChange) Reset()                    { *m = VoteChange{} }
func (m *VoteChange) String() string            { return proto.CompactTextString(m) }
func (*VoteChange) ProtoMessage()               {}
func (*VoteChange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

// USER POSITION
type UserPosition struct {
	GlobalCoordinates *core.Location `protobuf:"bytes,1,opt,name=global_coordinates,json=globalCoordinates" json:"global_coordinates,omitempty"`
	EstimationError   float32        `protobuf:"fixed32,2,opt,name=estimation_error,json=estimationError" json:"estimation_error,omitempty"`
}

func (m *UserPosition) Reset()                    { *m = UserPosition{} }
func (m *UserPosition) String() string            { return proto.CompactTextString(m) }
func (*UserPosition) ProtoMessage()               {}
func (*UserPosition) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *UserPosition) GetGlobalCoordinates() *core.Location {
	if m != nil {
		return m.GlobalCoordinates
	}
	return nil
}

// USER POSITION RANGE
type UserPositionRange struct {
	RangeInMeters float32 `protobuf:"fixed32,1,opt,name=range_in_meters,json=rangeInMeters" json:"range_in_meters,omitempty"`
}

func (m *UserPositionRange) Reset()                    { *m = UserPositionRange{} }
func (m *UserPositionRange) String() string            { return proto.CompactTextString(m) }
func (*UserPositionRange) ProtoMessage()               {}
func (*UserPositionRange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

// CREATE USER ACCOUNT
type CreateUserAccount struct {
	Name     string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Email    string `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	Password string `protobuf:"bytes,3,opt,name=password" json:"password,omitempty"`
	Phone    string `protobuf:"bytes,4,opt,name=phone" json:"phone,omitempty"`
	Fbid     string `protobuf:"bytes,5,opt,name=fbid" json:"fbid,omitempty"`
	Fbtoken  string `protobuf:"bytes,6,opt,name=fbtoken" json:"fbtoken,omitempty"`
	Picture  []byte `protobuf:"bytes,7,opt,name=picture,proto3" json:"picture,omitempty"`
}

func (m *CreateUserAccount) Reset()                    { *m = CreateUserAccount{} }
func (m *CreateUserAccount) String() string            { return proto.CompactTextString(m) }
func (*CreateUserAccount) ProtoMessage()               {}
func (*CreateUserAccount) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

// LINK ACCOUNT
type LinkAccount struct {
	UserId       int64                    `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	Provider     core.AccountProviderType `protobuf:"varint,2,opt,name=provider,enum=core.AccountProviderType" json:"provider,omitempty"`
	AccountId    string                   `protobuf:"bytes,3,opt,name=account_id,json=accountId" json:"account_id,omitempty"`
	AccountToken string                   `protobuf:"bytes,4,opt,name=account_token,json=accountToken" json:"account_token,omitempty"`
}

func (m *LinkAccount) Reset()                    { *m = LinkAccount{} }
func (m *LinkAccount) String() string            { return proto.CompactTextString(m) }
func (*LinkAccount) ProtoMessage()               {}
func (*LinkAccount) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

type NewAuthToken struct {
	Pass1 string   `protobuf:"bytes,1,opt,name=pass1" json:"pass1,omitempty"`
	Pass2 string   `protobuf:"bytes,2,opt,name=pass2" json:"pass2,omitempty"`
	Type  AuthType `protobuf:"varint,3,opt,name=type,enum=protocol.AuthType" json:"type,omitempty"`
}

func (m *NewAuthToken) Reset()                    { *m = NewAuthToken{} }
func (m *NewAuthToken) String() string            { return proto.CompactTextString(m) }
func (*NewAuthToken) ProtoMessage()               {}
func (*NewAuthToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

// ACCESS GRANTED / USER AUTH / GET ACCESS TOKEN
type AccessToken struct {
	UserId    int64  `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	AuthToken string `protobuf:"bytes,2,opt,name=auth_token,json=authToken" json:"auth_token,omitempty"`
}

func (m *AccessToken) Reset()                    { *m = AccessToken{} }
func (m *AccessToken) String() string            { return proto.CompactTextString(m) }
func (*AccessToken) ProtoMessage()               {}
func (*AccessToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

// INSTANCE ID TOKEN
type InstanceIDToken struct {
	Token string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
}

func (m *InstanceIDToken) Reset()                    { *m = InstanceIDToken{} }
func (m *InstanceIDToken) String() string            { return proto.CompactTextString(m) }
func (*InstanceIDToken) ProtoMessage()               {}
func (*InstanceIDToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

// SYNC GROUPS
type SyncGroups struct {
	Owner         int64              `protobuf:"varint,1,opt,name=owner" json:"owner,omitempty"`
	Groups        []*core.Group      `protobuf:"bytes,2,rep,name=groups" json:"groups,omitempty"`
	SyncBehaviour core.SyncBehaviour `protobuf:"varint,3,opt,name=sync_behaviour,json=syncBehaviour,enum=core.SyncBehaviour" json:"sync_behaviour,omitempty"`
}

func (m *SyncGroups) Reset()                    { *m = SyncGroups{} }
func (m *SyncGroups) String() string            { return proto.CompactTextString(m) }
func (*SyncGroups) ProtoMessage()               {}
func (*SyncGroups) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *SyncGroups) GetGroups() []*core.Group {
	if m != nil {
		return m.Groups
	}
	return nil
}

// CREATE FRIEND REQUEST
type CreateFriendRequest struct {
	Email string `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
}

func (m *CreateFriendRequest) Reset()                    { *m = CreateFriendRequest{} }
func (m *CreateFriendRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateFriendRequest) ProtoMessage()               {}
func (*CreateFriendRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

// CONFIRM FRIEND REQUEST
type ConfirmFriendRequest struct {
	FriendId int64                                      `protobuf:"varint,1,opt,name=friend_id,json=friendId" json:"friend_id,omitempty"`
	Response ConfirmFriendRequest_FriendRequestResponse `protobuf:"varint,2,opt,name=response,enum=protocol.ConfirmFriendRequest_FriendRequestResponse" json:"response,omitempty"`
}

func (m *ConfirmFriendRequest) Reset()                    { *m = ConfirmFriendRequest{} }
func (m *ConfirmFriendRequest) String() string            { return proto.CompactTextString(m) }
func (*ConfirmFriendRequest) ProtoMessage()               {}
func (*ConfirmFriendRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

// EVENT CANCELLED
type EventCancelled struct {
	WhoId   int64       `protobuf:"varint,1,opt,name=who_id,json=whoId" json:"who_id,omitempty"`
	EventId int64       `protobuf:"varint,2,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	Reason  string      `protobuf:"bytes,3,opt,name=reason" json:"reason,omitempty"`
	Event   *core.Event `protobuf:"bytes,4,opt,name=event" json:"event,omitempty"`
}

func (m *EventCancelled) Reset()                    { *m = EventCancelled{} }
func (m *EventCancelled) String() string            { return proto.CompactTextString(m) }
func (*EventCancelled) ProtoMessage()               {}
func (*EventCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

func (m *EventCancelled) GetEvent() *core.Event {
	if m != nil {
		return m.Event
	}
	return nil
}

// EVENT EXPIRED
type EventExpired struct {
	EventId int64 `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
}

func (m *EventExpired) Reset()                    { *m = EventExpired{} }
func (m *EventExpired) String() string            { return proto.CompactTextString(m) }
func (*EventExpired) ProtoMessage()               {}
func (*EventExpired) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

// INVITATION CANCELLED
type InvitationCancelled struct {
	EventId int64 `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
}

func (m *InvitationCancelled) Reset()                    { *m = InvitationCancelled{} }
func (m *InvitationCancelled) String() string            { return proto.CompactTextString(m) }
func (*InvitationCancelled) ProtoMessage()               {}
func (*InvitationCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

// ATTENDANCE STATUS
type AttendanceStatus struct {
	EventId          int64                    `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	AttendanceStatus []*core.EventParticipant `protobuf:"bytes,2,rep,name=attendance_status,json=attendanceStatus" json:"attendance_status,omitempty"`
	NumGuests        int32                    `protobuf:"varint,3,opt,name=num_guests,json=numGuests" json:"num_guests,omitempty"`
}

func (m *AttendanceStatus) Reset()                    { *m = AttendanceStatus{} }
func (m *AttendanceStatus) String() string            { return proto.CompactTextString(m) }
func (*AttendanceStatus) ProtoMessage()               {}
func (*AttendanceStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{22} }

func (m *AttendanceStatus) GetAttendanceStatus() []*core.EventParticipant {
	if m != nil {
		return m.AttendanceStatus
	}
	return nil
}

// EVENT CHANGE DATE PROPOSED
// EVENT CHANGE MESSAGE PROPOSED
// EVENT CHANGE PROPOSED
type EventChangeProposed struct {
	EventId   int64  `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	ChangeId  int32  `protobuf:"varint,2,opt,name=change_id,json=changeId" json:"change_id,omitempty"`
	StartDate int64  `protobuf:"varint,3,opt,name=start_date,json=startDate" json:"start_date,omitempty"`
	EndDate   int64  `protobuf:"varint,4,opt,name=end_date,json=endDate" json:"end_date,omitempty"`
	Message   string `protobuf:"bytes,5,opt,name=message" json:"message,omitempty"`
}

func (m *EventChangeProposed) Reset()                    { *m = EventChangeProposed{} }
func (m *EventChangeProposed) String() string            { return proto.CompactTextString(m) }
func (*EventChangeProposed) ProtoMessage()               {}
func (*EventChangeProposed) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{23} }

// VOTING STATUS
// VOTING FINISHED
type VotingStatus struct {
	EventId       int64  `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	ChangeId      int32  `protobuf:"varint,2,opt,name=change_id,json=changeId" json:"change_id,omitempty"`
	StartDate     int64  `protobuf:"varint,3,opt,name=start_date,json=startDate" json:"start_date,omitempty"`
	EndDate       int64  `protobuf:"varint,4,opt,name=end_date,json=endDate" json:"end_date,omitempty"`
	ElapsedTime   int64  `protobuf:"varint,5,opt,name=elapsed_time,json=elapsedTime" json:"elapsed_time,omitempty"`
	VotesReceived uint32 `protobuf:"varint,6,opt,name=votes_received,json=votesReceived" json:"votes_received,omitempty"`
	VotesTotal    uint32 `protobuf:"varint,7,opt,name=votes_total,json=votesTotal" json:"votes_total,omitempty"`
	Finished      bool   `protobuf:"varint,8,opt,name=finished" json:"finished,omitempty"`
}

func (m *VotingStatus) Reset()                    { *m = VotingStatus{} }
func (m *VotingStatus) String() string            { return proto.CompactTextString(m) }
func (*VotingStatus) ProtoMessage()               {}
func (*VotingStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{24} }

// CHANGE ACCEPTED
type ChangeAccepted struct {
	EventId  int64 `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	ChangeId int32 `protobuf:"varint,2,opt,name=change_id,json=changeId" json:"change_id,omitempty"`
}

func (m *ChangeAccepted) Reset()                    { *m = ChangeAccepted{} }
func (m *ChangeAccepted) String() string            { return proto.CompactTextString(m) }
func (*ChangeAccepted) ProtoMessage()               {}
func (*ChangeAccepted) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{25} }

// CHANGE DISCARDED
type ChangeDiscarded struct {
	EventId  int64 `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	ChangeId int32 `protobuf:"varint,2,opt,name=change_id,json=changeId" json:"change_id,omitempty"`
}

func (m *ChangeDiscarded) Reset()                    { *m = ChangeDiscarded{} }
func (m *ChangeDiscarded) String() string            { return proto.CompactTextString(m) }
func (*ChangeDiscarded) ProtoMessage()               {}
func (*ChangeDiscarded) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{26} }

// OK
type Ok struct {
	Type    uint32 `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Payload []byte `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *Ok) Reset()                    { *m = Ok{} }
func (m *Ok) String() string            { return proto.CompactTextString(m) }
func (*Ok) ProtoMessage()               {}
func (*Ok) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{27} }

// ERROR
type Error struct {
	Type  uint32 `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Error int32  `protobuf:"varint,2,opt,name=error" json:"error,omitempty"`
}

func (m *Error) Reset()                    { *m = Error{} }
func (m *Error) String() string            { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()               {}
func (*Error) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{28} }

// PING/PONG/CLOCK_RESPONSE
type TimeInfo struct {
	CurrentTime int64 `protobuf:"varint,1,opt,name=current_time,json=currentTime" json:"current_time,omitempty"`
}

func (m *TimeInfo) Reset()                    { *m = TimeInfo{} }
func (m *TimeInfo) String() string            { return proto.CompactTextString(m) }
func (*TimeInfo) ProtoMessage()               {}
func (*TimeInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{29} }

// READ EVENT
type ReadEvent struct {
	EventId int64 `protobuf:"varint,1,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
}

func (m *ReadEvent) Reset()                    { *m = ReadEvent{} }
func (m *ReadEvent) String() string            { return proto.CompactTextString(m) }
func (*ReadEvent) ProtoMessage()               {}
func (*ReadEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{30} }

// LIST AUTHORED EVENTS
// LIST PRIVATE EVENTS
// LIST PUBLIC EVENTS
// HISTORY AUTHORED EVENTS
// HISTORY PRIVATE EVENTS
// HISTORY PUBLIC EVENTS
type EventListRequest struct {
	StartWindow     int64          `protobuf:"varint,1,opt,name=start_window,json=startWindow" json:"start_window,omitempty"`
	EndWindow       int64          `protobuf:"varint,2,opt,name=end_window,json=endWindow" json:"end_window,omitempty"`
	UserCoordinates *core.Location `protobuf:"bytes,3,opt,name=user_coordinates,json=userCoordinates" json:"user_coordinates,omitempty"`
	RangeInMeters   uint32         `protobuf:"varint,4,opt,name=range_in_meters,json=rangeInMeters" json:"range_in_meters,omitempty"`
}

func (m *EventListRequest) Reset()                    { *m = EventListRequest{} }
func (m *EventListRequest) String() string            { return proto.CompactTextString(m) }
func (*EventListRequest) ProtoMessage()               {}
func (*EventListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{31} }

func (m *EventListRequest) GetUserCoordinates() *core.Location {
	if m != nil {
		return m.UserCoordinates
	}
	return nil
}

// EVENTS LIST
type EventsList struct {
	Event       []*core.Event `protobuf:"bytes,1,rep,name=event" json:"event,omitempty"`
	StartWindow int64         `protobuf:"varint,2,opt,name=startWindow" json:"startWindow,omitempty"`
	EndWindow   int64         `protobuf:"varint,3,opt,name=endWindow" json:"endWindow,omitempty"`
}

func (m *EventsList) Reset()                    { *m = EventsList{} }
func (m *EventsList) String() string            { return proto.CompactTextString(m) }
func (*EventsList) ProtoMessage()               {}
func (*EventsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{32} }

func (m *EventsList) GetEvent() []*core.Event {
	if m != nil {
		return m.Event
	}
	return nil
}

// FRIENDS LIST
type FriendsList struct {
	Friends []*core.Friend `protobuf:"bytes,1,rep,name=friends" json:"friends,omitempty"`
}

func (m *FriendsList) Reset()                    { *m = FriendsList{} }
func (m *FriendsList) String() string            { return proto.CompactTextString(m) }
func (*FriendsList) ProtoMessage()               {}
func (*FriendsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{33} }

func (m *FriendsList) GetFriends() []*core.Friend {
	if m != nil {
		return m.Friends
	}
	return nil
}

// GROUPS LIST
type GroupsList struct {
	Groups []*core.Group `protobuf:"bytes,1,rep,name=groups" json:"groups,omitempty"`
}

func (m *GroupsList) Reset()                    { *m = GroupsList{} }
func (m *GroupsList) String() string            { return proto.CompactTextString(m) }
func (*GroupsList) ProtoMessage()               {}
func (*GroupsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{34} }

func (m *GroupsList) GetGroups() []*core.Group {
	if m != nil {
		return m.Groups
	}
	return nil
}

// FRIEND REQUESTS LIST
type FriendRequestsList struct {
	FriendRequests []*core.FriendRequest `protobuf:"bytes,1,rep,name=friendRequests" json:"friendRequests,omitempty"`
}

func (m *FriendRequestsList) Reset()                    { *m = FriendRequestsList{} }
func (m *FriendRequestsList) String() string            { return proto.CompactTextString(m) }
func (*FriendRequestsList) ProtoMessage()               {}
func (*FriendRequestsList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{35} }

func (m *FriendRequestsList) GetFriendRequests() []*core.FriendRequest {
	if m != nil {
		return m.FriendRequests
	}
	return nil
}

func init() {
	proto.RegisterType((*AyiHeaderV2)(nil), "protocol.AyiHeaderV2")
	proto.RegisterType((*Hello)(nil), "protocol.Hello")
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
	proto.RegisterType((*LinkAccount)(nil), "protocol.LinkAccount")
	proto.RegisterType((*NewAuthToken)(nil), "protocol.NewAuthToken")
	proto.RegisterType((*AccessToken)(nil), "protocol.AccessToken")
	proto.RegisterType((*InstanceIDToken)(nil), "protocol.InstanceIDToken")
	proto.RegisterType((*SyncGroups)(nil), "protocol.SyncGroups")
	proto.RegisterType((*CreateFriendRequest)(nil), "protocol.CreateFriendRequest")
	proto.RegisterType((*ConfirmFriendRequest)(nil), "protocol.ConfirmFriendRequest")
	proto.RegisterType((*EventCancelled)(nil), "protocol.EventCancelled")
	proto.RegisterType((*EventExpired)(nil), "protocol.EventExpired")
	proto.RegisterType((*InvitationCancelled)(nil), "protocol.InvitationCancelled")
	proto.RegisterType((*AttendanceStatus)(nil), "protocol.AttendanceStatus")
	proto.RegisterType((*EventChangeProposed)(nil), "protocol.EventChangeProposed")
	proto.RegisterType((*VotingStatus)(nil), "protocol.VotingStatus")
	proto.RegisterType((*ChangeAccepted)(nil), "protocol.ChangeAccepted")
	proto.RegisterType((*ChangeDiscarded)(nil), "protocol.ChangeDiscarded")
	proto.RegisterType((*Ok)(nil), "protocol.Ok")
	proto.RegisterType((*Error)(nil), "protocol.Error")
	proto.RegisterType((*TimeInfo)(nil), "protocol.TimeInfo")
	proto.RegisterType((*ReadEvent)(nil), "protocol.ReadEvent")
	proto.RegisterType((*EventListRequest)(nil), "protocol.EventListRequest")
	proto.RegisterType((*EventsList)(nil), "protocol.EventsList")
	proto.RegisterType((*FriendsList)(nil), "protocol.FriendsList")
	proto.RegisterType((*GroupsList)(nil), "protocol.GroupsList")
	proto.RegisterType((*FriendRequestsList)(nil), "protocol.FriendRequestsList")
	proto.RegisterEnum("protocol.AuthType", AuthType_name, AuthType_value)
	proto.RegisterEnum("protocol.ConfirmFriendRequest_FriendRequestResponse", ConfirmFriendRequest_FriendRequestResponse_name, ConfirmFriendRequest_FriendRequestResponse_value)
}

func init() { proto.RegisterFile("protocol.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1531 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xb4, 0x57, 0x4b, 0x73, 0x1b, 0xc5,
	0x13, 0xff, 0x4b, 0xb6, 0x64, 0xa9, 0xf5, 0xb0, 0xbc, 0x4e, 0xf2, 0x57, 0xc2, 0x2b, 0x2c, 0x85,
	0x49, 0xa0, 0x50, 0x11, 0x43, 0x0e, 0x21, 0x45, 0x15, 0xb2, 0xe2, 0x24, 0x2a, 0x1c, 0xdb, 0x6c,
	0x8c, 0x39, 0x6e, 0xad, 0x77, 0x47, 0xf6, 0x94, 0x57, 0x3b, 0xcb, 0xce, 0xca, 0x8e, 0xa9, 0xe2,
	0xc4, 0x81, 0x33, 0x17, 0xce, 0xc0, 0x97, 0xe0, 0xce, 0x27, 0xa3, 0xa7, 0x67, 0xf6, 0x21, 0xe3,
	0x28, 0x55, 0x49, 0x71, 0x51, 0x6d, 0xff, 0xa6, 0xa7, 0xa7, 0x5f, 0xd3, 0xbf, 0x11, 0x74, 0xe3,
	0x44, 0xa4, 0xc2, 0x17, 0xe1, 0x80, 0x3e, 0xac, 0x46, 0x26, 0xdf, 0x02, 0x5f, 0x24, 0x4c, 0xa3,
	0xb6, 0x84, 0xd6, 0xf0, 0x82, 0x3f, 0x65, 0x5e, 0xc0, 0x92, 0xc3, 0x4d, 0xab, 0x0f, 0x2b, 0x67,
	0x2c, 0x91, 0x5c, 0x44, 0xfd, 0xca, 0xed, 0xca, 0x9d, 0x8e, 0x93, 0x89, 0xd6, 0x35, 0xa8, 0xa5,
	0xe2, 0x94, 0x45, 0xfd, 0x2a, 0xe1, 0x5a, 0xb0, 0x2c, 0x58, 0x4e, 0x2f, 0x62, 0xd6, 0x5f, 0x22,
	0x90, 0xbe, 0xad, 0xdb, 0xd0, 0x8a, 0xbd, 0x8b, 0x50, 0x78, 0xc1, 0x73, 0xfe, 0x23, 0xeb, 0x2f,
	0xd3, 0x52, 0x19, 0xb2, 0x7f, 0xaf, 0x40, 0xed, 0x29, 0x0b, 0x43, 0x61, 0xdd, 0x85, 0x5e, 0xe6,
	0x96, 0x3b, 0x7f, 0xf0, 0x6a, 0x86, 0x1f, 0x1a, 0x07, 0x3e, 0x84, 0xae, 0x1f, 0x72, 0x16, 0xa5,
	0xb9, 0xa2, 0xf2, 0xa4, 0xe9, 0x74, 0x34, 0x9a, 0xa9, 0xdd, 0x82, 0x46, 0x1c, 0x7a, 0xe9, 0x44,
	0x24, 0x53, 0xf2, 0xaa, 0xe9, 0xe4, 0x32, 0x9d, 0x66, 0xbe, 0x73, 0x23, 0xcb, 0xa4, 0xb3, 0x9a,
	0xe1, 0xc6, 0x8c, 0xfd, 0x77, 0x05, 0x5a, 0xa3, 0x84, 0x79, 0x29, 0xdb, 0x3e, 0x43, 0xeb, 0x2a,
	0x31, 0x53, 0x26, 0xa5, 0x77, 0xcc, 0xc8, 0xbf, 0xa6, 0x93, 0x89, 0xd6, 0xfb, 0xd0, 0xf6, 0x49,
	0x31, 0x70, 0x03, 0xfc, 0x25, 0xaf, 0x96, 0x9c, 0x96, 0xc1, 0x1e, 0xe1, 0x8f, 0xf5, 0x0e, 0x80,
	0x4c, 0xbd, 0x24, 0xd5, 0x0a, 0x4b, 0xa4, 0xd0, 0x24, 0x84, 0x96, 0x6f, 0x42, 0x83, 0x45, 0x66,
	0xf7, 0x32, 0x2d, 0xae, 0xa0, 0x4c, 0x4b, 0x36, 0xb4, 0x63, 0x54, 0xe3, 0x3e, 0x8f, 0xbd, 0x28,
	0x95, 0xfd, 0xda, 0xed, 0x25, 0x5c, 0x9e, 0xc3, 0x94, 0x6b, 0x31, 0xf7, 0xd3, 0x59, 0xc2, 0xfa,
	0x75, 0xdc, 0xdd, 0x76, 0x32, 0xd1, 0xfe, 0x1a, 0x63, 0xf0, 0x22, 0x9f, 0x85, 0x3a, 0x06, 0x75,
	0x8e, 0xfa, 0x70, 0x79, 0x40, 0x41, 0xa8, 0x73, 0x94, 0x3c, 0x0e, 0xac, 0x1b, 0x50, 0x47, 0x7f,
	0x65, 0x9e, 0x54, 0x23, 0xd9, 0x3b, 0xd0, 0x1a, 0x47, 0x67, 0x3c, 0x65, 0xdf, 0x49, 0x4c, 0xcc,
	0x22, 0x0b, 0x97, 0x3d, 0xad, 0xfe, 0xdb, 0x53, 0xfb, 0x10, 0xae, 0x6b, 0x7f, 0xc8, 0x1a, 0x19,
	0xf6, 0x52, 0x55, 0xb4, 0x37, 0xb4, 0xcb, 0x61, 0x6d, 0x24, 0xa2, 0x09, 0x4f, 0xa6, 0xc3, 0x34,
	0xc5, 0xd4, 0xa9, 0x33, 0x16, 0xd9, 0x7c, 0x00, 0x2d, 0xcf, 0x57, 0x07, 0xbb, 0xbe, 0x08, 0x74,
	0xc5, 0xba, 0x9b, 0xfd, 0x01, 0x5d, 0x8b, 0xc2, 0x82, 0xc3, 0x64, 0x2c, 0x22, 0xc9, 0x1c, 0xd0,
	0xca, 0x23, 0xd4, 0xb5, 0x7f, 0xc3, 0xbe, 0x78, 0x26, 0x02, 0x3e, 0xb9, 0x78, 0x65, 0x4e, 0x4b,
	0x2d, 0x53, 0x9d, 0x6f, 0x99, 0xd7, 0xef, 0x87, 0x52, 0xad, 0x6b, 0xf3, 0xb5, 0xe6, 0x00, 0x87,
	0x22, 0x65, 0xa3, 0x13, 0x2f, 0x3a, 0x5e, 0x18, 0xfc, 0x5b, 0xd0, 0xf4, 0x49, 0x49, 0xad, 0x29,
	0xc7, 0x6a, 0x4e, 0x43, 0x03, 0xb8, 0xf8, 0x01, 0x74, 0x3c, 0xdf, 0x67, 0x71, 0xea, 0x6a, 0x88,
	0x9c, 0x6b, 0x38, 0x6d, 0x0d, 0x6a, 0xe3, 0xf6, 0x0b, 0x68, 0xab, 0x02, 0xee, 0x0b, 0xc9, 0xa9,
	0x7a, 0x5f, 0x81, 0x75, 0x1c, 0x8a, 0x23, 0x2f, 0xc4, 0x74, 0x8a, 0x24, 0xe0, 0x11, 0x7a, 0x2a,
	0xe9, 0xd8, 0xd6, 0x66, 0x57, 0x67, 0x75, 0x47, 0xf8, 0x54, 0x69, 0x67, 0x4d, 0x6b, 0x8e, 0x0a,
	0x45, 0x75, 0x2b, 0x99, 0x4c, 0xf9, 0x94, 0x14, 0x5c, 0x96, 0x24, 0x22, 0x21, 0xbf, 0xaa, 0xce,
	0x6a, 0x81, 0x6f, 0x2b, 0xd8, 0x7e, 0x08, 0x6b, 0xe5, 0x93, 0x1d, 0x8a, 0x75, 0x03, 0x56, 0x13,
	0x1d, 0x4f, 0xe4, 0x4e, 0x59, 0x8a, 0x9d, 0x45, 0x67, 0x57, 0x9d, 0x0e, 0xc1, 0xe3, 0xe8, 0x19,
	0x81, 0xf6, 0x5f, 0x15, 0x6c, 0x13, 0xba, 0x95, 0xca, 0xc6, 0xd0, 0xf7, 0xc5, 0x0c, 0x0b, 0x88,
	0x13, 0x2c, 0xf2, 0xa6, 0xd9, 0xad, 0xa6, 0x6f, 0x35, 0xeb, 0xd8, 0xd4, 0xe3, 0xa1, 0xa9, 0x9b,
	0x16, 0x68, 0xb2, 0x78, 0x52, 0x9e, 0xa3, 0xe7, 0xf9, 0x64, 0x31, 0xb2, 0xda, 0x11, 0x9f, 0x88,
	0x88, 0x99, 0x71, 0xa2, 0x05, 0x65, 0x7b, 0x72, 0x84, 0x59, 0xae, 0x69, 0xdb, 0xea, 0x5b, 0x55,
	0x70, 0x72, 0xa4, 0x27, 0x69, 0x5d, 0x77, 0x85, 0x11, 0xcb, 0xb5, 0x5d, 0x99, 0xaf, 0xed, 0x1f,
	0xd8, 0x74, 0x3b, 0x3c, 0x3a, 0xcd, 0x7c, 0xfe, 0x3f, 0xac, 0xcc, 0x30, 0x84, 0xa2, 0xb8, 0x75,
	0x25, 0x62, 0xf9, 0xee, 0x83, 0x9a, 0xf2, 0x67, 0x1c, 0x87, 0xb9, 0xe9, 0xea, 0x9b, 0xa6, 0xab,
	0xf5, 0xce, 0x7d, 0xb3, 0x78, 0x80, 0x73, 0xda, 0xc9, 0x55, 0x55, 0x3f, 0x7a, 0x5a, 0x41, 0x99,
	0xd4, 0xb1, 0x35, 0x0d, 0x92, 0x37, 0x05, 0x2d, 0x6b, 0xc7, 0x75, 0x90, 0x6d, 0x03, 0x1e, 0x28,
	0xcc, 0x3e, 0x82, 0xf6, 0x2e, 0x3b, 0x1f, 0xce, 0xd2, 0x13, 0x92, 0x29, 0x23, 0x98, 0x9d, 0x7b,
	0x26, 0xb1, 0x5a, 0xc8, 0xd0, 0xcd, 0x2c, 0xb3, 0x24, 0x60, 0x05, 0x0b, 0x16, 0xe9, 0x6e, 0x5a,
	0x83, 0x9c, 0xb9, 0xc8, 0x9c, 0xf2, 0x95, 0xd6, 0xed, 0x6d, 0x24, 0x2b, 0x6c, 0x44, 0x29, 0xf5,
	0x11, 0x2f, 0x4d, 0x83, 0x8a, 0x07, 0x77, 0xba, 0x05, 0x61, 0xa9, 0x78, 0x32, 0xd7, 0xec, 0x8f,
	0x60, 0x75, 0x1c, 0xe1, 0x75, 0xc3, 0x3b, 0x3e, 0x7e, 0x94, 0x7b, 0xab, 0x95, 0x8d, 0xb7, 0x24,
	0xd8, 0x3f, 0x57, 0x00, 0x9e, 0x5f, 0x44, 0xfe, 0x93, 0x44, 0xcc, 0x62, 0xa9, 0x94, 0xc4, 0x79,
	0x84, 0xa9, 0xd5, 0xa7, 0x69, 0x01, 0xb3, 0x53, 0x3f, 0xa6, 0x75, 0x1a, 0x4d, 0xad, 0xcd, 0x96,
	0xce, 0x38, 0xed, 0x71, 0xcc, 0x92, 0xf5, 0x25, 0x74, 0x25, 0x1a, 0x72, 0x8f, 0xd8, 0x89, 0x77,
	0xc6, 0xc5, 0x2c, 0x31, 0xb1, 0xae, 0x6b, 0x65, 0x75, 0xc8, 0x56, 0xb6, 0xe4, 0x74, 0x64, 0x59,
	0xb4, 0x3f, 0x81, 0x75, 0xdd, 0xb6, 0x8f, 0x13, 0x24, 0xba, 0xc0, 0x61, 0x3f, 0xcc, 0xf0, 0x5e,
	0x14, 0x4d, 0x5a, 0x29, 0x35, 0xa9, 0x6a, 0xf2, 0x6b, 0x66, 0x16, 0xce, 0xab, 0xe3, 0xb5, 0x9f,
	0x10, 0x50, 0xa4, 0xab, 0xa1, 0x01, 0x4c, 0xd8, 0x3e, 0x34, 0x12, 0x33, 0xed, 0x4c, 0xdf, 0x7c,
	0x51, 0x14, 0xe1, 0x2a, 0x73, 0x83, 0x39, 0x29, 0x9f, 0x94, 0xb9, 0x15, 0xfb, 0x33, 0xb8, 0x7e,
	0xa5, 0x8a, 0x05, 0x50, 0x1f, 0x0d, 0x77, 0x47, 0xdb, 0x3b, 0xbd, 0xff, 0x59, 0x2d, 0x58, 0x19,
	0xed, 0xed, 0x3e, 0x1e, 0x3b, 0xcf, 0x7a, 0x15, 0xfb, 0x27, 0xe8, 0xd2, 0x48, 0xd5, 0x0c, 0x11,
	0xb2, 0xc0, 0xba, 0x0e, 0xf5, 0xf3, 0x13, 0x51, 0xf8, 0x5b, 0x43, 0x09, 0x9d, 0x2d, 0xcf, 0xb6,
	0xea, 0xcb, 0x68, 0x6c, 0xa9, 0x4c, 0x63, 0xc8, 0xd1, 0x35, 0x52, 0xa1, 0xce, 0xcd, 0x4b, 0x44,
	0xc7, 0x39, 0x7a, 0xc5, 0xbe, 0x0b, 0x6d, 0x92, 0xb7, 0x5f, 0xc4, 0x3c, 0x61, 0xc1, 0x82, 0x09,
	0x8a, 0xb1, 0xad, 0x17, 0xdc, 0x55, 0xb8, 0xbb, 0x60, 0xc7, 0xaf, 0x15, 0xe8, 0x15, 0xc4, 0xf2,
	0x1c, 0x77, 0xce, 0x16, 0x92, 0xe9, 0x08, 0xd6, 0xbc, 0x5c, 0xdd, 0x95, 0xa4, 0x6f, 0xda, 0xeb,
	0x46, 0xc9, 0xf7, 0xfd, 0x82, 0x04, 0x9d, 0x9e, 0x77, 0xd9, 0x3e, 0xde, 0x82, 0x68, 0x36, 0x75,
	0x8f, 0x55, 0xfa, 0x25, 0x25, 0xa4, 0xe6, 0x34, 0x11, 0x79, 0x42, 0x80, 0x1a, 0x2a, 0xeb, 0x3a,
	0xe1, 0x34, 0xd5, 0x71, 0x34, 0xc4, 0x42, 0x2e, 0x0c, 0x63, 0x31, 0x75, 0xbc, 0x11, 0xa9, 0x65,
	0x44, 0x59, 0x9b, 0x23, 0x4a, 0xfb, 0x97, 0x2a, 0xb4, 0x91, 0xd5, 0x78, 0x74, 0xfc, 0xea, 0x9c,
	0xfd, 0x47, 0xce, 0xe1, 0xf3, 0x8e, 0x85, 0x5e, 0x8c, 0x99, 0x71, 0x91, 0x8b, 0xb4, 0x87, 0xf8,
	0xbc, 0x33, 0xd8, 0x01, 0x42, 0xea, 0x65, 0x7a, 0x86, 0xd4, 0x2b, 0xdd, 0x84, 0xf9, 0x8c, 0x9f,
	0xb1, 0x80, 0x26, 0x7b, 0xc7, 0xe9, 0x10, 0xea, 0x18, 0xd0, 0x7a, 0x0f, 0x5a, 0x5a, 0x2d, 0x15,
	0xa9, 0x17, 0xd2, 0x8c, 0xef, 0x38, 0x40, 0xd0, 0x81, 0x42, 0x14, 0xc1, 0x4c, 0x78, 0xc4, 0xe5,
	0x09, 0x5a, 0x68, 0x10, 0xef, 0xe6, 0xb2, 0xfd, 0x14, 0xba, 0xba, 0x4e, 0x43, 0x62, 0xe2, 0xd7,
	0xaf, 0x93, 0x3d, 0x86, 0x55, 0x6d, 0xe9, 0x11, 0x97, 0xbe, 0x97, 0x04, 0x6f, 0x60, 0x6a, 0x13,
	0xaa, 0x7b, 0xa7, 0xf9, 0x7f, 0x80, 0x4a, 0xe9, 0x3f, 0x80, 0xe2, 0x32, 0xfd, 0xe0, 0xa7, 0x4d,
	0x8a, 0xcb, 0xb4, 0x68, 0xdf, 0x83, 0x1a, 0x71, 0xf9, 0x95, 0xdb, 0xd4, 0x4c, 0xcb, 0xf9, 0xbf,
	0xe6, 0x68, 0xc1, 0xfe, 0x14, 0x1a, 0x2a, 0xcf, 0xe3, 0x68, 0x22, 0xe8, 0xb5, 0x3d, 0x4b, 0x12,
	0xe5, 0x2c, 0x95, 0xa3, 0x62, 0x5e, 0xdb, 0x1a, 0x53, 0x6a, 0xf6, 0x06, 0x34, 0x1d, 0xfc, 0x3f,
	0xf3, 0xaa, 0xf7, 0x99, 0x1a, 0x95, 0x3d, 0x52, 0xda, 0xe1, 0x6a, 0x3e, 0xe9, 0x31, 0x89, 0xf6,
	0x75, 0xa3, 0x9c, 0xf3, 0x28, 0x10, 0xe7, 0x99, 0x7d, 0xc2, 0xbe, 0x27, 0x48, 0xf5, 0x92, 0x6a,
	0x16, 0xa3, 0xa0, 0x27, 0x50, 0x13, 0x11, 0xb3, 0xfc, 0x00, 0x7a, 0xc4, 0x4a, 0xe5, 0xb7, 0xd0,
	0xd2, 0x95, 0x6f, 0xa1, 0x55, 0xa5, 0x57, 0x7e, 0x09, 0x5d, 0xf1, 0x92, 0xd1, 0xff, 0x9e, 0x2e,
	0xbd, 0x64, 0x04, 0x00, 0x39, 0x2e, 0x95, 0xe7, 0xc5, 0x70, 0xab, 0x94, 0xf9, 0xa7, 0x3c, 0xdc,
	0xd4, 0x5f, 0xb2, 0x52, 0x04, 0xd9, 0x5f, 0x94, 0x72, 0x50, 0x6f, 0x43, 0x11, 0x42, 0x76, 0x3f,
	0x72, 0xc0, 0xbe, 0x0f, 0x2d, 0x3d, 0xcd, 0xf5, 0x89, 0x1b, 0xf8, 0x86, 0xd1, 0xa2, 0x39, 0xb3,
	0xad, 0xcf, 0x34, 0x13, 0x3f, 0x5b, 0xc4, 0x5a, 0x83, 0xa6, 0x4e, 0xda, 0x55, 0x10, 0x65, 0xe5,
	0xa5, 0x44, 0x69, 0x7f, 0x0b, 0xd6, 0x1c, 0x6f, 0xe8, 0xad, 0x0f, 0xa1, 0x3b, 0x99, 0x43, 0x8d,
	0x89, 0xf5, 0xb9, 0x73, 0x0d, 0xd3, 0x5c, 0x52, 0xfd, 0xf8, 0x0e, 0x34, 0xb2, 0x77, 0x84, 0xd5,
	0xc6, 0x6f, 0x77, 0x77, 0x78, 0x30, 0x3e, 0xdc, 0x46, 0xfe, 0xe9, 0x02, 0x0c, 0xdd, 0xc7, 0xc3,
	0xd1, 0xf6, 0xd6, 0xde, 0xde, 0x37, 0xbd, 0xca, 0xd6, 0x06, 0xbc, 0xcb, 0xe4, 0x20, 0x66, 0x2c,
	0x0e, 0xd9, 0xc0, 0x4b, 0xd8, 0x85, 0x98, 0xf1, 0x68, 0x20, 0x83, 0xd3, 0x41, 0xc4, 0x52, 0x7c,
	0xe7, 0x9d, 0xfe, 0x59, 0x5d, 0x1a, 0xee, 0x6f, 0x1d, 0xd5, 0x89, 0x1b, 0x3f, 0xff, 0x27, 0x00,
	0x00, 0xff, 0xff, 0x18, 0x10, 0xcd, 0x22, 0x63, 0x0f, 0x00, 0x00,
}
