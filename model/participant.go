package model

import (
	"peeple/areyouin/api"
)

type Participant struct {
	id               int64
	name             string
	response         api.AttendanceResponse
	invitationStatus api.InvitationStatus
}

func NewParticipant(id int64, name string, response api.AttendanceResponse,
	status api.InvitationStatus) *Participant {
	return &Participant{
		id:               id,
		name:             name,
		response:         response,
		invitationStatus: status,
	}
}

func newParticipantFromDTO(dto *api.ParticipantDTO) *Participant {
	return &Participant{
		id:               dto.UserId,
		name:             dto.Name,
		response:         dto.Response,
		invitationStatus: dto.InvitationStatus,
	}
}

func (p *Participant) Id() int64 {
	return p.id
}

func (p *Participant) Name() string {
	return p.name
}

func (p *Participant) Response() api.AttendanceResponse {
	return p.response
}

func (p *Participant) InvitationStatus() api.InvitationStatus {
	return p.invitationStatus
}

func (p *Participant) AsDTO() *api.ParticipantDTO {
	return &api.ParticipantDTO{
		UserId:           p.Id(),
		Name:             p.Name(),
		Response:         p.Response(),
		InvitationStatus: p.InvitationStatus(),
	}
}

/*func (p *EventParticipant) SetFields(response AttendanceResponse, status MessageStatus) {
	p.Response = response
	p.Delivered = status
}*/

func (p *Participant) asAnonym() *Participant {
	return &Participant{
		id:               p.id,
		response:         p.response,
		invitationStatus: p.invitationStatus,
	}
}
