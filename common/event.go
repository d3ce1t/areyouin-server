package common

import (
	"errors"
	"time"
	"peeple/areyouin/idgen"
)

const (
	MIN_DIF_IN_START_DATE = 30 * time.Minute             // 30 minutes
	MAX_DIF_IN_START_DATE = 365 * 24 * time.Hour         // 1 year
	MIN_DIF_IN_END_DATE   = 30*time.Minute - time.Second // 30 minutes (from start date)
	MAX_DIF_IN_END_DATE   = 7*24*time.Hour - time.Second // 1 week (from start date)
	EVENT_PICTURE_MAX_WIDTH    = 1280
	EVENT_PICTURE_MAX_HEIGHT   = 720
)

var (
	ErrInvalidStartDate = errors.New("invalid start date")
	ErrInvalidEndDate   = errors.New("invalid end date")
	ErrInvalidEventData = errors.New("invalidad event data")
)

func CreateNewEvent(author_id int64, author_name string, created_date int64, start_date int64,
	end_date int64, message string) *Event {
	event := &Event{
		EventId:       idgen.NewID(),
		AuthorId:      author_id,
		AuthorName:    author_name,
		CreatedDate:   created_date,
		InboxPosition: start_date,
		StartDate:     start_date,
		EndDate:       end_date,
		Message:       message,
		IsPublic:      false,
		NumAttendees:  0,
		NumGuests:     0,
		Participants: make(map[int64]*EventParticipant),
	}
	return event
}

func (event *Event) IsValid() (bool, error) {

	if event.EventId == 0 || event.AuthorId == 0 ||
		len(event.AuthorName) < USER_NAME_MIN_LENGTH || len(event.AuthorName) > USER_NAME_MAX_LENGTH ||
		event.Message == "" || len(event.Message) < EVENT_DESCRIPTION_MIN_LENGHT ||
		len(event.Message) > EVENT_DESCRIPTION_MAX_LENGHT || event.NumAttendees < 0 ||
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

	// I need only minute precision in order to emulate the same checking performed
	// by the client.
	createdDateMin := UnixMillisToTime(event.CreatedDate - (event.CreatedDate % 60000))
	startDate := UnixMillisToTime(event.StartDate)

	if startDate.Before(createdDateMin.Add(MIN_DIF_IN_START_DATE)) ||
		startDate.After(createdDateMin.Add(MAX_DIF_IN_START_DATE)) {
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

func (event *Event) SetParticipants(participants map[int64]*EventParticipant) {
	event.Participants = participants
	if participants != nil {
		event.NumGuests = int32(len(participants))
	}
}

func (event *Event) AddParticipant(participant *EventParticipant) {
	event.Participants[participant.UserId] = participant
	event.NumGuests = int32(len(event.Participants))
}

func (event *Event) GetEventWithoutParticipants() *Event {
	new_event := &Event{}
	*new_event = *event // copy
	new_event.SetParticipants(nil)
	return new_event
}

func (event *Event) GetStatus() EventState {

	currentDate := time.Now()
	startDate := UnixMillisToTime(event.StartDate)
	endDate := UnixMillisToTime(event.EndDate)

	if event.State == EventState_CANCELLED {
		return EventState_CANCELLED
	} else if startDate.After(currentDate) {
		return EventState_NOT_STARTED
	} else if endDate.Before(currentDate) || endDate.Equal(currentDate) {
		return EventState_FINISHED
	} else {
		return EventState_ONGOING
	}
}

func (event *Event) CloneFull() *Event {
	eventCopy := new(Event)
  *eventCopy = *event
  eventCopy.Participants = make(map[int64]*EventParticipant)
  for k, v := range event.Participants {
		eventCopy.Participants[k] = v
  }
	return eventCopy
}

func (event *Event) CloneEmpty() *Event {
	eventCopy := new(Event)
  *eventCopy = *event
	eventCopy.Participants = nil
	//eventCopy.NumGuests = 0
	//eventCopy.NumAttendees = 0
	return eventCopy
}
