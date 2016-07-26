package api

type AttendanceResponse int8

const (
	AttendanceResponse_NO_RESPONSE AttendanceResponse = 0
	AttendanceResponse_NO_ASSIST   AttendanceResponse = 1
	AttendanceResponse_MAYBE       AttendanceResponse = 2
	AttendanceResponse_ASSIST      AttendanceResponse = 3
)

type InvitationStatus int8

const (
	InvitationStatus_NO_DELIVERED     InvitationStatus = 0
	InvitationStatus_SERVER_DELIVERED InvitationStatus = 1
	InvitationStatus_CLIENT_DELIVERED InvitationStatus = 2
)

type EventState int8

const (
	EventState_NOT_STARTED EventState = 0
	EventState_ONGOING     EventState = 1
	EventState_FINISHED    EventState = 2
	EventState_CANCELLED   EventState = 3
)
