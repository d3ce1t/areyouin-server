package model

import (
	"bytes"
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"strings"
	"time"
)

type EventModifier interface {
	SetModifiedDate(date time.Time) EventModifier
	SetStartDate(date time.Time) EventModifier
	SetEndDate(date time.Time) EventModifier
	SetDescription(desc string) EventModifier
	ParticipantAdder() ParticipantAdder
	SetCancelled(cancelled bool) EventModifier
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
	pictureDigest      []byte
	// New fields for modification
	modifiedDate        time.Time
	cancelled           bool
	currentParticipants map[int64]*Participant
	startDateChanged    bool
	endDateChanged      bool
	sourceEvent         *Event
}

func (m *EventManager) NewEventModifier(event *Event, ownerID int64) EventModifier {

	b := &eventModifier{
		ownerID:             ownerID,
		modifiedDate:        utils.GetCurrentTimeUTC(),
		currentParticipants: make(map[int64]*Participant),
		participantBuilder:  m.newParticipantListCreator(),
		sourceEvent:         event, // event is immutable
		eventManager:        m,
	}

	if event != nil {
		b.eventID = event.id
		b.authorID = event.authorID
		b.authorName = event.authorName
		b.createdDate = event.createdDate
		b.startDate = event.startDate
		b.endDate = event.endDate
		b.description = event.description
		b.pictureDigest = bytes.Repeat(event.pictureDigest, 1)
		b.cancelled = event.cancelled

		for k, p := range event.Participants.participants {
			b.currentParticipants[k] = p.Clone()
		}

		b.participantBuilder.SetEventID(event.id)
	}

	b.participantBuilder.SetOwner(ownerID)

	return b
}

func (b *eventModifier) SetModifiedDate(date time.Time) EventModifier {
	// Do not truncate to seconds still because it's used also
	// to compute timestamp in microseconds
	b.modifiedDate = date
	return b
}

func (b *eventModifier) SetStartDate(date time.Time) EventModifier {
	b.startDate = date.Truncate(time.Second)
	b.startDateChanged = true
	return b
}

func (b *eventModifier) SetEndDate(date time.Time) EventModifier {
	b.endDate = date.Truncate(time.Second)
	b.endDateChanged = true
	return b
}

func (b *eventModifier) SetDescription(desc string) EventModifier {
	b.description = strings.TrimSpace(desc)
	return b
}

func (b *eventModifier) SetCancelled(cancelled bool) EventModifier {
	b.cancelled = cancelled
	return b
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
		inboxPosition: b.startDate,
		startDate:     b.startDate,
		endDate:       b.endDate,
		pictureDigest: bytes.Repeat(b.pictureDigest, 1),
		cancelled:     b.cancelled,
		Participants:  newParticipantList(),
		owner:         b.ownerID,
		modifiedDate:  b.modifiedDate.Truncate(time.Second),
		timestamp:     timestamp,
		isPersisted:   false,
		oldEvent:      b.sourceEvent,
	}

	// Event is cancelled
	if b.cancelled {
		event.inboxPosition = b.modifiedDate.Truncate(time.Second)
	}

	// Copy current participants
	for k, p := range b.currentParticipants {
		if b.sourceEvent != nil && !b.sourceEvent.isPersisted {
			p.nameTS = timestamp
			p.responseTS = timestamp
			p.statusTS = timestamp
		}
		// Participant is immutable so I can assign the pointer
		event.Participants.participants[k] = p
		if p.response == api.AttendanceResponse_ASSIST {
			event.Participants.numAttendees++
		}
	}

	// New participants added
	if b.participantBuilder.Len() > 0 {

		newParticipants, err := b.participantBuilder.Build()
		if err != nil {
			return nil, err
		}

		for k, v := range newParticipants.participants {
			// Participant is immutable so I can assign the pointer
			event.Participants.participants[k] = v
			if v.response == api.AttendanceResponse_ASSIST {
				event.Participants.numAttendees++
			}
		}
	}

	event.Participants.numGuests = len(event.Participants.participants)

	return event, nil
}

func (b *eventModifier) validateData() error {

	if b.eventID == 0 {
		return ErrInvalidEvent
	}

	if b.ownerID == 0 {
		return ErrInvalidOwner
	}

	if b.authorID == 0 || !IsValidName(b.authorName) {
		return ErrInvalidAuthor
	}

	if _, ok := b.currentParticipants[b.ownerID]; !ok {
		return ErrEventNotWritable
	}

	if b.ownerID != b.authorID {
		return ErrEventNotWritable
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
