package cqldao

import (
	"github.com/gocql/gocql"
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
)

/*func (dao *EventDAO) InsertEventCAS(event *core.Event) (bool, error) {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, author_id, author_name, message, start_date,
		end_date, public, num_attendees, num_guests, created_date)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		IF NOT EXISTS`

	q := dao.session.Query(stmt,
		event.EventId, event.AuthorId, event.AuthorName, event.Message, event.StartDate,
		event.EndDate, event.IsPublic, event.NumAttendees, event.NumGuests, event.CreatedDate)

	return q.ScanCAS(nil)
}*/

/*func (dao *EventDAO) InsertEvent(event *core.Event) error {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, author_id, author_name, message, start_date,
		end_date, public, num_attendees, num_guests, created_date)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	q := dao.session.Query(stmt,
		event.EventId, event.AuthorId, event.AuthorName, event.Message, event.StartDate,
		event.EndDate, event.IsPublic, event.NumAttendees, event.NumGuests, event.CreatedDate)

	return q.Exec()
}*/

/*func (dao *EventDAO) AddOrUpdateParticipant(event_id int64, participant *core.EventParticipant) error {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
		VALUES (?, ?, ?, ?, ?)`

	q := dao.session.Query(stmt, event_id, participant.UserId, participant.Name,
		participant.Response, participant.Delivered)

	return q.Exec()
}*/

/*func (dao *EventDAO) AddOrUpdateParticipants(event_id int64, participantList map[int64]*core.EventParticipant) error {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
		VALUES (?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.UnloggedBatch) // Use unlogged batches when making updates to the same partition key.

	for _, participant := range participantList {
		batch.Query(stmt, event_id, participant.UserId, participant.Name,
			participant.Response, participant.Delivered)
	}

	return dao.session.ExecuteBatch(batch)
}*/

// FIXME: Each event of event table is in its own partition, classify events by date
// or something in order to improve read performance.
/*func (dao *EventDAO) LoadUserEvents(user_id int64, fromDate int64) ([]*core.Event, error) {

	dao.checkSession()

	// Get index of events
	toDate := core.TimeToMillis(time.Now().Add(core.MAX_DIF_IN_START_DATE)) // One year
	events_inbox, err := dao.LoadUserInbox(user_id, fromDate, toDate)

	if err != nil {
		log.Println("LoadUserEvents 1 (", user_id, "):", err)
		return nil, err
	}

	event_id_list := make([]int64, 0, len(events_inbox))
	for _, event_inbox := range events_inbox {
		event_id_list = append(event_id_list, event_inbox.EventId)
	}

	// Read from event table to get the actual info
	events, err := dao.LoadEvent(event_id_list...)

	if err != nil {
		log.Println("LoadUserEvents 2 (", user_id, "):", err)
		return nil, err
	}

	return events, nil
}*/

// Load events in closed range [fromDate, ...], where upper bound isn't constrained
// and it's higher than fromDate. Retrieved events are ordered from newer to older.
// This function returns MAX_EVENTS_IN_RECENT_LIST events as much.
func (d *EventDAO) loadUserInbox(user_id int64, fromDate int64) ([]*userEvent, error) {

	checkSession(d.session)

	stmt := `SELECT event_id, author_id, author_name, start_date, message,
		response FROM events_by_user
		WHERE user_id = ? AND event_bucket = ? AND start_date >= ? LIMIT ?`

	event_bucket := 1 // TODO: Add bucket logic
	query := d.session.Query(stmt, user_id, event_bucket, fromDate, MAX_EVENTS_IN_RECENT_LIST)

	return d.loadUserInboxHelper(query)
}

// Load events in open range [..., fromDate), where lower bound isn't constrained
// and it's lower than fromDate. fromDate must be a date lower than current_time.
// Retrieved events are ordered from newer to older. This function returns
// MAX_EVENTS_IN_HISTORY_LIST events as much.
func (d *EventDAO) loadUserInboxReverse(user_id int64, fromDate int64) ([]*userEvent, error) {

	checkSession(d.session)

	currentTime := utils.GetCurrentTimeMillis()

	if fromDate > currentTime {
		fromDate = currentTime
	}

	stmt := `SELECT event_id, author_id, author_name, start_date, message, response
		FROM events_by_user
		WHERE user_id = ? AND event_bucket = ? AND start_date < ? LIMIT ?`

	event_bucket := 1 // TODO: Add bucket logic
	query := d.session.Query(stmt, user_id, event_bucket, fromDate, MAX_EVENTS_IN_HISTORY_LIST)

	return d.loadUserInboxHelper(query)
}

// Load events in open range (fromDate, toDate), where fromDate < toDate.  Retrieved
// events are ordered from older to newer. toDate must be a date lower than current_time.
// This function returns MAX_EVENTS_IN_HISTORY_LIST as much.
func (d *EventDAO) loadUserInboxBetween(user_id int64, fromDate int64,
	toDate int64) ([]*userEvent, error) {

	checkSession(d.session)

	currentTime := utils.GetCurrentTimeMillis()

	if toDate > currentTime {
		toDate = currentTime
	}

	if fromDate >= toDate {
		return nil, api.ErrInvalidArg
	}

	stmt := `SELECT event_id, author_id, author_name, start_date, message, response
		FROM events_by_user
		WHERE user_id = ? AND event_bucket = ? AND start_date > ? AND start_date < ?
		ORDER BY start_date ASC LIMIT ?`

	event_bucket := 1 // TODO: Add bucket logic
	query := d.session.Query(stmt, user_id, event_bucket, fromDate, toDate, MAX_EVENTS_IN_HISTORY_LIST)

	return d.loadUserInboxHelper(query)
}

func (d *EventDAO) loadUserInboxHelper(query *gocql.Query) ([]*userEvent, error) {

	iter := query.Iter()
	events := make([]*userEvent, 0, 20)

	var event_id int64
	var author_id int64
	var author_name string
	var start_date int64
	var message string
	var response int32

	for iter.Scan(&event_id, &author_id, &author_name, &start_date, &message, &response) {

		events = append(events, &userEvent{
			EventId:    event_id,
			AuthorId:   author_id,
			AuthorName: author_name,
			StartDate:  start_date,
			Message:    message,
			Response:   api.AttendanceResponse(response),
		})
	}

	if err := iter.Close(); err != nil {
		return nil, convErr(err)
	}

	if len(events) == 0 {
		return nil, api.ErrNoResults
	}

	return events, nil
}
