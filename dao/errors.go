package dao

import (
	"errors"
	"github.com/gocql/gocql"
)

var (
	ErrNotFound  = gocql.ErrNotFound
	ErrNoSession = errors.New("No session to Cassandra available")
)
