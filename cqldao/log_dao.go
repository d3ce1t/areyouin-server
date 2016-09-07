package cqldao

import (
	"fmt"
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"strconv"
	"time"
)

type LogDAO struct {
	session *GocqlSession
}

func (d *LogDAO) LogRegisteredUser(user_id int64, created_date int64) error {
	checkSession(d.session)
	time := utils.UnixMillisToTimeUTC(created_date)
	day := d.formatDate(time.Year(), int(time.Month()), time.Day())
	stmt := `INSERT INTO log_registered_users_by_day (day, user_id, created_date)
        VALUES (?, ?, ?)`
	return convErr(d.session.Query(stmt, day, user_id, created_date).Exec())
}

func (d *LogDAO) LogActiveSession(node int, user_id int64, last_time int64) error {
	checkSession(d.session)
	time := utils.UnixMillisToTimeUTC(last_time)
	day := d.formatDate(time.Year(), int(time.Month()), time.Day())
	stmt := `INSERT INTO log_active_sessions_by_day (node, day, user_id, last_time)
        VALUES (?, ?, ?, ?)`
	return convErr(d.session.Query(stmt, node, day, user_id, last_time).Exec())
}

func (d *LogDAO) FindActiveSessions(node int, forDay time.Time) ([]*api.ActiveSessionInfoDTO, error) {

	checkSession(d.session)

	day := d.formatDate(forDay.Year(), int(forDay.Month()), forDay.Day())

	stmt := `SELECT user_id, last_time FROM log_active_sessions_by_day
        WHERE node = ? AND day = ?`

	activeSessions := make([]*api.ActiveSessionInfoDTO, 0, 16)

	var userID int64
	var lastTime int64
	iter := d.session.Query(stmt, node, day).Iter()

	for iter.Scan(&userID, &lastTime) {
		activeSessions = append(activeSessions, &api.ActiveSessionInfoDTO{
			Node:     node,
			UserID:   userID,
			LastTime: lastTime,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, convErr(err)
	}

	return activeSessions, nil
}

func (d *LogDAO) formatDate(year int, month int, day int) int {
	currentDateStr := fmt.Sprintf("%04d%02d%02d", year, month, day)
	num, _ := strconv.Atoi(currentDateStr)
	return num
}
