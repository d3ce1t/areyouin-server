package model

import (
	"peeple/areyouin/api"
	"peeple/areyouin/idgen"
	"peeple/areyouin/utils"
	"strings"
	"time"
)

const (
	EVENT_DESCRIPTION_MIN_LENGHT = 15
	EVENT_DESCRIPTION_MAX_LENGHT = 500
	MIN_DIF_IN_START_DATE        = 30 * time.Minute     // 30 minutes
	MAX_DIF_IN_START_DATE        = 365 * 24 * time.Hour // 1 year
	MIN_DIF_IN_END_DATE          = 30 * time.Minute     // 30 minutes (from start date)
	MAX_DIF_IN_END_DATE          = 7 * 24 * time.Hour   // 1 week (from start date)
	EVENT_PICTURE_MAX_WIDTH      = 1280
	EVENT_PICTURE_MAX_HEIGHT     = 720
)

type Event struct {
	id            int64
	authorId      int64
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

func NewEvent(authorId int64, authorName string, createdDate int64, startDate int64,
	endDate int64, message string) *Event {
	event := &Event{
		id:            idgen.NewID(),
		authorId:      authorId,
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
		authorId:      dto.AuthorId,
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

func (e *Event) AuthorId() int64 {
	return e.authorId
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
	for i < utils.MinInt(5, len(fields)) {
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
	startDate := utils.UnixMillisToTime(e.startDate)
	endDate := utils.UnixMillisToTime(e.endDate)

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

func (e *Event) IsValid() (bool, error) {

	if e.id == 0 || e.authorId == 0 ||
		len(e.authorName) < USER_NAME_MIN_LENGTH || len(e.authorName) > USER_NAME_MAX_LENGTH ||
		e.description == "" || len(e.description) < EVENT_DESCRIPTION_MIN_LENGHT ||
		len(e.description) > EVENT_DESCRIPTION_MAX_LENGHT || e.numAttendees < 0 ||
		e.numGuests < 0 || e.numAttendees > e.numGuests {
		return false, ErrInvalidEventData
	}

	if !e.isValidStartDate() {
		return false, ErrInvalidStartDate
	}

	if !e.isValidEndDate() {
		return false, ErrInvalidEndDate
	}

	return true, nil
}

func (e *Event) isValidStartDate() bool {

	// I need only minute precision in order to emulate the same checking performed
	// by the client.
	createdDateMin := utils.UnixMillisToTime(e.createdDate - (e.createdDate % 60000))
	startDate := utils.UnixMillisToTime(e.startDate)

	if startDate.Before(createdDateMin.Add(MIN_DIF_IN_START_DATE)) ||
		startDate.After(createdDateMin.Add(MAX_DIF_IN_START_DATE)) {
		return false
	}

	return true
}

func (e *Event) isValidEndDate() bool {

	startDate := utils.UnixMillisToTime(e.startDate)
	endDate := utils.UnixMillisToTime(e.endDate)

	if endDate.Before(startDate.Add(MIN_DIF_IN_END_DATE)) ||
		endDate.After(startDate.Add(MAX_DIF_IN_END_DATE)) {
		return false
	}

	return true
}

func (e *Event) AsDTO() *api.EventDTO {

	dto := &api.EventDTO{
		Id:            e.id,
		AuthorId:      e.authorId,
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

/*func (e *Event) CloneFull() *Event {
	eventCopy := new(Event)
  *eventCopy = *e
  eventCopy.Participants = make(map[int64]*Participant)
  for k, v := range e.Participants {
		eventCopy.Participants[k] = v
  }
	return eventCopy
}*/

func (event *Event) CloneEmptyParticipants() *Event {
	eventCopy := new(Event)
	*eventCopy = *event
	eventCopy.participants = nil
	//eventCopy.NumGuests = 0
	//eventCopy.NumAttendees = 0
	return eventCopy
}
