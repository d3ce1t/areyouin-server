package dao

import (
	"errors"
	"github.com/gocql/gocql"
)

var (
	ErrNotFound                   = gocql.ErrNotFound
	ErrNotFoundEventOrParticipant = errors.New("event or participant not found")
	ErrEmptyInbox                 = errors.New("User inbox is empty")
	ErrNoSession                  = errors.New("No session to Cassandra available")
	ErrInvalidArg                 = errors.New("Invalid arguments")
	ErrInvalidUser                = errors.New("Invalid user account")
	ErrInvalidEmail               = errors.New("Invalid e-mail")
	ErrEmailAlreadyExists         = errors.New("Email already exists")
	ErrFacebookAlreadyExists      = errors.New("Facebook already exists")
	ErrGracePeriod                = errors.New("Grace period due to old and new conflict")
	ErrUnexpected                 = errors.New("Unexpected error")
	ErrNilPointer                 = errors.New("Nil pointer")
	ErrAccountMismatch            = errors.New("account mismatch")
)
