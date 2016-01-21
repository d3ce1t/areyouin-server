package dao

import (
	"errors"
	"github.com/gocql/gocql"
)

var (
	ErrNotFound     = gocql.ErrNotFound
	ErrEmptyInbox   = errors.New("User inbox is empty")
	ErrNoSession    = errors.New("No session to Cassandra available")
	ErrInvalidArg   = errors.New("Invalid arguments")
	ErrInvalidUser  = errors.New("Invalid user account")
	ErrInvalidEmail = errors.New("Invalid e-mail")
	ErrUnexpected   = errors.New("Unexpected error")
	ErrNilPointer   = errors.New("Nil pointer")
)
