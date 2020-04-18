package cqldao

import (
	"github.com/d3ce1t/areyouin-server/api"
)

type SettingsDAO struct {
	session *GocqlSession
}

func NewSettingsDAO(session api.DbSession) api.SettingsDAO {
	reconnectIfNeeded(session)
	return &SettingsDAO{session: session.(*GocqlSession)}
}

func (d *SettingsDAO) Find(key api.SettingOption) (string, error) {
	checkSession(d.session)
	stmt := `SELECT value FROM settings WHERE key = ?`
	var value string
	if err := d.session.Query(stmt, key).Scan(&value); err != nil {
		return "", convErr(err)
	}
	return value, nil
}

func (d *SettingsDAO) Insert(key api.SettingOption, value string) error {
	checkSession(d.session)
	stmt := `INSERT INTO settings (key, value) VALUES (?, ?)`
	q := d.session.Query(stmt, key, value)
	return convErr(q.Exec())
}
