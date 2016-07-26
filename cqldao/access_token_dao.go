package cqldao

import (
	"github.com/gocql/gocql"
	"peeple/areyouin/api"
)

type AccessTokenDAO struct {
	session *GocqlSession
}

func (d *AccessTokenDAO) Insert(accessToken *api.AccessTokenDTO) error {

	checkSession(d.session)

	if accessToken.UserId == 0 || accessToken.Token == "" {
		return api.ErrInvalidArg
	}

	stmt := `INSERT INTO user_access_token (user_id, access_token, created_date,
		last_used) VALUES (?, ?, ?, ?)`

	q := d.session.Query(stmt, accessToken.UserId, accessToken.Token,
		accessToken.CreatedDate, accessToken.LastUsed)

	return convErr(q.Exec())
}

func (d *AccessTokenDAO) Load(userId int64) (*api.AccessTokenDTO, error) {

	checkSession(d.session)

	if userId == 0 {
		return nil, api.ErrInvalidArg
	}

	stmt := `SELECT access_token, created_date, last_used FROM user_access_token
		WHERE user_id = ? LIMIT 1`

	q := d.session.Query(stmt, userId)

	dto := &api.AccessTokenDTO{UserId: userId}
	var storedToken gocql.UUID

	err := q.Scan(&storedToken, &dto.CreatedDate, &dto.LastUsed)
	if err != nil {
		return nil, convErr(err)
	}

	dto.Token = storedToken.String()

	return dto, nil
}

func (d *AccessTokenDAO) Remove(userId int64) error {

	checkSession(d.session)

	if userId == 0 {
		return api.ErrInvalidArg
	}

	stmt := `DELETE FROM user_access_token WHERE user_id = ?`
	err := d.session.Query(stmt, userId).Exec()
	return convErr(err)
}

func (d *AccessTokenDAO) SetLastUsed(user_id int64, time int64) error {

	checkSession(d.session)

	if user_id == 0 || time < 0 {
		return api.ErrInvalidArg
	}

	stmt := `UPDATE user_access_token SET last_used = ?	WHERE user_id = ?`
	err := d.session.Query(stmt, time, user_id).Exec()

	return convErr(err)
}
