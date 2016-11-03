package model

import (
	"peeple/areyouin/idgen"
	"peeple/areyouin/utils"
	"time"
)

type EventBuilder interface {
	SetAuthor(author *UserAccount)
	SetCreatedDate(date time.Time)
	SetStartDate(date time.Time)
	SetEndDate(date time.Time)
	SetDescription(desc string)
	ParticipantAdder() ParticipantAdder
	Build() (*Event, error)
}

type eventBuilder struct {
	ownerID int64
	// Event data
	eventID            int64
	authorID           int64
	authorName         string
	createdDate        time.Time
	startDate          time.Time
	endDate            time.Time
	description        string
	participantBuilder *participantListCreator
	eventManager       *EventManager
	//pictureDigest []byte
}

func (m *EventManager) NewEventBuilder(ownerID int64) EventBuilder {
	return &eventBuilder{
		ownerID:            ownerID,
		createdDate:        utils.GetCurrentTimeUTC(),
		participantBuilder: m.newParticipantListCreator(),
		eventManager:       m,
	}
}

func (b *eventBuilder) SetAuthor(author *UserAccount) {
	b.authorID = author.id
	b.authorName = author.name
	b.participantBuilder.SetAuthor(author.id)
}

func (b *eventBuilder) SetCreatedDate(date time.Time) {
	// Do not truncate to seconds still because it's used also
	// to compute timestamp in microseconds
	b.createdDate = date
}

func (b *eventBuilder) SetStartDate(date time.Time) {
	b.startDate = date.Truncate(time.Second)
}

func (b *eventBuilder) SetEndDate(date time.Time) {
	b.endDate = date.Truncate(time.Second)
}

func (b *eventBuilder) SetDescription(desc string) {
	b.description = desc
}

func (b *eventBuilder) ParticipantAdder() ParticipantAdder {
	return b.participantBuilder
}

func (b *eventBuilder) Build() (*Event, error) {

	// Event ID and timestamp
	b.eventID = idgen.NewID()
	b.participantBuilder.SetEventID(b.eventID)

	timestamp := b.createdDate.UnixNano() / 1000
	b.participantBuilder.SetTimestamp(timestamp)

	// Validate data
	if err := b.validateData(); err != nil {
		return nil, err
	}

	// Build participant list
	participants, err := b.participantBuilder.Build()
	if err != nil {
		return nil, err
	}

	// Build event
	event := &Event{
		id:            b.eventID,
		authorID:      b.authorID,
		authorName:    b.authorName,
		description:   b.description,
		createdDate:   b.createdDate.Truncate(time.Second),
		inboxPosition: b.startDate,
		startDate:     b.startDate,
		endDate:       b.endDate,
		numAttendees:  0,
		numGuests:     int32(len(participants)),
		participants:  participants,
		timestamp:     timestamp,
		owner:         b.ownerID,
		initialised:   true,
		isPersisted:   false,
		oldEvent:      nil,
	}

	return event, nil
}

func (b *eventBuilder) validateData() error {

	if b.eventID == 0 {
		return ErrInvalidEventData
	}

	if b.authorID == 0 || !IsValidName(b.authorName) {
		return ErrInvalidAuthor
	}

	if !IsValidDescription(b.description) {
		return ErrInvalidDescription
	}

	if !IsValidStartDate(b.startDate.Truncate(time.Second), b.createdDate) {
		return ErrInvalidStartDate
	}

	if !IsValidEndDate(b.endDate, b.startDate) {
		return ErrInvalidEndDate
	}

	if b.participantBuilder.Len() == 0 {
		return ErrParticipantsRequired
	}

	return nil
}
