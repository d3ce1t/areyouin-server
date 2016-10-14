package model

import (
	"peeple/areyouin/api"
	"peeple/areyouin/idgen"
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
	createdDate   int64
	inboxPosition int64
	startDate     int64
	endDate       int64
	numAttendees  int32
	numGuests     int32
	cancelled     bool
	participants  map[int64]*Participant
}

func newEvent(authorID int64, authorName string, createdDate int64, startDate int64,
	endDate int64, message string) *Event {
	event := &Event{
		id:            idgen.NewID(),
		authorID:      authorID,
		authorName:    authorName,
		description:   message,
		createdDate:   createdDate,
		inboxPosition: startDate,
		startDate:     startDate,
		endDate:       endDate,
		numAttendees:  0,
		numGuests:     0,
		participants:  make(map[int64]*Participant),
	}
	return event
}

func newEventFromDTO(dto *api.EventDTO) *Event {

	event := &Event{
		id:            dto.Id,
		authorID:      dto.AuthorId,
		authorName:    dto.AuthorName,
		description:   dto.Description,
		pictureDigest: dto.PictureDigest,
		createdDate:   dto.CreatedDate,
		inboxPosition: dto.InboxPosition,
		startDate:     dto.StartDate,
		endDate:       dto.EndDate,
		numAttendees:  dto.NumAttendees,
		numGuests:     dto.NumGuests,
		cancelled:     dto.Cancelled,
		participants:  make(map[int64]*Participant),
	}

	for _, p := range dto.Participants {
		event.participants[p.UserId] = newParticipantFromDTO(p)
	}

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

func (e *Event) CreatedDate() int64 {
	return e.createdDate
}

func (e *Event) StartDate() int64 {
	return e.startDate
}

func (e *Event) EndDate() int64 {
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

func (e *Event) InboxPosition() int64 {
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
	startDate := utils.MillisToTimeUTC(e.startDate)
	endDate := utils.MillisToTimeUTC(e.endDate)

	if e.IsCancelled() {
		return api.EventState_CANCELLED
	} else if startDate.After(currentDate) {
		return api.EventState_NOT_STARTED
	} else if endDate.Before(currentDate) || endDate.Equal(currentDate) {
		return api.EventState_FINISHED
	} else {
		return api.EventState_ONGOING
	}
}

func (e *Event) IsCancelled() bool {
	return e.cancelled
}

func (e *Event) AsDTO() *api.EventDTO {

	dto := &api.EventDTO{
		Id:            e.id,
		AuthorId:      e.authorID,
		AuthorName:    e.authorName,
		Description:   e.description,
		PictureDigest: e.pictureDigest,
		CreatedDate:   e.createdDate,
		InboxPosition: e.inboxPosition,
		StartDate:     e.startDate,
		EndDate:       e.endDate,
		NumAttendees:  e.numAttendees,
		NumGuests:     e.numGuests,
		Cancelled:     e.cancelled,
		Participants:  make(map[int64]*api.ParticipantDTO),
	}

	for _, v := range e.participants {
		dto.Participants[v.id] = v.AsDTO()
	}

	return dto
}

/*func (e *Event) SetParticipants(participants map[int64]*Participant) {
	e.Participants = participants
	if participants != nil {
		e.NumGuests = int32(len(participants))
	}
}*/

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

func (e *Event) addParticipant(p *Participant) {
	e.participants[p.id] = p
	e.numGuests = int32(len(e.participants))
}

/*func (event *Event) GetEventWithoutParticipants() *Event {
	new_event := &Event{}
	*new_event = *event // copy
	new_event.SetParticipants(nil)
	return new_event
}*/

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
	//eventCopy.NumGuests = 0
	//eventCopy.NumAttendees = 0
	return eventCopy
}
