package cqldao

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"time"

	"github.com/gocql/gocql"
)

const (
	MAX_EVENTS_IN_HISTORY_LIST = 15
)

type EventHistoryDAO struct {
	session *GocqlSession
}

func NewEventHistoryDAO(session api.DbSession) api.EventHistoryDAO {
	reconnectIfNeeded(session)
	return &EventHistoryDAO{session: session.(*GocqlSession)}
}

func (d *EventHistoryDAO) Insert(userID int64, entry *api.TimeLineEntryDTO) error {
	checkSession(d.session)
	stmt := `INSERT INTO events_history_by_user (user_id, position, event_id) VALUES (?, ?, ?)`
	q := d.session.Query(stmt, userID, entry.Position, entry.EventID)
	return convErr(q.Exec())
}

func (d *EventHistoryDAO) InsertBatch(entry *api.TimeLineEntryDTO, userIDs ...int64) error {

	checkSession(d.session)

	stmt := `INSERT INTO events_history_by_user (user_id, position, event_id) VALUES (?, ?, ?)`

	batch := d.session.NewBatch(gocql.LoggedBatch)

	for _, userID := range userIDs {
		batch.Query(stmt, userID, entry.Position, entry.EventID)
	}

	return convErr(d.session.ExecuteBatch(batch))
}

func (d *EventHistoryDAO) DeleteAll() error {
	checkSession(d.session)
	return d.session.Query(`TRUNCATE events_history_by_user`).Exec()
}

func (d *EventHistoryDAO) FindAllBackward(userID int64, fromDate time.Time) ([]int64, error) {

	checkSession(d.session)

	fromDateMillis := utils.TimeToMillis(fromDate)
	currentTimeMillis := utils.GetCurrentTimeMillis()

	if fromDateMillis > currentTimeMillis {
		fromDateMillis = currentTimeMillis
	}

	stmt := `SELECT event_id FROM events_history_by_user
        WHERE user_id = ? and position < ? LIMIT ?`

	query := d.session.Query(stmt, userID, fromDate, MAX_EVENTS_IN_HISTORY_LIST)

	return d.findAllAux(query)
}

func (d *EventHistoryDAO) FindAllForward(userID int64, fromDate time.Time) ([]int64, error) {

	checkSession(d.session)

	fromDateMillis := utils.TimeToMillis(fromDate)
	currentTimeMillis := utils.GetCurrentTimeMillis()

	if fromDateMillis > currentTimeMillis {
		fromDateMillis = currentTimeMillis
	}

	stmt := `SELECT event_id FROM events_history_by_user
        WHERE user_id = ? and position > ? ORDER BY position ASC LIMIT ?`

	query := d.session.Query(stmt, userID, fromDate, MAX_EVENTS_IN_HISTORY_LIST)

	return d.findAllAux(query)
}

func (d *EventHistoryDAO) findAllAux(query *gocql.Query) ([]int64, error) {

	iter := query.Iter()
	events := make([]int64, 0, MAX_EVENTS_IN_HISTORY_LIST)

	var eventID int64

	for iter.Scan(&eventID) {
		events = append(events, eventID)
	}

	if err := iter.Close(); err != nil {
		return nil, convErr(err)
	}

	if len(events) == 0 {
		return nil, api.ErrNoResults
	}

	return events, nil
}
