package common

import (
	proto "peeple/areyouin/protocol"
)

func CreateNewEvent(event_id uint64, author_id uint64, author_name string, start_date int64,
	end_date int64, message string) *proto.Event {
	event := &proto.Event{
		EventId:      event_id,
		AuthorId:     author_id,
		AuthorName:   author_name,
		StartDate:    start_date,
		EndDate:      end_date,
		Message:      message,
		IsPublic:     false,
		NumAttendees: 0,
		NumGuests:    0,
		CreatedDate:  GetCurrentTimeMillis(),
	}
	return event
}
