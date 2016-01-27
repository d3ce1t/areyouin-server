package common

type Participant struct {
	EventParticipant
	EventId        uint64
	EventStartDate int64
}

func (p *EventParticipant) SetFields(response AttendanceResponse, status MessageStatus) {
	p.Response = response
	p.Delivered = status
}

func (p *EventParticipant) AsAnonym() *EventParticipant {
	return &EventParticipant{
		UserId:    p.UserId,
		Response:  p.Response,
		Delivered: p.Delivered,
	}
}
