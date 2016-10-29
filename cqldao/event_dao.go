package cqldao

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"time"

	"github.com/gocql/gocql"
)

const (
	MAX_NUM_GUESTS            = 100
	MAX_EVENTS_IN_RECENT_LIST = 100
)

type EventDAO struct {
	session *GocqlSession
}

func NewEventDAO(session api.DbSession) api.EventDAO {
	reconnectIfNeeded(session)
	return &EventDAO{session: session.(*GocqlSession)}
}

func (d *EventDAO) Insert(event *api.EventDTO) error {

	checkSession(d.session)

	stmt_event := `INSERT INTO event (event_id, author_id, author_name, message,
		start_date, end_date, num_attendees, num_guests, created_date,
		inbox_position, event_state)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var status int32
	if event.Cancelled {
		status = 3
	}

	var err error

	if len(event.Participants) > 0 {

		stmt_participant := `INSERT INTO event (event_id, guest_id, guest_name,
			guest_response, guest_status) VALUES (?, ?, ?, ?, ?)`

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

// Adds a participant to an existing event. If a participant with same id
// already exists it gets overwritten.
func (dao *EventDAO) AddParticipantToEvent(participant *api.ParticipantDTO, event *api.EventDTO) error {

	checkSession(dao.session)

	stmt := `INSERT INTO event (event_id, guest_id, guest_name,
		guest_response, guest_status) VALUES (?, ?, ?, ?, ?)`

	q := dao.session.Query(stmt, event.Id, participant.UserId, participant.Name,
		participant.Response, participant.InvitationStatus)

	return convErr(q.Exec())
}

func (d *EventDAO) RangeAll(f func(*api.EventDTO) error) error {

	checkSession(d.session)

	stmt := `SELECT event_id, author_id, author_name, message, picture_digest,
		created_date, inbox_position, start_date, end_date, num_attendees, num_guests,
		event_state, guest_id, guest_name, guest_response, guest_status
		FROM event`

	q := d.session.Query(stmt)
	return d.findAllAux(q, f)
}

func (d *EventDAO) RangeEvents(f func(*api.EventDTO) error, event_ids ...int64) error {

	checkSession(d.session)

	if event_ids == nil || len(event_ids) == 0 {
		return api.ErrInvalidArg
	}

	stmt := `SELECT event_id, author_id, author_name, message, picture_digest,
		created_date, inbox_position, start_date, end_date, num_attendees, num_guests,
		event_state, guest_id, guest_name, guest_response, guest_status
		FROM event WHERE event_id IN (` + utils.GenParams(len(event_ids)) + `)`

	values := utils.GenValues(event_ids)
	q := d.session.Query(stmt, values...)
	return d.findAllAux(q, f)
}

// Read one or more events and their participants. Event info and participants are in the
// same partition
func (d *EventDAO) LoadEvents(event_ids ...int64) ([]*api.EventDTO, error) {

	checkSession(d.session)

	if event_ids == nil || len(event_ids) == 0 {
		return nil, api.ErrInvalidArg
	}

	stmt := `SELECT event_id, author_id, author_name, message, picture_digest,
		created_date, inbox_position, start_date, end_date, num_attendees, num_guests,
		event_state, guest_id, guest_name, guest_response, guest_status
		FROM event WHERE event_id IN (` + utils.GenParams(len(event_ids)) + `)`

	values := utils.GenValues(event_ids)
	q := d.session.Query(stmt, values...)

	eventList := make(map[int64]*api.EventDTO)

	err := d.findAllAux(q, func(event *api.EventDTO) error {
		eventList[event.Id] = event
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Build result

	events := make([]*api.EventDTO, len(event_ids))

	for pos, eventID := range event_ids {
		if event, ok := eventList[eventID]; ok {
			events[pos] = event
		}
	}

	return events, nil
}

func (d *EventDAO) findAllAux(query *gocql.Query, f func(*api.EventDTO) error) error {

	checkSession(d.session)

	iter := query.Iter()

	var dto api.EventDTO
	var status int32
	var guest_id int64
	var guest_name string
	var guest_response int32
	var guest_status int32

	var err error
	var currentEvent *api.EventDTO

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&dto.Id, &dto.AuthorId, &dto.AuthorName, &dto.Description,
		&dto.PictureDigest, &dto.CreatedDate, &dto.InboxPosition, &dto.StartDate,
		&dto.EndDate, &dto.NumAttendees, &dto.NumGuests, &status,
		&guest_id, &guest_name, &guest_response, &guest_status) {

		if currentEvent == nil || currentEvent.Id != dto.Id {

			// Send currentEvent to f
			if currentEvent != nil {
				if err = f(currentEvent); err != nil {
					currentEvent = nil
					break
				}
			}

			// Read new event
			currentEvent = new(api.EventDTO)
			*currentEvent = dto
			currentEvent.Participants = make(map[int64]*api.ParticipantDTO)
			if status == 3 {
				currentEvent.Cancelled = true
			}
		}

		if guest_id != 0 {
			participant := &api.ParticipantDTO{
				UserId:           guest_id,
				Name:             guest_name,
				Response:         api.AttendanceResponse(guest_response),
				InvitationStatus: api.InvitationStatus(guest_status),
			}
			currentEvent.Participants[participant.UserId] = participant
		}
	}

	if err != nil {
		iter.Close()
	} else {
		err = iter.Close()
		if err == nil && currentEvent != nil {
			err = f(currentEvent)
		}
	}

	return convErr(err)
}

// FIXME: Each event of event table is in its own partition, classify events by date
// or something in order to improve read performance.
/*func (dao *EventDAO) LoadRecentEventsFromUser(user_id int64,
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

func (d *EventDAO) CancelEvent(eventID int64, oldPosition time.Time, newPosition time.Time, userIDs []int64) error {

	checkSession(d.session)

	updateStmt := `UPDATE event SET inbox_position = ?, event_state = ? WHERE event_id = ?`
	deleteStmt := `DELETE FROM events_timeline WHERE bucket = ? AND position = ? AND event_id = ?`
	insertStmt := `INSERT INTO events_timeline (bucket, event_id, position) VALUES (?, ?, ?)`
	historyStmt := `INSERT INTO events_history_by_user (user_id, position, event_id) VALUES (?, ?, ?)`

	batch := d.session.NewBatch(gocql.LoggedBatch)

	// Set event inboxPosition and mark as cancelled
	batch.Query(updateStmt, utils.TimeToMillis(newPosition), api.EventState_CANCELLED, eventID)

	// Change event timeline
	batch.Query(deleteStmt, oldPosition.Year(), utils.TimeToMillis(oldPosition), eventID)
	batch.Query(insertStmt, newPosition.Year(), eventID, utils.TimeToMillis(newPosition))

	for _, userID := range userIDs {
		// Insert event into user events history
		batch.Query(historyStmt, userID, utils.TimeToMillis(newPosition), eventID)
	}

	return convErr(d.session.ExecuteBatch(batch))
}

// Compare-and-set (read-before) update operation
/*func (dao *EventDAO) SetNumGuests(eventId int64, numGuests int) (ok bool, err error) {

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
}*/

func (dao *EventDAO) SetNumGuests(eventId int64, numGuests int) (bool, error) {

	checkSession(dao.session)

	stmt := `UPDATE event SET num_guests = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, numGuests, eventId)
	if err := q.Exec(); err != nil {
		return false, err
	}

	return true, nil
}

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

func (d *EventDAO) SetParticipantResponse(participant int64, response api.AttendanceResponse,
	event *api.EventDTO) error {

	checkSession(d.session)

	stmt := `UPDATE event SET guest_response = ? WHERE event_id = ? AND guest_id = ?`
	q := d.session.Query(stmt, response, event.Id, participant)

	return convErr(q.Exec())
}
