package protocol

import (
	"errors"
)

const (
	E_NO_ERROR                 int32 = iota
	E_INVALID_USER_OR_PASSWORD       // Auth, AuthNewToken
	E_EMAIL_EXISTS                   // Create User Account
	E_FB_EXISTS                      // Create User Account
	E_FB_INVALID_ACCESS              // Create User Account
	E_INVALID_INPUT                  // Create User Account
	E_MALFORMED_MESSAGE              // User Friend
	E_EVENT_PARTICIPANTS_REQUIRED
	E_OPERATION_FAILED
	E_INVALID_EVENT_OR_PARTICIPANT
	E_INVALID_PARTICIPANT
	E_EVENT_OUT_OF_CREATE_WINDOW
	E_EVENT_INVALID_START_DATE
	E_EVENT_INVALID_END_DATE
)

var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrTimeout          = errors.New("input/output timeout")
	ErrInvalidSocket    = errors.New("invalid connection socket")
)
