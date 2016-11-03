package model

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"strings"
	"time"
)

type Event struct {
	id            int64
	authorID      int64
	authorName    string
	description   string
	pictureDigest []byte
	createdDate   time.Time // Seconds precision
	modifiedDate  time.Time // Seconds precision
	inboxPosition time.Time // Seconds precision
	startDate     time.Time // Seconds precision
	endDate       time.Time // Seconds precision
	numAttendees  int32
	numGuests     int32
	cancelled     bool
	participants  map[int64]*Participant

	// Owner of this event object in RAM
	owner int64

	// Used to compute the timestamp for this event version when stored in DB
	timestamp int64

	// If this event is a modification of another one, oldEvent must
	// point to that object
	oldEvent *Event

	// Indicate if this object has a copy in database. For instance,
	// an event loaded from db will have isPersisted set. However, a
	// modified event will have it unset.
	isPersisted bool

	// Indicate if this object has been created and initialised
	// from this package. This is used to filter Event{} objects
	// in API calls
	initialised bool
}

func newEventFromDTO(dto *api.EventDTO) *Event {

	event := &Event{
		id:            dto.Id,
		authorID:      dto.AuthorId,
		authorName:    dto.AuthorName,
		description:   dto.Description,
		pictureDigest: dto.PictureDigest,
		createdDate:   utils.MillisToTimeUTC(dto.CreatedDate).Truncate(time.Second),
		inboxPosition: utils.MillisToTimeUTC(dto.InboxPosition).Truncate(time.Second),
		startDate:     utils.MillisToTimeUTC(dto.StartDate).Truncate(time.Second),
		endDate:       utils.MillisToTimeUTC(dto.EndDate).Truncate(time.Second),
		cancelled:     dto.Cancelled,
		participants:  make(map[int64]*Participant),
		timestamp:     dto.Timestamp,
		initialised:   true,
	}

	for _, p := range dto.Participants {
		event.participants[p.UserID] = newParticipantFromDTO(p)
		if p.Response == api.AttendanceResponse_ASSIST {
			event.numAttendees++
		}
	}

	event.numGuests = int32(len(event.participants))

	return event
}

func newEventListFromDTO(dtos []*api.EventDTO) []*Event {
	results := make([]*Event, 0, len(dtos))
	for _, eventDTO := range dtos {
		results = append(results, newEventFromDTO(eventDTO))
	}
	return results
}

func (e *Event) Id() int64 {
	return e.id
}

func (e *Event) AuthorID() int64 {
	return e.authorID
}

func (e *Event) AuthorName() string {
	return e.authorName
}

func (e *Event) CreatedDate() time.Time {
	return e.createdDate
}

func (e *Event) ModifiedDate() time.Time {
	return e.modifiedDate
}

func (e *Event) StartDate() time.Time {
	return e.startDate
}

func (e *Event) EndDate() time.Time {
	return e.endDate
}

func (e *Event) Title() string {

	var str string

	pos := strings.Index(e.description, "\n")
	if pos != -1 {
		str = e.description[0:pos]
	} else {
		str = e.description
	}

	fields := strings.Fields(str)
	title := fields[0]

	i := 1
	for i < utils.MinInt(10, len(fields)) {
		title += " " + fields[i]
		i++
	}

	if i < len(fields) {
		title += "..."
	}

	return title
}

func (e *Event) Description() string {
	return e.description
}

func (e *Event) InboxPosition() time.Time {
	return e.inboxPosition
}

func (e *Event) PictureDigest() []byte {
	return e.pictureDigest
}

func (e *Event) NumAttendees() int {
	return int(e.numAttendees)
}

func (e *Event) NumGuests() int {
	return int(e.numGuests)
}

func (e *Event) Status() api.EventState {

	currentDate := time.Now()

	if e.IsCancelled() {
		return api.EventState_CANCELLED
	} else if e.startDate.After(currentDate) {
		return api.EventState_NOT_STARTED
	} else if e.endDate.Before(currentDate) || e.endDate.Equal(currentDate) {
		return api.EventState_FINISHED
	}

	return api.EventState_ONGOING
}

func (e *Event) IsCancelled() bool {
	return e.cancelled
}

func (e *Event) Timestamp() int64 {
	return e.timestamp
}

func (e *Event) AsDTO() *api.EventDTO {

	dto := &api.EventDTO{
		Id:            e.id,
		AuthorId:      e.authorID,
		AuthorName:    e.authorName,
		Description:   e.description,
		PictureDigest: e.pictureDigest,
		CreatedDate:   utils.TimeToMillis(e.createdDate),
		InboxPosition: utils.TimeToMillis(e.inboxPosition),
		StartDate:     utils.TimeToMillis(e.startDate),
		EndDate:       utils.TimeToMillis(e.endDate),
		Cancelled:     e.cancelled,
		Participants:  make(map[int64]*api.ParticipantDTO),
		Timestamp:     e.timestamp,
	}

	for _, v := range e.participants {
		dto.Participants[v.id] = v.AsDTO()
	}

	return dto
}

func (e *Event) GetParticipant(id int64) *Participant {
	v, _ := e.participants[id]
	return v
}

func (e *Event) ParticipantIds() []int64 {
	keys := make([]int64, 0, len(e.participants))
	for k := range e.participants {
		keys = append(keys, k)
	}
	return keys
}

func (e *Event) Participants() []*Participant {
	values := make([]*Participant, 0, len(e.participants))
	for _, v := range e.participants {
		values = append(values, v)
	}
	return values
}

func (e *Event) Clone() *Event {
	eventCopy := new(Event)
	*eventCopy = *e
	eventCopy.pictureDigest = make([]byte, len(e.pictureDigest))
	copy(eventCopy.pictureDigest, e.pictureDigest)
	eventCopy.participants = make(map[int64]*Participant)
	for k, v := range e.participants {
		eventCopy.participants[k] = v
	}
	return eventCopy
}

func (e *Event) CloneEmptyParticipants() *Event {
	eventCopy := new(Event)
	*eventCopy = *e
	eventCopy.pictureDigest = make([]byte, len(e.pictureDigest))
	copy(eventCopy.pictureDigest, e.pictureDigest)
	eventCopy.participants = nil
	return eventCopy
}
