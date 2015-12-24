package common

import (
	"time"
)

const (
	MIN_DIF_IN_END_DATE   = 60 * time.Minute   // 1 hour in minutes
	MAX_DIF_IN_END_DATE   = 24 * time.Hour     // 1 day in minutes
	MIN_DIF_IN_START_DATE = 30 * time.Minute   // 30 minutes
	MAX_DIF_IN_START_DATE = 24 * 7 * time.Hour // 1 week
)

func CreateNewEvent(event_id uint64, author_id uint64, author_name string, start_date int64,
	end_date int64, message string) *Event {
	event := &Event{
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

func (event *Event) IsValid() bool {

	if event.EventId == 0 || event.AuthorId == 0 || len(event.AuthorName) < 3 ||
		event.Message == "" || len(event.Message) < 8 || event.NumAttendees < 0 ||
		event.NumGuests < 0 || event.NumAttendees > event.NumGuests {
		return false
	}

	if !event.isValidStartDate() || !event.isValidEndDate() {
		return false
	}

	return true
}

func (event *Event) isValidStartDate() bool {
	createdDate := UnixMillisToTime(event.CreatedDate)
	startDate := UnixMillisToTime(event.StartDate)

	if startDate.After(createdDate.Add(MIN_DIF_IN_START_DATE)) &&
		startDate.Before(createdDate.Add(MAX_DIF_IN_START_DATE)) {
		return true
	}

	return false
}

func (event *Event) isValidEndDate() bool {
	startDate := UnixMillisToTime(event.StartDate)
	endDate := UnixMillisToTime(event.EndDate)

	if endDate.After(startDate.Add(MIN_DIF_IN_END_DATE)) &&
		endDate.Before(startDate.Add(MAX_DIF_IN_END_DATE)) {
		return true
	}

	return false
}
