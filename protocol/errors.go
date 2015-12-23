package protocol

import (
	"errors"
)

const (
	E_NO_ERROR          int32 = iota
	E_INVALID_USER            // Auth, AuthNewToken
	E_USER_EXISTS             // Create User Account
	E_FB_MISSING_DATA         // Create User Account
	E_FB_INVALID_TOKEN        // Create User Account
	E_MALFORMED_MESSAGE       // User Friend
	E_EVENT_PARTICIPANTS_REQUIRED
	E_OPERATION_FAILED
	E_INVALID_EVENT_OR_PARTICIPANT
)

var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrTimeout          = errors.New("input/output timeout")
)
