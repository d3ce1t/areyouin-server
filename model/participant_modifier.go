package model

import (
	"peeple/areyouin/api"
	"time"
)

type ParticipantModifier interface {
	SetResponse(resp api.AttendanceResponse)
	SetInvitationStatus(status api.InvitationStatus)
	Build() (*Participant, error)
}

type participantModifier struct {
	eventID    int64
	pID        int64
	name       string
	response   api.AttendanceResponse
	status     api.InvitationStatus
	responseTS int64
	statusTS   int64
}

func (m *EventManager) NewParticipantModifier(p *Participant) (ParticipantModifier, error) {

	if p == nil {
		return nil, ErrIllegalArgument
	}

	b := &participantModifier{
		eventID:    p.eventID,
		pID:        p.id,
		name:       p.name,
		response:   p.response,
		status:     p.invitationStatus,
		responseTS: p.responseTS,
		statusTS:   p.statusTS,
	}

	return b, nil
}

func (b *participantModifier) SetResponse(resp api.AttendanceResponse) {
	b.response = resp
	b.responseTS = time.Now().UnixNano() / 1000
}

func (b *participantModifier) SetInvitationStatus(status api.InvitationStatus) {
	b.status = status
	b.statusTS = time.Now().UnixNano() / 1000
}

func (b *participantModifier) Build() (*Participant, error) {

	participant := &Participant{
		id:               b.pID,
		eventID:          b.eventID,
		name:             b.name,
		response:         b.response,
		invitationStatus: b.status,
		responseTS:       b.responseTS,
		statusTS:         b.statusTS,
	}

	return participant, nil
}
