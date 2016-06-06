package dao

import (
	"github.com/gocql/gocql"
	core "peeple/areyouin/common"
)

func NewAccessTokenDAO(session *gocql.Session) core.AccessTokenDAO {
	return &AccessTokenDAO{session: session}
}

type AccessTokenDAO struct {
	session *gocql.Session
}

func (dao *AccessTokenDAO) GetSession() *gocql.Session {
	return dao.session
}

func (dao *AccessTokenDAO) Insert(user_id int64, token string) error {

	checkSession(dao)

	if user_id == 0 || token == "" {
		return ErrInvalidArg
	}

	stmt := `INSERT INTO user_access_token (user_id, access_token, created_date)
			VALUES (?, ?, ?)`

	return dao.session.Query(stmt, user_id, token, core.GetCurrentTimeMillis()).Exec()
}

func (dao *AccessTokenDAO) CheckAccessToken(user_id int64, access_token string) (bool, error) {

	checkSession(dao)

	stmt := `SELECT access_token FROM user_access_token WHERE user_id = ? LIMIT 1`
	q := dao.session.Query(stmt, user_id)

	var stored_token gocql.UUID

	err := q.Scan(&stored_token)

	if err == gocql.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if access_token != stored_token.String() {
		return false, nil
	}

	return true, nil
}

func (dao *AccessTokenDAO) SetLastUsed(user_id int64, time int64) error {

	checkSession(dao)

	if user_id == 0 || time < 0 {
		return ErrInvalidArg
	}

	stmt := `UPDATE user_access_token SET last_used = ?	WHERE user_id = ?`
	return dao.session.Query(stmt, time, user_id).Exec()
}

func (dao *AccessTokenDAO) Remove(user_id int64) error {

	checkSession(dao)

	if user_id == 0 {
		return ErrInvalidArg
	}

	return dao.session.Query(`DELETE FROM user_access_token WHERE user_id = ?`,
		user_id).Exec()
}
