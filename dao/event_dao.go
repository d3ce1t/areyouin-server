package dao

import (
	"github.com/gocql/gocql"
	"log"
	core "peeple/areyouin/common"
	"time"
)

const (
	MAX_NUM_GUESTS = 100
	//TIME_MARGIN_IN_SEC = 4 * 3600 // 4 hours
)

type EventDAO struct {
	session *gocql.Session
}

func NewEventDAO(session *gocql.Session) core.EventDAO {
	return &EventDAO{session: session}
}

func (dao *EventDAO) InsertEventCAS(event *core.Event) (bool, error) {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, author_id, author_name, message, start_date,
		end_date, public, num_attendees, num_guests, created_date)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		IF NOT EXISTS`

	q := dao.session.Query(stmt,
		event.EventId, event.AuthorId, event.AuthorName, event.Message, event.StartDate,
		event.EndDate, event.IsPublic, event.NumAttendees, event.NumGuests, event.CreatedDate)

	return q.ScanCAS(nil)
}

func (dao *EventDAO) InsertEvent(event *core.Event) error {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, author_id, author_name, message, start_date,
		end_date, public, num_attendees, num_guests, created_date)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	q := dao.session.Query(stmt,
		event.EventId, event.AuthorId, event.AuthorName, event.Message, event.StartDate,
		event.EndDate, event.IsPublic, event.NumAttendees, event.NumGuests, event.CreatedDate)

	return q.Exec()
}

func (dao *EventDAO) InsertEventAndParticipants(event *core.Event) error {

	dao.checkSession()

	stmt_event := `INSERT INTO event (event_id, author_id, author_name, message, start_date,
		end_date, public, num_attendees, num_guests, created_date)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt_participant := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
			VALUES (?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.UnloggedBatch) // Use unlogged batches when making updates to the same partition key.

	batch.Query(stmt_event, event.EventId, event.AuthorId, event.AuthorName, event.Message, event.StartDate,
		event.EndDate, event.IsPublic, event.NumAttendees, event.NumGuests, event.CreatedDate)

	for _, participant := range event.Participants {
		batch.Query(stmt_participant, event.EventId, participant.UserId, participant.Name,
			participant.Response, participant.Delivered)
	}

	return dao.session.ExecuteBatch(batch)
}

// Load the participant of an event and returns it. If not found returns a nil participant
// and error. Whatever else returns (nil, error)
func (dao *EventDAO) LoadParticipant(event_id uint64, user_id uint64) (*core.Participant, error) {

	dao.checkSession()

	stmt := `SELECT guest_name, guest_response, guest_status, start_date FROM event
		WHERE event_id = ? AND guest_id = ? LIMIT 1`

	q := dao.session.Query(stmt, event_id, user_id)

	var name string
	var response, status int32
	var start_date int64
	var participant *core.Participant

	err := q.Scan(&name, &response, &status, &start_date)

	if err == nil {
		participant = &core.Participant{
			EventParticipant: core.EventParticipant{
				UserId:    user_id,
				Name:      name,
				Response:  core.AttendanceResponse(response),
				Delivered: core.MessageStatus(status),
			},
			EventId:        event_id,
			EventStartDate: start_date,
		}

	} else if err != gocql.ErrNotFound {
		log.Println("LoadParticipant:", err)
	}

	return participant, err
}

func (dao *EventDAO) AddOrUpdateParticipant(event_id uint64, participant *core.EventParticipant) error {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
		VALUES (?, ?, ?, ?, ?)`

	q := dao.session.Query(stmt, event_id, participant.UserId, participant.Name,
		participant.Response, participant.Delivered)

	return q.Exec()
}

func (dao *EventDAO) AddOrUpdateParticipants(event_id uint64, participantList []*core.EventParticipant) error {

	dao.checkSession()

	stmt := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
		VALUES (?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.UnloggedBatch) // Use unlogged batches when making updates to the same partition key.

	for _, participant := range participantList {
		batch.Query(stmt, event_id, participant.UserId, participant.Name,
			participant.Response, participant.Delivered)
	}

	return dao.session.ExecuteBatch(batch)
}

func (dao *EventDAO) AddEventToUserInbox(user_id uint64, event *core.Event) error {

	dao.checkSession()

	stmt_insert := `INSERT INTO events_by_user (user_id, event_bucket, start_date, event_id, author_id, author_name, message, response)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	stmt_update := `UPDATE event SET guest_status = ? WHERE event_id = ? AND guest_id = ?`

	var response core.AttendanceResponse

	if event.AuthorId == user_id {
		response = core.AttendanceResponse_ASSIST
	} else {
		response = core.AttendanceResponse_NO_RESPONSE
	}

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	event_bucket := 1 // TODO: Implement bucket logic properly

	batch.Query(stmt_insert, user_id, event_bucket, event.StartDate, event.EventId, event.AuthorId,
		event.AuthorName, event.Message, response)
	batch.Query(stmt_update, core.MessageStatus_SERVER_DELIVERED, event.EventId, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *EventDAO) LoadUserInbox(user_id uint64, fromDate int64, toDate int64) ([]*core.EventInbox, error) {

	stmt := `SELECT event_id, author_id, author_name, start_date, message, response FROM events_by_user
		WHERE user_id = ? AND event_bucket = ? AND start_date >= ? AND start_date < ?`

	event_bucket := 1 // TODO: Add bucket logic
	iter := dao.session.Query(stmt, user_id, event_bucket, fromDate, toDate).Iter()
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
		log.Println("LoadUserInbox Error (user", user_id, "):", err)
		return nil, err
	}

	if len(events) == 0 {
		return nil, ErrEmptyInbox
	}

	return events, nil
}

// Read one or more events and their participants. Event info and participants are in the
// same partition
func (dao *EventDAO) LoadEventAndParticipants(event_ids ...uint64) (events []*core.Event, err error) {

	if event_ids == nil || len(event_ids) == 0 {
		return nil, ErrInvalidArg
	}

	events = make([]*core.Event, 0, len(event_ids))

	stmt := `SELECT event_id, author_id, author_name, message, created_date, start_date,
									end_date, num_attendees, num_guests, guest_id, guest_name,
									guest_response, guest_status
						FROM event WHERE event_id IN (` + GenParams(len(event_ids)) + `)`

	values := GenValues(event_ids)
	iter := dao.session.Query(stmt, values...).Iter()

	var event_id uint64
	var author_id uint64
	var author_name string
	var message string
	var created_date int64
	var start_date int64
	var end_date int64
	var num_attendees int32
	var num_guests int32
	var guest_id uint64
	var guest_name string
	var guest_response, guest_status int32

	events_index := make(map[uint64]*core.Event)

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&event_id, &author_id, &author_name, &message, &created_date, &start_date,
		&end_date, &num_attendees, &num_guests, &guest_id, &guest_name, &guest_response, &guest_status) {

		event, ok := events_index[event_id]

		if !ok {
			event = &core.Event{
				EventId:      event_id,
				AuthorId:     author_id,
				AuthorName:   author_name,
				StartDate:    start_date,
				EndDate:      end_date,
				Message:      message,
				IsPublic:     false,
				NumAttendees: num_attendees,
				NumGuests:    num_guests,
				CreatedDate:  created_date,
				Participants: make(map[uint64]*core.EventParticipant),
			}

			events_index[event_id] = event
			events = append(events, event)
		}

		guest := &core.EventParticipant{
			UserId:    guest_id,
			Name:      guest_name,
			Response:  core.AttendanceResponse(guest_response),
			Delivered: core.MessageStatus(guest_status),
		}

		event.Participants[guest_id] = guest
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadEventAndParticipants Error:", err)
		return nil, err
	}

	return events, nil
}

// Read one or more events and their participants. Event info and participants are in the
// same partition
func (dao *EventDAO) LoadEvent(event_ids ...uint64) (events []*core.Event, err error) {

	if event_ids == nil || len(event_ids) == 0 {
		return nil, ErrInvalidArg
	}

	events = make([]*core.Event, 0, len(event_ids))

	stmt := `SELECT DISTINCT event_id, author_id, author_name, message, created_date, start_date,
									end_date, num_attendees, num_guests
						FROM event WHERE event_id IN (` + GenParams(len(event_ids)) + `)`

	values := GenValues(event_ids)
	iter := dao.session.Query(stmt, values...).Iter()

	var event_id uint64
	var author_id uint64
	var author_name string
	var message string
	var created_date int64
	var start_date int64
	var end_date int64
	var num_attendees int32
	var num_guests int32

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&event_id, &author_id, &author_name, &message, &created_date, &start_date,
		&end_date, &num_attendees, &num_guests) {

		event := &core.Event{
			EventId:      event_id,
			AuthorId:     author_id,
			AuthorName:   author_name,
			StartDate:    start_date,
			EndDate:      end_date,
			Message:      message,
			IsPublic:     false,
			NumAttendees: num_attendees,
			NumGuests:    num_guests,
			CreatedDate:  created_date,
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
func (dao *EventDAO) LoadUserEvents(user_id uint64, fromDate int64) ([]*core.Event, error) {

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
}

// FIXME: Each event of event table is in its own partition, classify events by date
// or something in order to improve read performance.
func (dao *EventDAO) LoadUserEventsAndParticipants(user_id uint64, fromDate int64) ([]*core.Event, error) {

	dao.checkSession()

	// Get index of events
	toDate := core.TimeToMillis(time.Now().Add(core.MAX_DIF_IN_START_DATE)) // One year
	events_inbox, err := dao.LoadUserInbox(user_id, fromDate, toDate)

	if err != nil {
		log.Println("LoadUserEventsAndParticipants 1 (", user_id, "):", err)
		return nil, err
	}

	event_id_list := make([]uint64, 0, len(events_inbox))
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

// Compare-and-set (read-before) update operation
func (dao *EventDAO) CompareAndSetNumGuests(event_id uint64, num_guests int32) (ok bool, err error) {

	dao.checkSession()

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

func (dao *EventDAO) SetNumGuests(event_id uint64, num_guests int32) error {

	dao.checkSession()

	stmt := `UPDATE event SET num_guests = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_guests, event_id)
	return q.Exec()
}

// Compare-and-set (read-before) update operation
func (dao *EventDAO) CompareAndSetNumAttendees(event_id uint64, num_attendees int) (ok bool, err error) {

	dao.checkSession()

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

func (dao *EventDAO) SetNumAttendees(event_id uint64, num_attendees int) error {

	dao.checkSession()

	stmt := `UPDATE event SET num_attendees = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_attendees, event_id)
	return q.Exec()
}

func (dao *EventDAO) SetParticipantStatus(user_id uint64, event_id uint64, status core.MessageStatus) error {

	dao.checkSession()

	stmt := `UPDATE event SET guest_status = ? WHERE event_id = ? AND guest_id = ?`
	q := dao.session.Query(stmt, status, event_id, user_id)
	return q.Exec()
}

func (dao *EventDAO) SetParticipantResponse(participant *core.Participant, response core.AttendanceResponse) error {

	dao.checkSession()

	stmt_event := `UPDATE event SET guest_response = ? WHERE event_id = ? AND guest_id = ?`
	stmt_events_by_user := `UPDATE events_by_user SET response = ? WHERE user_id = ? AND event_bucket = ? AND start_date = ? AND event_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(stmt_event, response, participant.EventId, participant.UserId)
	batch.Query(stmt_events_by_user, response, participant.UserId, 1, participant.EventStartDate, participant.EventId)

	return dao.session.ExecuteBatch(batch)
}

func GenParams(size int) string {
	result := "?"
	for i := 1; i < size; i++ {
		result += ", ?"
	}
	return result
}

func GenValues(values []uint64) []interface{} {

	result := make([]interface{}, 0, len(values))

	for _, val := range values {
		result = append(result, val)
	}

	return result
}

func (dao *EventDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
