package cqldao

import (
	"peeple/areyouin/api"
)

type userEvent struct {
	UserId     int64
	EventId    int64
	AuthorId   int64
	AuthorName string
	StartDate  int64
	Message    string
	Response   api.AttendanceResponse
}

type emailCredential struct {
	Email    string
	Password [32]byte
	Salt     [32]byte
	UserId   int64
}

type fbCredential struct {
	FbId        string
	FbToken     string
	UserId      int64
	CreatedDate int64
}
