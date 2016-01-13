package main

import (
	"errors"
)

var (
	ErrAuthRequired               = errors.New("auth required")
	ErrUnhandledMessage           = errors.New("unhandled message")
	ErrUnknownMessage             = errors.New("unknown message")
	ErrNonFriendsIgnored          = errors.New("ignored non friends participants")
	ErrUnregisteredFriendsIgnored = errors.New("ignored unregistered participants")
)
