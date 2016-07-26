package model

import (
	"peeple/areyouin/api"
	"peeple/areyouin/idgen"
	"peeple/areyouin/utils"
	"time"
)

const (
	EVENT_DESCRIPTION_MIN_LENGHT = 15
	EVENT_DESCRIPTION_MAX_LENGHT = 500
	MIN_DIF_IN_START_DATE        = 30 * time.Minute             // 30 minutes
	MAX_DIF_IN_START_DATE        = 365 * 24 * time.Hour         // 1 year
	MIN_DIF_IN_END_DATE          = 30*time.Minute - time.Second // 30 minutes (from start date)
	MAX_DIF_IN_END_DATE          = 7*24*time.Hour - time.Second // 1 week (from start date)
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

func NewEventFromDTO(dto *api.EventDTO) *Event {

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
	}

	for _, p := range dto.Participants {
		event.participants[p.UserId] = NewParticipantFromDTO(p)
	}

	return event
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

func (event *Event) cloneEmpty() *Event {
	eventCopy := new(Event)
	*eventCopy = *event
	eventCopy.participants = nil
	//eventCopy.NumGuests = 0
	//eventCopy.NumAttendees = 0
	return eventCopy
}

type ParticipantList struct {
	m map[int64]*Participant
}

func NewParticipantList() *ParticipantList {
	pl := &ParticipantList{
		m: make(map[int64]*Participant),
	}
	return pl
}

func (l *ParticipantList) add(p *Participant) {
	l.m[p.Id()] = p
}

func (l *ParticipantList) Get(id int64) (*Participant, bool) {
	v, ok := l.m[id]
	return v, ok
}

func (l *ParticipantList) Len() int {
	return len(l.m)
}

func (l *ParticipantList) Range(f func(k int64, v *Participant)) {
	for k, v := range l.m {
		f(k, v)
	}
}

func (l *ParticipantList) UserIds() []int64 {
	keys := make([]int64, 0, len(l.m))
	for k := range l.m {
		keys = append(keys, k)
	}
	return keys
}

/*type EventBuilder struct {
	e *Event
}

func NewEventBuilder() *EventBuilder {
	return &EventBuilder{
		e: &Event{},
	}
}

func (b *EventBuilder) SetId(id int64) *EventBuilder {
	b.e.ev.EventId = id
	return b
}

func (b *EventBuilder) SetAuthorId(authorId int64) *EventBuilder {
	b.e.ev.AuthorId = authorId
	return b
}

func (b *EventBuilder) SetAuthorName(name string) *EventBuilder {
	b.e.ev.AuthorName = name
	return b
}

func (b *EventBuilder) SetDescription(description string) *EventBuilder {
	b.e.ev.Message = description
	return b
}

func (b *EventBuilder) SetStartDate(startDate int64) *EventBuilder {
	b.e.ev.StartDate = startDate
	return b
}

func (b *EventBuilder) SetEndDate(endDate int64) *EventBuilder {
	b.e.ev.EndDate = endDate
	return b
}

func (b *EventBuilder) SetCreatedDate(createdDate int64) *EventBuilder {
	b.e.ev.CreatedDate = createdDate
	return b
}

func (b *EventBuilder) SetInboxPosition(timestamp int64) *EventBuilder {
	b.e.ev.InboxPosition = timestamp
	return b
}

func (b *EventBuilder) SetPictureDigest(pictureDigest []byte) *EventBuilder {
	b.e.ev.PictureDigest = pictureDigest
	return b
}

func (b *EventBuilder) SetPublic(isPublic bool) *EventBuilder {
	b.e.ev.IsPublic = isPublic
	return b
}

func (b *EventBuilder) SetNumAttendees(numAttendees int32) *EventBuilder {
	b.e.ev.NumAttendees = numAttendees
	return b
}

func (b *EventBuilder) SetNumGuests(numGuests int32) *EventBuilder {
	b.e.ev.NumGuests = numGuests
	return b
}

func (b *EventBuilder) SetCancelled(isCancelled bool) *EventBuilder {
	if isCancelled {
		b.e.ev.State = pb.EventState_CANCELLED
	}
	return b
}

func (b *EventBuilder) AddMember(participant *Participant) *EventBuilder {
	b.e.pl.m[participant.Id()] = participant
	return b
}

func (b *EventBuilder) Build() *Event {
	return b.e
}

type ParticipantListBuilder struct {
	pl *ParticipantList
}

func NewParticipantListBuilder() *ParticipantListBuilder {
	return &ParticipantListBuilder{
		pl: NewParticipantList(),
	}
}

func (b *ParticipantListBuilder) AddParticipant(p *Participant) *ParticipantListBuilder {
	b.pl.m[p.Id()] = p
	return b
}

func (b *ParticipantListBuilder) Build() *ParticipantList {
	return b.pl
}*/
