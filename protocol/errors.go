package protocol

const (
	E_NO_ERROR          int32 = iota
	E_INVALID_USER            // Auth, AuthNewToken
	E_USER_EXISTS             // Create User Account
	E_FB_MISSING_DATA         // Create User Account
	E_FB_INVALID_TOKEN        // Create User Account
	E_MALFORMED_MESSAGE       // User Friend
	E_EVENT_CREATION_ERROR
)

const (
	OK_AUTH int32 = iota
	OK_ACK
)
