package main

import (
	"errors"
)

var (
	ErrAuthRequired     = errors.New("auth required")
	ErrUnhandledMessage = errors.New("unhandled message")
	ErrUnknownMessage   = errors.New("unknown message")
)
