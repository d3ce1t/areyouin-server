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
	E_INPUT_INVALID_EMAIL_ADDRESS
	E_INPUT_INVALID_USER_NAME
	E_EVENT_CANNOT_BE_MODIFIED
	E_AUTH_REQUIRED
	E_INVALID_EVENT
	E_EVENT_AUTHOR_MISMATCH
	E_ACCESS_DENIED
)

var (
	ErrConnectionClosed   = errors.New("connection closed")
	ErrTimeout            = errors.New("input/output timeout")
	ErrInvalidSocket      = errors.New("invalid connection socket")
	ErrMaxPayloadExceeded = errors.New("max payload exceeded")
	ErrIncompleteWrite    = errors.New("incomplete write")
	ErrUnknownMessage     = errors.New("unknown message")
	ErrNoPayload          = errors.New("packet conveys no message")
	//ErrMalformedHeader    = errors.New("Malformed header")
)
