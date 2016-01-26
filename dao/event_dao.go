package dao

import (
	"github.com/gocql/gocql"
	"log"
	core "peeple/areyouin/common"
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
func (dao *EventDAO) LoadParticipant(event_id uint64, user_id uint64) (*core.EventParticipant, error) {

	dao.checkSession()

	stmt := `SELECT guest_name, guest_response, guest_status FROM event
		WHERE event_id = ? AND guest_id = ? LIMIT 1`

	q := dao.session.Query(stmt, event_id, user_id)

	var name string
	var response, status int32
	var participant *core.EventParticipant

	err := q.Scan(&name, &response, &status)

	if err == nil {
		participant = &core.EventParticipant{
			UserId:    user_id,
			Name:      name,
			Response:  core.AttendanceResponse(response),
			Delivered: core.MessageStatus(status),
		}
	} else if err != gocql.ErrNotFound {
		log.Println("LoadParticipant:", err)
	}

	return participant, err
}

func (dao *EventDAO) LoadAllParticipants(event_id uint64) ([]*core.EventParticipant, error) {

	dao.checkSession()

	stmt := `SELECT guest_id, guest_name, guest_response, guest_status FROM event
		WHERE event_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, event_id, MAX_NUM_GUESTS).Iter()

	if iter == nil {
		return nil, ErrNilPointer
	}

	participants := make([]*core.EventParticipant, 0, 10)

	var user_id uint64
	var name string
	var response int32
	var status int32

	for iter.Scan(&user_id, &name, &response, &status) {
		participants = append(participants, &core.EventParticipant{
			UserId:    user_id,
			Name:      name,
			Response:  core.AttendanceResponse(response),
			Delivered: core.MessageStatus(status),
		})
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadParticipants (", event_id, "):", err)
		return nil, err
	}

	return participants, nil
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

	stmt_insert := `INSERT INTO user_events (user_id, event_id, end_date, response)
		VALUES (?, ?, ?, ?)`

	stmt_update := `UPDATE event SET guest_status = ? WHERE event_id = ? AND guest_id = ?`

	var response core.AttendanceResponse

	if event.AuthorId == user_id {
		response = core.AttendanceResponse_ASSIST
	} else {
		response = core.AttendanceResponse_NO_RESPONSE
	}

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(stmt_insert, user_id, event.EventId, event.EndDate, response)
	batch.Query(stmt_update, core.MessageStatus_SERVER_DELIVERED, event.EventId, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *EventDAO) LoadUserInbox(user_id uint64, fromDate int64) ([]uint64, error) {

	stmt := `SELECT event_id FROM user_events
		WHERE user_id = ? AND end_date >= ?`

	iter := dao.session.Query(stmt, user_id, fromDate).Iter()
	event_id_list := make([]uint64, 0, 20)
	var event_id uint64

	for iter.Scan(&event_id) {
		event_id_list = append(event_id_list, event_id)
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadUserInbox Error (user", user_id, "):", err)
		return nil, err
	}

	if len(event_id_list) == 0 {
		return nil, ErrEmptyInbox
	}

	return event_id_list, nil
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
	event_id_list, err := dao.LoadUserInbox(user_id, fromDate)

	if err != nil {
		log.Println("LoadUserEvents 1 (", user_id, "):", err)
		return nil, err
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
	event_id_list, err := dao.LoadUserInbox(user_id, fromDate)

	if err != nil {
		log.Println("LoadUserEventsAndParticipants 1 (", user_id, "):", err)
		return nil, err
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

func (dao *EventDAO) SetParticipantResponse(user_id uint64, event_id uint64, response core.AttendanceResponse) error {

	dao.checkSession()

	stmt := `UPDATE event SET guest_response = ? WHERE event_id = ? AND guest_id = ?`
	q := dao.session.Query(stmt, response, event_id, user_id)
	return q.Exec()
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
