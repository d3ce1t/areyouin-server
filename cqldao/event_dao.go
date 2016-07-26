package cqldao

import (
	"github.com/gocql/gocql"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
)

const (
	MAX_NUM_GUESTS             = 100
	MAX_EVENTS_IN_HISTORY_LIST = 15
	MAX_EVENTS_IN_RECENT_LIST  = 100
)

type EventDAO struct {
	session *GocqlSession
}

// TODO: This method is not consistent with AddParticipantToEvent, i.e., it inserts
// the event into cassandra and, if there is any, inserts also participants. But
// it doesn't insert the event into events_by_user.
func (d *EventDAO) Insert(event *api.EventDTO) error {

	checkSession(d.session)

	stmt_event := `INSERT INTO event (event_id, author_id, author_name, message,
		start_date, end_date, num_attendees, num_guests, created_date,
		inbox_position, event_state)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var status int32
	if event.Cancelled {
		status = 3
	}

	var err error

	if len(event.Participants) > 0 {

		stmt_participant := `INSERT INTO event (event_id, guest_id, guest_name,
			guest_response, guest_status)	VALUES (?, ?, ?, ?, ?)`

		// Use unlogged batches when making updates to the same partition key.
		batch := d.session.NewBatch(gocql.UnloggedBatch)

		batch.Query(stmt_event, event.Id, event.AuthorId, event.AuthorName,
			event.Description, event.StartDate, event.EndDate, event.NumAttendees,
			event.NumGuests, event.CreatedDate, event.InboxPosition, status)

		for _, p := range event.Participants {
			batch.Query(stmt_participant, event.Id, p.UserId, p.Name, p.Response,
				p.InvitationStatus)
		}

		err = d.session.ExecuteBatch(batch)

	} else {

		q := d.session.Query(stmt_event, event.Id, event.AuthorId, event.AuthorName,
			event.Description, event.StartDate, event.EndDate, event.NumAttendees,
			event.NumGuests, event.CreatedDate, event.InboxPosition, status)

		err = q.Exec()

	}

	return convErr(err)
}

// Load the participant of an event and returns it. If not found returns a nil participant
// and error. Whatever else returns (nil, error)
/*func (dao *EventDAO) LoadParticipant(event_id int64, user_id int64) (*core.Participant, error) {

	checkSession(dao.session)

	stmt := `SELECT guest_name, guest_response, guest_status, start_date, event_state FROM event
		WHERE event_id = ? AND guest_id = ? LIMIT 1`

	q := dao.session.Query(stmt, event_id, user_id)

	var name string
	var response, status int32
	var start_date int64
	var event_state int32

	err := q.Scan(&name, &response, &status, &start_date, &event_state)

	if err == nil {
		participant := &core.Participant{
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

		return participant, nil

	} else {
		return nil, err
	}
}*/

// Adds a participant to an existing event. In cassandra, this implies to add
// the participant to event and to events_by_user. If participant with same id
// already exists it gets overwritten.
func (dao *EventDAO) AddParticipantToEvent(participant *api.ParticipantDTO, event *api.EventDTO) error {

	checkSession(dao.session)

	stmt_insert := `INSERT INTO events_by_user (user_id, event_bucket, start_date,
		event_id, author_id, author_name, message, response)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	stmt_event_update := `INSERT INTO event (event_id, guest_id, guest_name,
		guest_response, guest_status) VALUES (?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	event_bucket := 1 // TODO: Implement bucket logic properly

	batch.Query(stmt_insert, participant.UserId, event_bucket, event.StartDate, event.Id,
		event.AuthorId, event.AuthorName, event.Description, participant.Response)

	batch.Query(stmt_event_update, event.Id, participant.UserId, participant.Name,
		participant.Response, participant.InvitationStatus)

	return convErr(dao.session.ExecuteBatch(batch))
}

// Read one or more events and their participants. Event info and participants are in the
// same partition
func (ed *EventDAO) LoadEvents(event_ids ...int64) (events []*api.EventDTO, err error) {

	checkSession(ed.session)

	if event_ids == nil || len(event_ids) == 0 {
		return nil, api.ErrInvalidArg
	}

	stmt := `SELECT event_id, author_id, author_name, message, picture_digest,
		created_date, inbox_position, start_date, end_date, num_attendees, num_guests,
		event_state, guest_id, guest_name, guest_response, guest_status
		FROM event WHERE event_id IN (` + utils.GenParams(len(event_ids)) + `)`

	values := utils.GenValues(event_ids)
	iter := ed.session.Query(stmt, values...).Iter()

	var dto api.EventDTO
	var status int32
	var guest_id int64
	var guest_name string
	var guest_response int32
	var guest_status int32

	eventList := make(map[int64]*api.EventDTO)

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&dto.Id, &dto.AuthorId, &dto.AuthorName, &dto.Description,
		&dto.PictureDigest, &dto.CreatedDate, &dto.InboxPosition, &dto.StartDate,
		&dto.EndDate, &dto.NumAttendees, &dto.NumGuests, &status,
		&guest_id, &guest_name, &guest_response, &guest_status) {

		event, ok := eventList[dto.Id]
		if !ok {
			event = new(api.EventDTO)
			*event = dto
			if status == 3 {
				event.Cancelled = true
			}
			eventList[dto.Id] = event
		}

		if guest_id != 0 {
			participant := &api.ParticipantDTO{
				UserId:           guest_id,
				Name:             guest_name,
				Response:         api.AttendanceResponse(guest_response),
				InvitationStatus: api.InvitationStatus(guest_status),
			}

			event.Participants[participant.UserId] = participant
		}
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadEventAndParticipants Error:", err)
		return nil, convErr(err)
	}

	// Build result

	events = make([]*api.EventDTO, len(event_ids))

	for pos, eventId := range event_ids {
		if event, ok := eventList[eventId]; ok {
			events[pos] = event
		}
	}

	return events, nil
}

// Read one or more events and do NOT include participants
/*func (dao *EventDAO) LoadEvent(event_ids ...int64) (events []*core.Event, err error) {

	checkSession(dao.session)

	if event_ids == nil || len(event_ids) == 0 {
		return nil, ErrInvalidArg
	}

	events = make([]*core.Event, 0, len(event_ids))

	stmt := `SELECT DISTINCT event_id, author_id, author_name, message, picture_digest,
		created_date, inbox_position, start_date,	end_date, num_attendees, num_guests,
		event_state
		FROM event WHERE event_id IN (` + utils.GenParams(len(event_ids)) + `)`

	values := utils.GenValues(event_ids)
	iter := dao.session.Query(stmt, values...).Iter()

	var dto event_dto

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&dto.eventId, &dto.authorId, &dto.authorName, &dto.description,
		&dto.pictureDigest, &dto.createdDate, &dto.inboxPosition, &dto.startDate,
		&dto.endDate, &dto.numAttendees, &dto.numGuests, &dto.eventState) {

		events = append(events, dto.toEventBuilder().Build())
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadEvent Error:", err)
		return nil, err
	}

	return events, nil
}*/

// FIXME: Each event of event table is in its own partition, classify events by date
// or something in order to improve read performance.
func (dao *EventDAO) LoadRecentEventsFromUser(user_id int64,
	fromDate int64) ([]*api.EventDTO, error) {

	checkSession(dao.session)

	// Get index of events
	//toDate := core.TimeToMillis(time.Now().Add(core.MAX_DIF_IN_START_DATE)) // One year
	userEvents, err := dao.loadUserInbox(user_id, fromDate)
	if err != nil {
		return nil, err
	}

	event_id_list := make([]int64, 0, len(userEvents))
	for _, userEvent := range userEvents {
		event_id_list = append(event_id_list, userEvent.EventId)
	}

	// Read from event table to get the actual info
	events, err := dao.LoadEvents(event_id_list...)
	if err != nil {
		log.Println("LoadUserEventsAndParticipants 2 (", user_id, "):", err)
		return nil, err
	}

	return events, nil
}

// Load user events in window delimited by (fromDate, toDate) where fromDate < toDate
func (dao *EventDAO) LoadEventsHistoryFromUser(user_id int64, fromDate int64,
	toDate int64) ([]*api.EventDTO, error) {

	checkSession(dao.session)

	var userEvents []*userEvent
	var err error

	if fromDate < toDate {
		userEvents, err = dao.loadUserInboxBetween(user_id, fromDate, toDate)
	} else {
		userEvents, err = dao.loadUserInboxReverse(user_id, fromDate)
	}

	if err != nil {
		log.Println("LoadUserEventsHistoryAndparticipants 1 (", user_id, "):", err)
		return nil, err
	}

	event_id_list := make([]int64, 0, len(userEvents))
	for _, userEvent := range userEvents {
		event_id_list = append(event_id_list, userEvent.EventId)
	}

	// Read from event table to get the actual info
	events, err := dao.LoadEvents(event_id_list...)

	if err != nil {
		log.Println("LoadUserEventsHistoryAndparticipants 2 (", user_id, "):", err)
		return nil, err
	}

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

func (de *EventDAO) LoadEventPicture(event_id int64) (*api.PictureDTO, error) {

	checkSession(de.session)

	stmt := `SELECT picture, picture_digest FROM event WHERE event_id = ?`
	q := de.session.Query(stmt, event_id)

	pic := new(api.PictureDTO)

	err := q.Scan(&pic.RawData, &pic.Digest)
	if err != nil {
		return nil, convErr(err)
	}

	return pic, nil
}

func (dao *EventDAO) SetEventPicture(event_id int64, picture *api.PictureDTO) error {

	checkSession(dao.session)

	stmt := `UPDATE event SET picture = ?, picture_digest = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, picture.RawData, picture.Digest, event_id)

	return convErr(q.Exec())
}

func (dao *EventDAO) SetEventStateAndInboxPosition(event_id int64,
	new_status api.EventState, new_position int64) error {

	checkSession(dao.session)

	stmt := `UPDATE event SET inbox_position = ?, event_state = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, new_position, new_status, event_id)

	return convErr(q.Exec())
}

// Compare-and-set (read-before) update operation
func (dao *EventDAO) SetNumGuests(eventId int64, numGuests int) (ok bool, err error) {

	checkSession(dao.session)

	read_stmt := `SELECT num_guests FROM event WHERE event_id = ?`
	q := dao.session.Query(read_stmt, eventId)

	var oldNumGuests int32
	var write_stmt string

	if err := q.Scan(&oldNumGuests); err != nil {
		return false, err
	}

	write_stmt = `UPDATE event SET num_guests = ? WHERE event_id = ?
								IF num_guests = ?`

	q = dao.session.Query(write_stmt, numGuests, eventId, oldNumGuests)
	return q.ScanCAS(nil)
}

/*func (dao *EventDAO) SetNumGuests(event_id int64, num_guests int32) error {

	checkSession(dao)

	stmt := `UPDATE event SET num_guests = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_guests, event_id)
	return q.Exec()
}*/

// Compare-and-set (read-before) update operation
/*func (dao *EventDAO) CompareAndSetNumAttendees(event_id int64, num_attendees int) (ok bool, err error) {

	checkSession(dao.session)

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
}*/

/*func (dao *EventDAO) SetNumAttendees(event_id int64, num_attendees int) error {

	checkSession(dao)

	stmt := `UPDATE event SET num_attendees = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_attendees, event_id)
	return q.Exec()
}*/

func (dao *EventDAO) SetParticipantInvitationStatus(user_id int64, event_id int64,
	status api.InvitationStatus) error {

	checkSession(dao.session)

	stmt := `UPDATE event SET guest_status = ? WHERE event_id = ? AND guest_id = ?`
	q := dao.session.Query(stmt, status, event_id, user_id)

	return convErr(q.Exec())
}

func (dao *EventDAO) SetParticipantResponse(participant int64, response api.AttendanceResponse,
	event *api.EventDTO) error {

	checkSession(dao.session)

	stmt_event := `UPDATE event SET guest_response = ? WHERE event_id = ? AND guest_id = ?`
	stmt_events_by_user := `UPDATE events_by_user SET response = ?
		WHERE user_id = ? AND event_bucket = ? AND start_date = ? AND event_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(stmt_event, response, event.Id, participant)
	batch.Query(stmt_events_by_user, response, participant, 1, event.StartDate, event.Id)

	return convErr(dao.session.ExecuteBatch(batch))
}
