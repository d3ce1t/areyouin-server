package model

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"time"
)

type EventModifier interface {
	SetModifiedDate(date time.Time)
	SetStartDate(date time.Time)
	SetEndDate(date time.Time)
	SetDescription(desc string)
	ParticipantAdder() ParticipantAdder
	SetCancelled(cancelled bool)
	Build() (*Event, error)
}

type eventModifier struct {
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
	// New fields for modification
	modifiedDate        time.Time
	cancelled           bool
	currentParticipants map[int64]*Participant
	startDateChanged    bool
	endDateChanged      bool
	sourceEvent         *Event
}

func (m *EventManager) NewEventModifier(event *Event, ownerID int64) (EventModifier, error) {

	if event == nil {
		return nil, ErrIllegalArgument
	}

	if !event.initialised {
		return nil, ErrNotInitialised
	}

	if event.Status() != api.EventState_NOT_STARTED {
		return nil, ErrEventNotWritable
	}

	b := &eventModifier{
		// Event data
		eventID:            event.id,
		authorID:           event.authorID,
		authorName:         event.authorName,
		createdDate:        event.createdDate,
		startDate:          event.startDate,
		endDate:            event.endDate,
		description:        event.description,
		participantBuilder: m.newParticipantListCreator(),
		eventManager:       m,
		// New fields
		modifiedDate:        utils.GetCurrentTimeUTC(),
		cancelled:           event.cancelled,
		currentParticipants: make(map[int64]*Participant),
		sourceEvent:         event, // event is immutable
		ownerID:             ownerID,
	}

	b.participantBuilder.SetEventID(event.id)
	b.participantBuilder.SetAuthor(ownerID)

	for k, v := range event.participants {
		// Participant is immutable so I can assign the pointer
		b.currentParticipants[k] = v
	}

	return b, nil
}

func (b *eventModifier) SetModifiedDate(date time.Time) {
	// Do not truncate to seconds still because it's used also
	// to compute timestamp in microseconds
	b.modifiedDate = date
}

func (b *eventModifier) SetStartDate(date time.Time) {
	b.startDate = date.Truncate(time.Second)
	b.startDateChanged = true
}

func (b *eventModifier) SetEndDate(date time.Time) {
	b.endDate = date.Truncate(time.Second)
	b.endDateChanged = true
}

func (b *eventModifier) SetDescription(desc string) {
	b.description = desc
}

func (b *eventModifier) SetCancelled(cancelled bool) {
	b.cancelled = cancelled
}

func (b *eventModifier) ParticipantAdder() ParticipantAdder {
	return b.participantBuilder
}

func (b *eventModifier) Build() (*Event, error) {

	if err := b.validateData(); err != nil {
		return nil, err
	}

	timestamp := b.modifiedDate.UnixNano() / 1000
	b.participantBuilder.SetTimestamp(timestamp)

	event := &Event{
		id:            b.eventID,
		authorID:      b.authorID,
		authorName:    b.authorName,
		description:   b.description,
		createdDate:   b.createdDate,
		modifiedDate:  b.modifiedDate.Truncate(time.Second),
		inboxPosition: b.startDate,
		startDate:     b.startDate,
		endDate:       b.endDate,
		cancelled:     b.cancelled,
		numAttendees:  0,
		numGuests:     int32(len(b.currentParticipants)),
		participants:  make(map[int64]*Participant),
		owner:         b.ownerID,
		timestamp:     timestamp,
		initialised:   true,
		isPersisted:   false,
		oldEvent:      b.sourceEvent,
	}

	// Event is cancelled
	if b.cancelled {
		event.inboxPosition = b.modifiedDate.Truncate(time.Second)
	}

	// Copy current participants
	for k, v := range b.currentParticipants {
		// Participant is immutable so I can assign the pointer
		event.participants[k] = v
		if v.response == api.AttendanceResponse_ASSIST {
			event.numAttendees++
		}
	}

	// New participants added
	if b.participantBuilder.Len() > 0 {

		newParticipants, err := b.participantBuilder.Build()
		if err != nil {
			return nil, err
		}

		for k, v := range newParticipants {
			// Participant is immutable so I can assign the pointer
			event.participants[k] = v
		}

		event.numGuests += int32(len(newParticipants))
	}

	return event, nil
}

func (b *eventModifier) validateData() error {

	if b.eventID == 0 {
		return ErrInvalidEventData
	}

	if b.authorID == 0 || !IsValidName(b.authorName) {
		return ErrInvalidAuthor
	}

	if !IsValidDescription(b.description) {
		return ErrInvalidDescription
	}

	if b.startDateChanged && !IsValidStartDate(b.startDate, b.modifiedDate.Truncate(time.Second)) {
		return ErrInvalidStartDate
	}

	if b.endDateChanged && !IsValidEndDate(b.endDate, b.startDate) {
		return ErrInvalidEndDate
	}

	totalParticipants := b.participantBuilder.Len() + len(b.currentParticipants)
	if totalParticipants == 0 {
		return ErrParticipantsRequired
	}

	return nil
}
