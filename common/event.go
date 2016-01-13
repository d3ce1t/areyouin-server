package common

import (
	"errors"
	"time"
)

const (
	MIN_DIF_IN_START_DATE = 30 * time.Minute             // 30 minutes
	MAX_DIF_IN_START_DATE = 7 * 24 * time.Hour           // 1 week
	MIN_DIF_IN_END_DATE   = 30*time.Minute - time.Second // 30 minutes
	MAX_DIF_IN_END_DATE   = 24*time.Hour - time.Second   // 1 day
)

var (
	ErrInvalidStartDate = errors.New("invalid start date")
	ErrInvalidEndDate   = errors.New("invalid end date")
	ErrInvalidEventData = errors.New("invalidad event data")
)

func CreateNewEvent(event_id uint64, author_id uint64, author_name string, created_date int64, start_date int64,
	end_date int64, message string) *Event {
	event := &Event{
		EventId:      event_id,
		AuthorId:     author_id,
		AuthorName:   author_name,
		CreatedDate:  created_date,
		StartDate:    start_date,
		EndDate:      end_date,
		Message:      message,
		IsPublic:     false,
		NumAttendees: 0,
		NumGuests:    0,
	}
	return event
}

func (event *Event) IsValid() (bool, error) {

	if event.EventId == 0 || event.AuthorId == 0 || len(event.AuthorName) < 3 ||
		event.Message == "" || len(event.Message) < 8 || event.NumAttendees < 0 ||
		event.NumGuests < 0 || event.NumAttendees > event.NumGuests {
		return false, ErrInvalidEventData
	}

	if !event.IsValidStartDate() {
		return false, ErrInvalidStartDate
	}

	if !event.IsValidEndDate() {
		return false, ErrInvalidEndDate
	}

	return true, nil
}

func (event *Event) IsValidStartDate() bool {

	createdDate := UnixMillisToTime(event.CreatedDate)
	startDate := UnixMillisToTime(event.StartDate)

	if startDate.Before(createdDate.Add(MIN_DIF_IN_START_DATE)) ||
		startDate.After(createdDate.Add(MAX_DIF_IN_START_DATE)) {
		return false
	}

	return true
}

func (event *Event) IsValidEndDate() bool {

	startDate := UnixMillisToTime(event.StartDate)
	endDate := UnixMillisToTime(event.EndDate)

	if endDate.Before(startDate.Add(MIN_DIF_IN_END_DATE)) ||
		endDate.After(startDate.Add(MAX_DIF_IN_END_DATE)) {
		return false
	}

	return true
}
