package api

import (
	"errors"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrNoResults             = errors.New("user inbox is empty")
	ErrInvalidArg            = errors.New("invalid arguments")
	ErrInvalidEmail          = errors.New("invalid e-mail")
	ErrEmailAlreadyExists    = errors.New("e-mail already exists")
	ErrFacebookAlreadyExists = errors.New("facebook already exists")
	ErrUnexpected            = errors.New("unexpected error")
	ErrAccountMismatch       = errors.New("account mismatch")
)
