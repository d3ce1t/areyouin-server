package model

import (
	"peeple/areyouin/api"
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
	eventID            int64
	author             *UserAccount
	createdDate        time.Time
	startDate          time.Time
	endDate            time.Time
	description        string
	participantBuilder *participantListCreator
	eventManager       *EventManager
	//pictureDigest []byte
}

func (m *EventManager) newEventBuilder() EventBuilder {
	return &eventBuilder{
		createdDate:        utils.GetCurrentTimeUTC(),
		participantBuilder: m.newParticipantListCreator(),
		eventManager:       m,
	}
}

func (b *eventBuilder) SetAuthor(author *UserAccount) {
	if author != nil {
		b.author = author
		b.participantBuilder.SetOwner(author.id)
	}
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

	// Event ID
	b.eventID = idgen.NewID()
	b.participantBuilder.SetEventID(b.eventID)

	// Timestamp
	timestamp := b.createdDate.UnixNano() / 1000
	b.participantBuilder.SetTimestamp(timestamp)

	// Add author to the event
	if b.author != nil {
		pAuthor := NewParticipant(b.author.Id(), b.author.Name(),
			api.AttendanceResponse_ASSIST, api.InvitationStatus_SERVER_DELIVERED)
		b.ParticipantAdder().AddParticipant(pAuthor)
	}

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
		authorID:      b.author.id,
		authorName:    b.author.name,
		description:   b.description,
		createdDate:   b.createdDate.Truncate(time.Second),
		inboxPosition: b.startDate,
		startDate:     b.startDate,
		endDate:       b.endDate,
		Participants:  participants,
		modifiedDate:  b.createdDate.Truncate(time.Second),
		timestamp:     timestamp,
		owner:         b.author.id,
		isPersisted:   false,
		oldEvent:      nil,
	}

	return event, nil
}

func (b *eventBuilder) validateData() error {

	if b.eventID == 0 {
		return ErrInvalidEvent
	}

	if b.author == nil || b.author.IsZero() || !b.author.isPersisted ||
		b.author.id == 0 || !IsValidName(b.author.name) {
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

	// Build() always insert author as participant. So Len() will never return 0
	/*if b.participantBuilder.Len() == 0 {
		return ErrParticipantsRequired
	}*/

	return nil
}
