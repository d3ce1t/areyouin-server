package protocol

type PacketType uint8

// Modifiers
const (
	M_CREATE_EVENT PacketType = 0x00 + iota
	M_CANCEL_EVENT
	M_INVITE_USERS
	M_CANCEL_USERS_INVITATION
	M_CONFIRM_ATTENDANCE
	M_MODIFY_EVENT_DATE
	M_MODIFY_EVENT_MESSAGE
	M_MODIFY_EVENT
	M_VOTE_CHANGE
	M_USER_POSITION
	M_USER_POSITION_RANGE
	M_USER_CREATE_ACCOUNT
	M_USER_NEW_AUTH_TOKEN
	M_USER_AUTH
)

// Notifications
const (
	M_EVENT_CREATED PacketType = 0x40 + iota
	M_EVENT_CANCELLED
	M_EVENT_EXPIRED
	M_EVENT_DATE_MODIFIED
	M_EVENT_MESSAGE_MODIFIED
	M_EVENT_MODIFIED
	M_INVITATION_RECEIVED
	M_INVITATION_CANCELLED
	M_ATTENDANCE_STATUS
	M_EVENT_CHANGE_DATE_PROPOSED
	M_EVENT_CHANGE_MESSAGE_PROPOSED
	M_EVENT_CHANGE_PROPOSED
	M_VOTING_STATUS
	M_VOTING_FINISHED
	M_CHANGE_ACCEPTED
	M_CHANGE_DISCARDED
	M_ACCESS_GRANTED
	M_OK    = 0x7E
	M_ERROR = 0x7F
)

// Requests
const (
	M_PING PacketType = 0x80 + iota
	M_READ_EVENT
	M_LIST_AUTHORED_EVENTS
	M_LIST_PRIVATE_EVENTS
	M_LIST_PUBLIC_EVENTS
	M_HISTORY_AUTHORED_EVENTS
	M_HISTORY_PRIVATE_EVENTS
	M_HISTORY_PUBLIC_EVENTS
	M_USER_FRIENDS
)

// Responses
const (
	M_PONG PacketType = 0xC0 + iota
	M_EVENT_INFO
	M_EVENTS_LIST
	M_FRIENDS_LIST
)
