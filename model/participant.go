package model

import (
	"time"

	"github.com/d3ce1t/areyouin-server/api"
)

type Participant struct {
	// Participant info
	id               int64
	name             string
	response         api.AttendanceResponse
	invitationStatus api.InvitationStatus

	// 0 if not attached to an event
	eventID int64

	// Used to compute the name version when stored in DB
	nameTS int64

	// Used to compute the response version when stored in DB
	responseTS int64

	// Used to compute the invitation status version when stored in DB
	statusTS int64
}

func NewParticipant(id int64, name string, response api.AttendanceResponse,
	status api.InvitationStatus) *Participant {
	timestamp := time.Now().UnixNano() / 1000
	return &Participant{
		id:               id,
		name:             name,
		response:         response,
		invitationStatus: status,
		nameTS:           timestamp,
		responseTS:       timestamp,
		statusTS:         timestamp,
	}
}

func newParticipantFromDTO(dto *api.ParticipantDTO) *Participant {
	return &Participant{
		id:               dto.UserID,
		eventID:          dto.EventID,
		name:             dto.Name,
		response:         dto.Response,
		invitationStatus: dto.InvitationStatus,
		nameTS:           dto.NameTS,
		responseTS:       dto.ResponseTS,
		statusTS:         dto.StatusTS,
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

func (p *Participant) EventID() int64 {
	return p.eventID
}

func (p *Participant) InvitationStatus() api.InvitationStatus {
	return p.invitationStatus
}

func (p *Participant) Equal(other *Participant) bool {
	return *p == *other
}

func (p *Participant) Clone() *Participant {
	copy := new(Participant)
	*copy = *p
	return copy
}

func (p *Participant) AsDTO() *api.ParticipantDTO {
	return &api.ParticipantDTO{
		UserID:           p.id,
		EventID:          p.eventID,
		Name:             p.name,
		Response:         p.response,
		InvitationStatus: p.invitationStatus,
		NameTS:           p.nameTS,
		ResponseTS:       p.responseTS,
		StatusTS:         p.statusTS,
	}
}
