package common

func (p *EventParticipant) SetFields(response AttendanceResponse, status MessageStatus) {
	p.Response = response
	p.Delivered = status
}
