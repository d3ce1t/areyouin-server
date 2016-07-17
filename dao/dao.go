package dao

import (
  core "peeple/areyouin/common"
)

func NewEventDAO(session core.DbSession) core.EventDAO {
  reconnectIfNeeded(session)
	return &EventDAO{session: session.(*GocqlSession)}
}

func NewUserDAO(session core.DbSession) core.UserDAO {
  reconnectIfNeeded(session)
	return &UserDAO{session: session.(*GocqlSession)}
}

func NewFriendDAO(session core.DbSession) core.FriendDAO {
  reconnectIfNeeded(session)
	return &FriendDAO{session: session.(*GocqlSession)}
}

func NewThumbnailDAO(session core.DbSession) core.ThumbnailDAO {
  reconnectIfNeeded(session)
	return &ThumbnailDAO{session: session.(*GocqlSession)}
}

func NewAccessTokenDAO(session core.DbSession) core.AccessTokenDAO {
  reconnectIfNeeded(session)
	return &AccessTokenDAO{session: session.(*GocqlSession)}
}

func checkSession(session *GocqlSession) {
  if session == nil || !session.IsValid() {
    panic(ErrNoSession)
  }
}

func checkUserID(user_id int64) {
  if user_id == 0 {
    panic(ErrInvalidArg)
  }
}

func reconnectIfNeeded(session core.DbSession) {
  if session != nil && (!session.IsValid() || session.Closed()) {
    session.Connect()
  }
}
