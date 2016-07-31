package cqldao

import (
	"peeple/areyouin/api"
)

func NewEventDAO(session api.DbSession) api.EventDAO {
	reconnectIfNeeded(session)
	return &EventDAO{session: session.(*GocqlSession)}
}

func NewUserDAO(session api.DbSession) api.UserDAO {
	reconnectIfNeeded(session)
	return &UserDAO{session: session.(*GocqlSession)}
}

func NewFriendDAO(session api.DbSession) api.FriendDAO {
	reconnectIfNeeded(session)
	return &FriendDAO{session: session.(*GocqlSession)}
}

func NewFriendRequestDAO(session api.DbSession) api.FriendRequestDAO {
	reconnectIfNeeded(session)
	return &FriendRequestDAO{session: session.(*GocqlSession)}
}

func NewThumbnailDAO(session api.DbSession) api.ThumbnailDAO {
	reconnectIfNeeded(session)
	return &ThumbnailDAO{session: session.(*GocqlSession)}
}

func NewAccessTokenDAO(session api.DbSession) api.AccessTokenDAO {
	reconnectIfNeeded(session)
	return &AccessTokenDAO{session: session.(*GocqlSession)}
}

func NewLogDAO(session api.DbSession) api.LogDAO {
	reconnectIfNeeded(session)
	return &LogDAO{session: session.(*GocqlSession)}
}

func checkSession(session *GocqlSession) {
	if session == nil || !session.IsValid() {
		panic(ErrNoSession)
	}
}

/*func checkUserID(user_id int64) {
	if user_id == 0 {
		panic(api.ErrInvalidArg)
	}
}*/

func reconnectIfNeeded(session api.DbSession) {
	if session != nil && (!session.IsValid() || session.Closed()) {
		session.Connect()
	}
}
