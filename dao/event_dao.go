package dao

import (
	core "areyouin/common"
	proto "areyouin/protocol"
	"github.com/gocql/gocql"
	"log"
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

func (dao *EventDAO) Insert(event *proto.Event) (ok bool, err error) {

	stmt := `INSERT INTO event (event_id, author_id, author_name, message, start_date,
		end_date, public, num_attendees, num_guests, created_date)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		IF NOT EXISTS`

	q := dao.session.Query(stmt,
		event.EventId, event.AuthorId, event.AuthorName, event.Message, event.StartDate,
		event.EndDate, event.IsPublic, event.NumAttendees, event.NumGuests, event.CreatedDate)

	return q.ScanCAS(nil)
}

// FIXME: There is no different between error and not found
func (dao *EventDAO) EventHasParticipant(event_id uint64, user_id uint64) bool {

	stmt := `SELECT event_id FROM event_participants
		WHERE event_id = ? AND user_id = ? LIMIT 1`

	exists := false

	if err := dao.session.Query(stmt, event_id, user_id).Scan(nil); err == nil {
		exists = true
	}

	return exists
}

func (dao *EventDAO) LoadParticipants(event_id uint64) []*proto.EventParticipant {

	stmt := `SELECT user_id, name, response, status FROM event_participants
		WHERE event_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, event_id, MAX_NUM_GUESTS).Iter()

	if iter == nil {
		log.Println("LoadParticipants iter is nil!!")
		return nil
	}

	participants := make([]*proto.EventParticipant, 0, 10)

	var user_id uint64
	var name string
	var response int32
	var status int32

	for iter.Scan(&user_id, &name, &response, &status) {
		participants = append(participants, &proto.EventParticipant{
			UserId:    user_id,
			Name:      name,
			Response:  proto.AttendanceResponse(response),
			Delivered: proto.MessageStatus(status),
		})
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadParticipants (", event_id, "):", err)
	}

	return participants
}

func (dao *EventDAO) AddOrUpdateParticipant(event_id uint64, participant *proto.EventParticipant) error {

	stmt := `INSERT INTO event_participants (event_id, user_id, name, response, status)
		VALUES (?, ?, ?, ?, ?)`

	q := dao.session.Query(stmt, event_id, participant.UserId, participant.Name,
		participant.Response, participant.Delivered)

	return q.Exec()
}

func (dao *EventDAO) AddOrUpdateParticipants(event_id uint64, participantList []*proto.EventParticipant) error {

	stmt := `INSERT INTO event_participants (event_id, user_id, name, response, status)
		VALUES (?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	for _, participant := range participantList {
		batch.Query(stmt, event_id, participant.UserId, participant.Name,
			participant.Response, participant.Delivered)
	}

	return dao.session.ExecuteBatch(batch)
}

func (dao *EventDAO) AddEventToUserInbox(user_id uint64, event *proto.Event, response proto.AttendanceResponse) error {

	stmt := `INSERT INTO user_events (user_id, event_id, end_date, response)
		VALUES (?, ?, ?, ?)`

	q := dao.session.Query(stmt, user_id, event.EventId, event.EndDate, response)

	return q.Exec()
}

// FIXME: Each event of event table is in its own partition, classify events by date
// or something in order to improve read performance
func (dao *EventDAO) LoadUserEvents(user_id uint64, fromDate int64) (events []*proto.Event, err error) {

	// First read events from user_events to get the IDs
	stmt := `SELECT event_id FROM user_events
		WHERE user_id = ? AND end_date >= ?`
	iter := dao.session.Query(stmt, user_id, fromDate).Iter()
	event_id_list := make([]interface{}, 0, 20)
	var event_id uint64

	for iter.Scan(&event_id) {
		event_id_list = append(event_id_list, event_id)
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadUserEvents 1 (", user_id, "):", err)
		return nil, err
	}

	// Read from event table to get the actual info
	events = make([]*proto.Event, 0, len(event_id_list))

	if len(event_id_list) > 0 {

		stmt = `SELECT event_id, author_id, author_name, message, start_date, end_date, num_attendees, num_guests, created_date
						FROM event WHERE event_id IN (` + GenParams(len(event_id_list)) + `)`

		iter = dao.session.Query(stmt, event_id_list...).Iter()

		var author_id uint64
		var author_name string
		var message string
		var start_date int64
		var end_date int64
		var num_attendees int32
		var num_guests int32
		var created_date int64

		for iter.Scan(&event_id, &author_id, &author_name, &message, &start_date, &end_date,
			&num_attendees, &num_guests, &created_date) {

			events = append(events, &proto.Event{
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
			})
		}

		if err := iter.Close(); err != nil {
			log.Println("LoadUserEvents 2 (", user_id, "):", err)
			return nil, err
		}
	}

	return events, nil
}

func (dao *EventDAO) SetNumGuests(event_id uint64, num_guests int32) error {
	stmt := `UPDATE event SET num_guests = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_guests, event_id)
	return q.Exec()
}

func (dao *EventDAO) SetNumAttendees(event_id uint64, num_attendees int) error {
	stmt := `UPDATE event SET num_attendees = ? WHERE event_id = ?`
	q := dao.session.Query(stmt, num_attendees, event_id)
	return q.Exec()
}

func (dao *EventDAO) SetParticipantStatus(user_id uint64, event_id uint64, status proto.MessageStatus) error {
	stmt := `UPDATE event_participants SET status = ? WHERE event_id = ? AND user_id = ?`
	q := dao.session.Query(stmt, status, event_id, user_id)
	return q.Exec()
}

func (dao *EventDAO) SetParticipantResponse(user_id uint64, event_id uint64, response proto.AttendanceResponse) error {
	stmt := `UPDATE event_participants SET response = ? WHERE event_id = ? AND user_id = ?`
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
