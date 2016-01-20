package dao

import (
	"errors"
	"github.com/gocql/gocql"
)

var (
	ErrNotFound    = gocql.ErrNotFound
	ErrEmptyInbox  = errors.New("user inbox is empty")
	ErrNoSession   = errors.New("No session to Cassandra available")
	ErrInvalidArg  = errors.New("Invalid arguments")
	ErrInvalidUser = errors.New("Invalid user account")
	ErrUnexpected  = errors.New("Unexpected error")
)
