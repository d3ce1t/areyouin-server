package dao

import (
	"log"
	core "peeple/areyouin/common"
	"github.com/gocql/gocql"
)

const (
	MAX_NUM_GUESTS = 100
	MAX_EVENTS_IN_HISTORY_LIST = 15
	MAX_EVENTS_IN_RECENT_LIST = 100
	//TIME_MARGIN_IN_SEC = 4 * 3600 // 4 hours
)

type EventDAO struct {
	session *gocql.Session
}

func (dao *EventDAO) GetSession() *gocql.Session {
	return dao.session
}

func (dao *EventDAO) InsertEventAndParticipants(event *core.Event) error {

	checkSession(dao)

	stmt_event := `INSERT INTO event (event_id, author_id, author_name, message, start_date,
		end_date, public, num_attendees, num_guests, created_date, inbox_position)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt_participant := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
			VALUES (?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.UnloggedBatch) // Use unlogged batches when making updates to the same partition key.

	batch.Query(stmt_event, event.EventId, event.AuthorId, event.AuthorName, event.Message, event.StartDate,
		event.EndDate, event.IsPublic, event.NumAttendees, event.NumGuests, event.CreatedDate, event.InboxPosition)

	for _, participant := range event.Participants {
		batch.Query(stmt_participant, event.EventId, participant.UserId, participant.Name,
			participant.Response, participant.Delivered)
	}

	return dao.session.ExecuteBatch(batch)
}

// Load the participant of an event and returns it. If not found returns a nil participant
// and error. Whatever else returns (nil, error)
func (dao *EventDAO) LoadParticipant(event_id int64, user_id int64) (*core.Participant, error) {

	checkSession(dao)

	stmt := `SELECT guest_name, guest_response, guest_status, start_date, event_state FROM event
		WHERE event_id = ? AND guest_id = ? LIMIT 1`

	q := dao.session.Query(stmt, event_id, user_id)

	var name string
	var response, status int32
	var start_date int64
	var event_state int32
	var participant *core.Participant

	err := q.Scan(&name, &response, &status, &start_date, &event_state)

	if err == nil {
		participant = &core.Participant{
			EventParticipant: core.EventParticipant{
				UserId:    user_id,
				Name:      name,
				Response:  core.AttendanceResponse(response),
				Delivered: core.MessageStatus(status),
			},
			EventId:    event_id,
			StartDate:  start_date,
			EventState: core.EventState(event_state),
		}

	} else if err == gocql.ErrNotFound {
		err = ErrNotFoundEventOrParticipant
	} else {
		log.Println("LoadParticipant:", err)
	}

	return participant, err
}

// Adds an event to participant inbox and also adds the participant into the event participant list.
// If participant already exists, it is replaced. If not, participant is created
func (dao *EventDAO) AddOrUpdateEventToUserInbox(participant *core.EventParticipant, event *core.Event) error {

	checkSession(dao)

	stmt_insert := `INSERT INTO events_by_user (user_id, event_bucket, start_date, event_id, author_id, author_name, message, response)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	stmt_event_update := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status) VALUES (?, ?, ?, ?, ?)`

	if event.AuthorId == participant.UserId {
		participant.Response = core.AttendanceResponse_ASSIST
	}

	participant.Delivered = core.MessageStatus_SERVER_DELIVERED
	batch := dao.session.NewBatch(gocql.LoggedBatch)
	event_bucket := 1 // TODO: Implement bucket logic properly

	batch.Query(stmt_insert, participant.UserId, event_bucket, event.StartDate, event.EventId, event.AuthorId,
		event.AuthorName, event.Message, participant.Response)
	batch.Query(stmt_event_update, event.EventId, participant.UserId, participant.Name, participant.Response, participant.Delivered)

	return dao.session.ExecuteBatch(batch)
}

// Adds an event to participant inbox and also updates the participant delivery info in the event participant list.
// Whereas the above function is used whenever a new participant is invited to an existing event. This one is used
// when the event is first created. The main difference is that the second statement only updates guest_status and
// not all of the fields like in the above function. This function assumes that when the event is first created it
// already includes participants, so when inserting into the inbox it is needed to only update guest_status of each
// participant
func (dao *EventDAO) InsertEventToUserInbox(participant *core.EventParticipant, event *core.Event) error {

	checkSession(dao)

	stmt_insert := `INSERT INTO events_by_user (user_id, event_bucket, start_date, event_id, author_id, author_name, message, response)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	stmt_update := `UPDATE event SET guest_status = ? WHERE event_id = ? AND guest_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	event_bucket := 1 // TODO: Implement bucket logic properly

	batch.Query(stmt_insert, participant.UserId, event_bucket, event.StartDate, event.EventId, event.AuthorId,
		event.AuthorName, event.Message, participant.Response)
	batch.Query(stmt_update, participant.Delivered, event.EventId, participant.UserId)

	return dao.session.ExecuteBatch(batch)
}

// Load events in closed range [fromDate, ...], where upper bound isn't constrained
// and it's higher than fromDate. Retrieved events are ordered from newer to older.
// This function returns MAX_EVENTS_IN_RECENT_LIST events as much.
func (dao *EventDAO) LoadUserInbox(user_id int64, fromDate int64) ([]*core.EventInbox, error) {

		checkSession(dao)

		stmt := `SELECT event_id, author_id, author_name, start_date, message, response FROM events_by_user
			WHERE user_id = ? AND event_bucket = ? AND start_date >= ? LIMIT ?`

		event_bucket := 1 // TODO: Add bucket logic
		query := dao.session.Query(stmt, user_id, event_bucket, fromDate, MAX_EVENTS_IN_RECENT_LIST)

		return dao.loadUserInboxHelper(query)
}

// Load events in open range [..., fromDate), where lower bound isn't constrained
// and it's lower than fromDate. fromDate must be a date lower than current_time.
// Retrieved events are ordered from newer to older. This function returns
// MAX_EVENTS_IN_HISTORY_LIST events as much.
func (dao *EventDAO) LoadUserInboxReverse(user_id int64, fromDate int64) ([]*core.EventInbox, error) {

	checkSession(dao)

	currentTime := core.GetCurrentTimeMillis()

	if fromDate > currentTime {
		fromDate = currentTime
	}

	stmt := `SELECT event_id, author_id, author_name, start_date, message, response FROM events_by_user
		WHERE user_id = ? AND event_bucket = ? AND start_date < ? LIMIT ?`

	event_bucket := 1 // TODO: Add bucket logic
	query := dao.session.Query(stmt, user_id, event_bucket, fromDate, MAX_EVENTS_IN_HISTORY_LIST)

	return dao.loadUserInboxHelper(query)
}

// Load events in open range (fromDate, toDate), where fromDate < toDate.  Retrieved
// events are ordered from older to newer. toDate must be a date lower than current_time.
// This function returns MAX_EVENTS_IN_HISTORY_LIST as much.
func (dao *EventDAO) LoadUserInboxBetween(user_id int64, fromDate int64, toDate int64) ([]*core.EventInbox, error) {

	checkSession(dao)

	currentTime := core.GetCurrentTimeMillis()

	if toDate > currentTime {
		toDate = currentTime
	}

	if fromDate >= toDate {
		return nil, ErrInvalidArg
	}

	stmt := `SELECT event_id, author_id, author_name, start_date, message, response FROM events_by_user
		WHERE user_id = ? AND event_bucket = ? AND start_date > ? AND start_date < ?
		ORDER BY start_date ASC LIMIT ?`

	event_bucket := 1 // TODO: Add bucket logic
	query := dao.session.Query(stmt, user_id, event_bucket, fromDate, toDate, MAX_EVENTS_IN_HISTORY_LIST)

	return dao.loadUserInboxHelper(query)
}

// Read one or more events and their participants. Event info and participants are in the
// same partition
func (dao *EventDAO) LoadEventAndParticipants(event_ids ...int64) (events []*core.Event, err error) {

	checkSession(dao)

	if event_ids == nil || len(event_ids) == 0 {
		return nil, ErrInvalidArg
	}

	// Keep input order in output slice. Because of this, it's need to index which
	// position takes each id in the input vector
	eventPosition := make(map[int64]int)
	for pos, value := range event_ids {
		eventPosition[value] = pos
	}

	events = make([]*core.Event, len(event_ids))

	stmt := `SELECT event_id, author_id, author_name, message, picture_digest, created_date, inbox_position, start_date,
									end_date, num_attendees, num_guests, event_state, guest_id, guest_name,
									guest_response, guest_status
						FROM event WHERE event_id IN (` + core.GenParams(len(event_ids)) + `)`

	values := core.GenValues(event_ids)
	iter := dao.session.Query(stmt, values...).Iter()

	var event_id int64
	var author_id int64
	var author_name string
	var message string
	var digest []byte
	var created_date int64
	var inbox_position int64
	var start_date int64
	var end_date int64
	var num_attendees int32
	var num_guests int32
	var guest_id int64
	var guest_name string
	var guest_response int32
	var guest_status int32
	var event_state int32

	events_index := make(map[int64]*core.Event)

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&event_id, &author_id, &author_name, &message, &digest, &created_date, &inbox_position, &start_date,
		&end_date, &num_attendees, &num_guests, &event_state, &guest_id, &guest_name, &guest_response, &guest_status) {

		event, ok := events_index[event_id]

		if !ok {
			event = &core.Event{
				EventId:       event_id,
				AuthorId:      author_id,
				AuthorName:    author_name,
				InboxPosition: inbox_position,
				StartDate:     start_date,
				EndDate:       end_date,
				Message:       message,
				PictureDigest: digest,
				IsPublic:      false,
				NumAttendees:  num_attendees,
				NumGuests:     num_guests,
				CreatedDate:   created_date,
				State:         core.EventState(event_state),
				Participants:  make(map[int64]*core.EventParticipant),
			}

			events_index[event_id] = event
			events[eventPosition[event.EventId]] = event
		}

		if guest_id != 0 {
			guest := &core.EventParticipant{
				UserId:    guest_id,
				Name:      guest_name,
				Response:  core.AttendanceResponse(guest_response),
				Delivered: core.MessageStatus(guest_status),
			}

			event.Participants[guest_id] = guest
		}
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadEventAndParticipants Error:", err)
		return nil, err
	}

	return events, nil
}

// Read one or more events and do NOT include participants
func (dao *EventDAO) LoadEvent(event_ids ...int64) (events []*core.Event, err error) {

	checkSession(dao)

	if event_ids == nil || len(event_ids) == 0 {
		return nil, ErrInvalidArg
	}

	events = make([]*core.Event, 0, len(event_ids))

	stmt := `SELECT DISTINCT event_id, author_id, author_name, message, picture_digest, created_date, inbox_position,
									start_date,	end_date, num_attendees, num_guests, event_state
						FROM event WHERE event_id IN (` + core.GenParams(len(event_ids)) + `)`

	values := core.GenValues(event_ids)
	iter := dao.session.Query(stmt, values...).Iter()

	var event_id int64
	var author_id int64
	var author_name string
	var message string
	var digest []byte
	var created_date int64
	var inbox_position int64
	var start_date int64
	var end_date int64
	var num_attendees int32
	var num_guests int32
	var event_state int32

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&event_id, &author_id, &author_name, &message, &digest, &created_date, &inbox_position,
		&start_date, &end_date, &num_attendees, &num_guests, &event_state) {

		event := &core.Event{
			EventId:       event_id,
			AuthorId:      author_id,
			AuthorName:    author_name,
			InboxPosition: inbox_position,
			StartDate:     start_date,
			EndDate:       end_date,
			Message:       message,
			PictureDigest: digest,
			IsPublic:      false,
			NumAttendees:  num_attendees,
			NumGuests:     num_guests,
			CreatedDate:   created_date,
			State:         core.EventState(event_state),
		}

		events = append(events, event)
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadEvent Error:", err)
		return nil, err
	}

	return events, nil
}

// FIXME: Each event of event table is in its own partition, classify events by date
// or something in order to improve read performance.
func (dao *EventDAO) LoadUserEventsAndParticipants(user_id int64, fromDate int64) ([]*core.Event, error) {

	checkSession(dao)

	// Get index of events
	//toDate := core.TimeToMillis(time.Now().Add(core.MAX_DIF_IN_START_DATE)) // One year
	events_inbox, err := dao.LoadUserInbox(user_id, fromDate)

	if err != nil {
		return nil, err
	}

	event_id_list := make([]int64, 0, len(events_inbox))
	for _, event_inbox := range events_inbox {
		event_id_list = append(event_id_list, event_inbox.EventId)
	}

	// Read from event table to get the actual info
	events, err := dao.LoadEventAndParticipants(event_id_list...)

	if err != nil {
		log.Println("LoadUserEventsAndParticipants 2 (", user_id, "):", err)
		return nil, err
	}

	return events, nil
}


// Load user events in window delimited by (fromDate, toDate) where fromDate < toDate
func (dao *EventDAO) LoadUserEventsHistoryAndparticipants(user_id int64, fromDate int64, toDate int64) ([]*core.Event, error) {

	checkSession(dao)

	var events_inbox []*core.EventInbox
	var err error

	if fromDate < toDate {
		events_inbox, err = dao.LoadUserInboxBetween(user_id, fromDate, toDate)
	} else {
		events_inbox, err = dao.LoadUserInboxReverse(user_id, fromDate)
	}

	if err != nil {
		log.Println("LoadUserEventsHistoryAndparticipants 1 (", user_id, "):", err)
		return nil, err
	}

	event_id_list := make([]int64, 0, len(events_inbox))
	for _, event_inbox := range events_inbox {
		event_id_list = append(event_id_list, event_inbox.EventId)
	}

	// Read from event table to get the actual info
	events, err := dao.LoadEventAndParticipants(event_id_list...)

	if err != nil {
		log.Println("LoadUserEventsHistoryAndparticipants 2 (", user_id, "):", err)
		return nil, err
	}

	/*for i := 0; i<len(events_inbox); i++ {
		fmt.Printf("StartDate: %v Id: %v, Id: %v\n", core.UnixMillisToTime(events_inbox[i].StartDate), events_inbox[i].EventId, events[i].EventId)
	}*/


	return events, nil
}

// Removes and insert the same row into events_by_user but change position. This is the only way
// to do this in Cassandra because of position being part of the primary key.
/*func (dao *EventDAO) SetUserEventInboxPosition(participant *core.EventParticipant, event *core.Event, new_position int64) error {
	checkSession(dao)

	stmt_remove := `DELETE FROM events_by_user WHERE user_id = ? AND event_bucket = ? AND position = ? AND event_id = ?`
	stmt_insert := `INSERT INTO events_by_user (user_id, event_bucket, position, event_id, author_id, author_name, start_date, message, response)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(stmt_remove, participant.UserId, 1, event.InboxPosition, event.EventId)
	batch.Query(stmt_insert, participant.UserId, 1, new_position, event.EventId, event.AuthorId, event.AuthorName, event.StartDate, event.Message, participant.Response)

	return dao.session.ExecuteBatch(batch)
}*/

func (dao *EventDAO) SetEventStateAndInboxPosition(event_id int64, new_status core.EventState, new_position int64) error {

	checkSession(dao)

	stmt := `UPDATE event SET inbox_position = ?, event_state = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, new_position, new_status, event_id)

	return q.Exec()
}

func (dao *EventDAO) SetEventPicture(event_id int64, picture *core.Picture) error {

	checkSession(dao)

	stmt := `UPDATE event SET picture = ?, picture_digest = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, picture.RawData, picture.Digest, event_id)

	return q.Exec()
}

func (dao *EventDAO) LoadEventPicture(event_id int64) ([]byte, error) {

	checkSession(dao)

	stmt := `SELECT picture FROM event WHERE event_id = ?`
	q := dao.session.Query(stmt, event_id)

	var picture []byte

	err := q.Scan(&picture)
	if err != nil {
		return nil, err
	}

	return picture, nil
}

// Compare-and-set (read-before) update operation
func (dao *EventDAO) CompareAndSetNumGuests(event_id int64, num_guests int) (ok bool, err error) {

	checkSession(dao)

	read_stmt := `SELECT num_guests FROM event WHERE event_id = ?`
	q := dao.session.Query(read_stmt, event_id)

	var old_num_guests int32
	var write_stmt string

	if err := q.Scan(&old_num_guests); err != nil {
		return false, err
	}

	write_stmt = `UPDATE event SET num_guests = ? WHERE event_id = ?
								IF num_guests = ?`

	q = dao.session.Query(write_stmt, num_guests, event_id, old_num_guests)
	return q.ScanCAS(nil)
}

/*func (dao *EventDAO) SetNumGuests(event_id int64, num_guests int32) error {

	checkSession(dao)

	stmt := `UPDATE event SET num_guests = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_guests, event_id)
	return q.Exec()
}*/

// Compare-and-set (read-before) update operation
func (dao *EventDAO) CompareAndSetNumAttendees(event_id int64, num_attendees int) (ok bool, err error) {

	checkSession(dao)

	read_stmt := `SELECT num_attendees FROM event WHERE event_id = ?`
	q := dao.session.Query(read_stmt, event_id)

	var old_attendees int32
	var write_stmt string

	if err := q.Scan(&old_attendees); err != nil {
		return false, err
	}

	write_stmt = `UPDATE event SET num_attendees = ? WHERE event_id = ?
								IF num_attendees = ?`

	q = dao.session.Query(write_stmt, num_attendees, event_id, old_attendees)

	return q.ScanCAS(nil)
}

/*func (dao *EventDAO) SetNumAttendees(event_id int64, num_attendees int) error {

	checkSession(dao)

	stmt := `UPDATE event SET num_attendees = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_attendees, event_id)
	return q.Exec()
}*/

func (dao *EventDAO) SetParticipantStatus(user_id int64, event_id int64, status core.MessageStatus) error {

	checkSession(dao)

	stmt := `UPDATE event SET guest_status = ? WHERE event_id = ? AND guest_id = ?`
	q := dao.session.Query(stmt, status, event_id, user_id)

	return q.Exec()
}

func (dao *EventDAO) SetParticipantResponse(participant *core.Participant, response core.AttendanceResponse) error {

	checkSession(dao)

	stmt_event := `UPDATE event SET guest_response = ? WHERE event_id = ? AND guest_id = ?`
	stmt_events_by_user := `UPDATE events_by_user SET response = ?
		WHERE user_id = ? AND event_bucket = ? AND start_date = ? AND event_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(stmt_event, response, participant.EventId, participant.UserId)
	batch.Query(stmt_events_by_user, response, participant.UserId, 1, participant.StartDate, participant.EventId)

	return dao.session.ExecuteBatch(batch)
}
