package cqldao

import (
	"errors"
	"peeple/areyouin/api"

	"github.com/gocql/gocql"
)

var (
	ErrNoSession        = errors.New("no session to Cassandra available")
	ErrGracePeriod      = errors.New("grace period due to old and new conflict")
	ErrUnexpected       = errors.New("unexpected error")
	ErrInconsistency    = errors.New("db inconsistency detected")
	ErrIllegalArguments = errors.New("illegal arguments")
)

func convErr(err error) error {
	switch err {
	case gocql.ErrNotFound:
		return api.ErrNotFound
	}
	return err
}
