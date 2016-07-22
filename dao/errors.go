package dao

import (
	"errors"

	"github.com/gocql/gocql"
)

var (
	ErrNotFound                   = gocql.ErrNotFound
	ErrNoResults                  = errors.New("user inbox is empty")
	ErrNoSession                  = errors.New("no session to Cassandra available")
	ErrInvalidArg                 = errors.New("invalid arguments")
	ErrInvalidEmail               = errors.New("invalid e-mail")
	ErrEmailAlreadyExists         = errors.New("e-mail already exists")
	ErrFacebookAlreadyExists      = errors.New("facebook already exists")
	ErrGracePeriod                = errors.New("grace period due to old and new conflict")
	ErrUnexpected                 = errors.New("unexpected error")
	ErrNilPointer                 = errors.New("nil pointer")
	ErrAccountMismatch            = errors.New("account mismatch")
	ErrNotSupportedOperation      = errors.New("operation not supported")
)
