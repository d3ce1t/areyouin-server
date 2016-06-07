package dao

import (
  core "peeple/areyouin/common"
  "github.com/gocql/gocql"
)

type DataAccessObject interface {
  GetSession() *gocql.Session
}

func NewEventDAO(session *gocql.Session) core.EventDAO {
	return &EventDAO{session: session}
}

func NewUserDAO(session *gocql.Session) core.UserDAO {
	return &UserDAO{session: session}
}

func NewFriendDAO(session *gocql.Session) core.FriendDAO {
	return &FriendDAO{session: session}
}

func NewThumbnailDAO(session *gocql.Session) core.ThumbnailDAO {
	return &ThumbnailDAO{session: session}
}

func checkSession(dao DataAccessObject) {
	if dao.GetSession() == nil {
		panic(ErrNoSession)
	}
}

func checkUserID(user_id int64) {
  if user_id == 0 {
    panic(ErrInvalidArg)
  }
}
