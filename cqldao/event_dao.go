package cqldao

import (
	"fmt"
	"peeple/areyouin/api"
	"peeple/areyouin/utils"

	"github.com/gocql/gocql"
)

const (
	MAX_NUM_GUESTS            = 100 // Not used by now
	MAX_EVENTS_IN_RECENT_LIST = 100 // Not used by now

	queryCols = `event_id, author_id, author_name, message, picture_digest,
		created_date, inbox_position, start_date, end_date, event_state, event_timestamp,
		guest_id, guest_name, guest_response, guest_status, writetime(guest_name) as guest_name_ts, 
		writetime(guest_response) as guest_response_ts,	writetime(guest_status) as guest_status_ts`
)

type EventDAO struct {
	session *GocqlSession
}

func NewEventDAO(session api.DbSession) api.EventDAO {
	reconnectIfNeeded(session)
	return &EventDAO{session: session.(*GocqlSession)}
}

// Insert event into database with timestamp the one of the event. Participant will have the same timestamp.
func (d *EventDAO) Insert(event *api.EventDTO) error {

	checkSession(d.session)

	if event == nil || event.Id == 0 {
		return ErrIllegalArguments
	}

	stmtTimeline := `INSERT INTO events_timeline (bucket, event_id, position)
		VALUES (?, ?, ?) USING TIMESTAMP ?`

	stmtEvent := `INSERT INTO event (event_id, author_id, author_name, message,
		start_date, end_date, created_date, inbox_position, event_state, event_timestamp)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)  USING TIMESTAMP ?`

	var status int32
	if event.Cancelled {
		status = 3
	}

	// Prepare batch
	batch := d.session.NewBatch(gocql.LoggedBatch)
	//batch.WithTimestamp(event.Timestamp)

	timeLineBucket := utils.MillisToTimeUTC(event.EndDate).Year()
	batch.Query(stmtTimeline, timeLineBucket, event.Id, event.EndDate, event.Timestamp)

	batch.Query(stmtEvent, event.Id, event.AuthorId, event.AuthorName,
		event.Description, event.StartDate, event.EndDate, event.CreatedDate,
		event.InboxPosition, status, event.Timestamp, event.Timestamp)

	if len(event.Participants) > 0 {
		stmtParticipant := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
			VALUES (?, ?, ?, ?, ?) USING TIMESTAMP ?`
		for _, p := range event.Participants {
			batch.Query(stmtParticipant, event.Id, p.UserID, p.Name, p.Response, p.InvitationStatus, event.Timestamp)
		}
	}

	// Execute
	err := d.session.ExecuteBatch(batch)

	return convErr(err)
}

// Replace modifies information of an existing event. This implementation only
// changes message, start_date, end_date, inbox_position and event state. Moreover,
// it doesn't modify information related to existing participants but only can add
// new ones where version isn't set.
// NOTE: This implementation takes into account the single use case where an event update
// contains new participants only.
func (d *EventDAO) Replace(oldEvent *api.EventDTO, newEvent *api.EventDTO) error {

	// TODO: Optimise use case when only it has to write to event (with and without participants)

	if oldEvent == nil || newEvent == nil || oldEvent.Id != newEvent.Id || newEvent.Id == 0 {
		return ErrIllegalArguments
	}

	checkSession(d.session)

	// Get newPosition and oldPosition in timeline
	newPosition := utils.MillisToTimeUTC(newEvent.EndDate)
	if newEvent.Cancelled {
		newPosition = utils.MillisToTimeUTC(newEvent.InboxPosition)
	}

	oldPosition := utils.MillisToTimeUTC(oldEvent.EndDate)
	if oldEvent.Cancelled {
		oldPosition = utils.MillisToTimeUTC(oldEvent.InboxPosition)
	}

	// Prepare batch
	batch := d.session.NewBatch(gocql.LoggedBatch)

	// WORKAROUND: WithTimestamp isn't working so make use of USING TIMESTAMP as part of the Query.
	//batch.WithTimestamp(newEvent.Timestamp)

	// WORKAROUND: By now, do not move cancelled events in time line in order to
	// let a fresh reboot of the server to load cancelled events and give clients
	// a chance to receive the last event state.
	if !newEvent.Cancelled && newPosition != oldPosition {
		stmtTimelineDelete := `DELETE FROM events_timeline USING TIMESTAMP ? WHERE bucket = ? AND position = ? AND event_id = ?`
		stmtTimelineAdd := `INSERT INTO events_timeline (bucket, event_id, position) VALUES (?, ?, ?) USING TIMESTAMP ?`
		newBucket := newPosition.Year()
		oldBucket := oldPosition.Year()
		batch.Query(stmtTimelineDelete, newEvent.Timestamp, oldBucket, utils.TimeToMillis(oldPosition), oldEvent.Id)
		batch.Query(stmtTimelineAdd, newBucket, newEvent.Id, utils.TimeToMillis(newPosition), newEvent.Timestamp)
	}

	// TODO: Replace status by cancelled in DB
	var status int32
	if newEvent.Cancelled {
		status = 3
	}

	stmtEvent := `INSERT INTO event (event_id, message, start_date,	end_date,
		inbox_position, event_state, event_timestamp) VALUES (?, ?, ?, ?, ?, ?, ?) USING TIMESTAMP ?`
	batch.Query(stmtEvent, newEvent.Id, newEvent.Description, newEvent.StartDate, newEvent.EndDate,
		newEvent.InboxPosition, status, newEvent.Timestamp, newEvent.Timestamp)

	// Only add new participants when updating/replacing
	newParticipants := d.extractNewParticipants(newEvent, oldEvent)
	if len(newParticipants) > 0 {
		stmtParticipant := `INSERT INTO event (event_id, guest_id, guest_name, guest_response, guest_status)
		VALUES (?, ?, ?, ?, ?) USING TIMESTAMP ?`
		for _, p := range newParticipants {
			batch.Query(stmtParticipant, newEvent.Id, p.UserID, p.Name, p.Response, p.InvitationStatus, newEvent.Timestamp)
		}
	}

	if !oldEvent.Cancelled && newEvent.Cancelled {
		// Event was just cancelled. Insert event into user events history
		historyStmt := `INSERT INTO events_history_by_user (user_id, position, event_id) VALUES (?, ?, ?) USING TIMESTAMP ?`
		for pID := range newEvent.Participants {
			batch.Query(historyStmt, pID, utils.TimeToMillis(newPosition), newEvent.Id, newEvent.Timestamp)
		}
	}

	// Execute
	err := d.session.ExecuteBatch(batch)

	return convErr(err)
}

func (d *EventDAO) InsertParticipant(p *api.ParticipantDTO) error {

	checkSession(d.session)

	infoStmt := `INSERT INTO event (event_id, guest_id, guest_name) VALUES (?, ?, ?) USING TIMESTAMP ?`
	responseStmt := `INSERT INTO event (event_id, guest_id, guest_response) VALUES (?, ?, ?) USING TIMESTAMP ?`
	statusStmt := `INSERT INTO event (event_id, guest_id, guest_status) VALUES (?, ?, ?) USING TIMESTAMP ?`

	batch := d.session.NewBatch(gocql.UnloggedBatch)
	batch.Query(infoStmt, p.EventID, p.UserID, p.Name, p.NameTS)
	batch.Query(responseStmt, p.EventID, p.UserID, p.Response, p.ResponseTS)
	batch.Query(statusStmt, p.EventID, p.UserID, p.InvitationStatus, p.StatusTS)

	return convErr(d.session.ExecuteBatch(batch))
}

func (d *EventDAO) RangeAll(f func(*api.EventDTO) error) error {
	checkSession(d.session)
	stmt := fmt.Sprintf("SELECT %v FROM event", queryCols)
	q := d.session.Query(stmt)
	return d.findAllAux(q, f)
}

func (d *EventDAO) RangeEvents(f func(*api.EventDTO) error, event_ids ...int64) error {

	checkSession(d.session)

	if event_ids == nil || len(event_ids) == 0 {
		return api.ErrInvalidArg
	}

	stmt := fmt.Sprintf("SELECT %v FROM event WHERE event_id IN (%v)", queryCols, utils.GenParams(len(event_ids)))
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

	stmt := fmt.Sprintf("SELECT %v FROM event WHERE event_id IN (%v)", queryCols, utils.GenParams(len(event_ids)))
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

func (d *EventDAO) LoadParticipant(participantID int64, eventID int64) (*api.ParticipantDTO, error) {

	checkSession(d.session)

	if participantID == 0 || eventID == 0 {
		return nil, api.ErrInvalidArg
	}

	cols := `guest_id, guest_name, guest_response, guest_status, writetime(guest_name) as guest_name_ts,
		writetime(guest_response) as guest_response_ts,	writetime(guest_status) as guest_status_ts`

	stmt := fmt.Sprintf("SELECT %v FROM event WHERE event_id = ? AND guest_id = ?", cols)
	q := d.session.Query(stmt, eventID, participantID)

	var guestID int64
	var guestName string
	var guestResponse, guestStatus int32
	var guestNameTS, guestResponseTS, guestStatusTS int64

	err := q.Scan(&guestID, &guestName, &guestResponse, &guestStatus, &guestNameTS, &guestResponseTS, &guestStatusTS)
	if err != nil {
		return nil, convErr(err)
	}

	participant := &api.ParticipantDTO{
		UserID:           guestID,
		EventID:          eventID,
		Name:             guestName,
		Response:         api.AttendanceResponse(guestResponse),
		InvitationStatus: api.InvitationStatus(guestStatus),
		NameTS:           guestNameTS,
		ResponseTS:       guestResponseTS,
		StatusTS:         guestStatusTS,
	}

	return participant, nil
}

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

func (d *EventDAO) DeleteAll() error {

	checkSession(d.session)

	if err := d.session.Query("TRUNCATE event").Exec(); err != nil {
		return err
	}

	if err := d.session.Query("TRUNCATE events_timeline").Exec(); err != nil {
		return err
	}

	if err := d.session.Query("TRUNCATE events_history_by_user").Exec(); err != nil {
		return err
	}

	return nil
}

func (d *EventDAO) findAllAux(query *gocql.Query, handler func(*api.EventDTO) error) error {

	checkSession(d.session)

	iter := query.Iter()

	var dto api.EventDTO
	var status int32
	var guestID int64
	var guestName string
	var guestResponse, guestStatus int32
	var guestNameTS, guestResponseTS, guestStatusTS int64

	var err error
	var currentEvent *api.EventDTO

	// Except guest attributes, all of the attributes are STATIC in cassandra
	for iter.Scan(&dto.Id, &dto.AuthorId, &dto.AuthorName, &dto.Description, &dto.PictureDigest,
		&dto.CreatedDate, &dto.InboxPosition, &dto.StartDate, &dto.EndDate, &status, &dto.Timestamp,
		&guestID, &guestName, &guestResponse, &guestStatus, &guestNameTS, &guestResponseTS, &guestStatusTS) {

		if currentEvent == nil || currentEvent.Id != dto.Id {

			// Send currentEvent to handler
			if currentEvent != nil {
				if err = handler(currentEvent); err != nil {
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

		if guestID != 0 {
			participant := &api.ParticipantDTO{
				UserID:           guestID,
				EventID:          dto.Id,
				Name:             guestName,
				Response:         api.AttendanceResponse(guestResponse),
				InvitationStatus: api.InvitationStatus(guestStatus),
				NameTS:           guestNameTS,
				ResponseTS:       guestResponseTS,
				StatusTS:         guestStatusTS,
			}
			currentEvent.Participants[participant.UserID] = participant
		}
	}

	if err != nil {
		iter.Close()
	} else {
		err = iter.Close()
		if err == nil && currentEvent != nil {
			// Send last loaded event
			err = handler(currentEvent)
		}
	}

	return convErr(err)
}

// ExtractNewParticipants extracts participants from extractList that are not in baseList
func (d *EventDAO) extractNewParticipants(extractEvent *api.EventDTO, baseEvent *api.EventDTO) map[int64]*api.ParticipantDTO {

	newParticipants := make(map[int64]*api.ParticipantDTO)

	for pID, participant := range extractEvent.Participants {
		if _, ok := baseEvent.Participants[pID]; !ok {
			newParticipants[pID] = participant
		}
	}

	return newParticipants
}
