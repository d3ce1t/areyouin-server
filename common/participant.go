package common

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
