package dao

import (
	core "peeple/areyouin/common"
	"github.com/gocql/gocql"
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

/*func (dao *EventDAO) AddOrUpdateParticipant(event_id uint64, participant *core.EventParticipant) error {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
		VALUES (?, ?, ?, ?, ?)`

	q := dao.session.Query(stmt, event_id, participant.UserId, participant.Name,
		participant.Response, participant.Delivered)

	return q.Exec()
}*/

/*func (dao *EventDAO) AddOrUpdateParticipants(event_id uint64, participantList map[uint64]*core.EventParticipant) error {

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
/*func (dao *EventDAO) LoadUserEvents(user_id uint64, fromDate int64) ([]*core.Event, error) {

	dao.checkSession()

	// Get index of events
	toDate := core.TimeToMillis(time.Now().Add(core.MAX_DIF_IN_START_DATE)) // One year
	events_inbox, err := dao.LoadUserInbox(user_id, fromDate, toDate)

	if err != nil {
		log.Println("LoadUserEvents 1 (", user_id, "):", err)
		return nil, err
	}

	event_id_list := make([]uint64, 0, len(events_inbox))
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

func (dao *EventDAO) loadUserInboxHelper(query *gocql.Query) ([]*core.EventInbox, error) {

	iter := query.Iter()
	events := make([]*core.EventInbox, 0, 20)

	var event_id uint64
	var author_id uint64
	var author_name string
	var start_date int64
	var message string
	var response int32

	for iter.Scan(&event_id, &author_id, &author_name, &start_date, &message, &response) {

		events = append(events, &core.EventInbox{
			EventId:    event_id,
			AuthorId:   author_id,
			AuthorName: author_name,
			StartDate:  start_date,
			Message:    message,
			Response:   core.AttendanceResponse(response),
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, ErrEmptyInbox
	}

	return events, nil
}

func (dao *EventDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
