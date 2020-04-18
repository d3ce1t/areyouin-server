package model

import (
	"time"

	"github.com/d3ce1t/areyouin-server/api"
)

type ParticipantModifier interface {
	SetResponse(resp api.AttendanceResponse) ParticipantModifier
	SetInvitationStatus(status api.InvitationStatus) ParticipantModifier
	Build() (*Participant, error)
}

type participantModifier struct {
	eventID    int64
	pID        int64
	name       string
	response   api.AttendanceResponse
	status     api.InvitationStatus
	nameTS     int64
	responseTS int64
	statusTS   int64
}

func (m *EventManager) NewParticipantModifier(p *Participant) ParticipantModifier {

	if p == nil {
		return &participantModifier{}
	}

	b := &participantModifier{
		eventID:    p.eventID,
		pID:        p.id,
		name:       p.name,
		response:   p.response,
		status:     p.invitationStatus,
		nameTS:     p.nameTS,
		responseTS: p.responseTS,
		statusTS:   p.statusTS,
	}

	return b
}

func (b *participantModifier) SetResponse(resp api.AttendanceResponse) ParticipantModifier {
	b.response = resp
	b.responseTS = time.Now().UnixNano() / 1000
	return b
}

func (b *participantModifier) SetInvitationStatus(status api.InvitationStatus) ParticipantModifier {
	b.status = status
	b.statusTS = time.Now().UnixNano() / 1000
	return b
}

func (b *participantModifier) Build() (*Participant, error) {

	// Validate data
	if err := b.validateData(); err != nil {
		return nil, err
	}

	// Build participant
	participant := &Participant{
		id:               b.pID,
		eventID:          b.eventID,
		name:             b.name,
		response:         b.response,
		invitationStatus: b.status,
		nameTS:           b.nameTS,
		responseTS:       b.responseTS,
		statusTS:         b.statusTS,
	}

	return participant, nil
}

func (b *participantModifier) validateData() error {

	if b.eventID == 0 {
		return ErrInvalidEvent
	}

	if b.pID == 0 {
		return ErrInvalidParticipant
	}

	if !IsValidName(b.name) {
		return ErrInvalidName
	}

	return nil
}
