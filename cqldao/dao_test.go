package cqldao

import (
	"flag"
	"fmt"
	"log"
	"os"
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"testing"
	"time"
)

var session *GocqlSession
var eventsGenerated int64
var timeLinesGenerated int64

func TestMain(m *testing.M) {
	session = NewSession("areyouin_test", 4, "192.168.1.10")

	if err := session.Connect(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(-1)
	}

	if err := session.Query("TRUNCATE event").Exec(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(-1)
	}

	d1 := NewTimeLineDAO(session)
	d1.DeleteAll()
	flag.Parse()
	os.Exit(m.Run())
}

func generateEvent(numParticipants int) *api.EventDTO {

	eventID := int64(eventsGenerated + 1)
	createdDate := time.Now().UTC()
	timestamp := time.Now().UnixNano() / 1000
	startDate := createdDate.Add(35 * time.Minute)
	endDate := createdDate.Add(1 * time.Hour)

	eventDTO := &api.EventDTO{
		Id:            eventID,
		AuthorId:      eventID,
		AuthorName:    fmt.Sprintf("Author %v", eventID),
		Description:   fmt.Sprintf("Test %v", eventID),
		CreatedDate:   utils.TimeToMillis(createdDate),
		InboxPosition: utils.TimeToMillis(startDate),
		StartDate:     utils.TimeToMillis(startDate),
		EndDate:       utils.TimeToMillis(endDate),
		Timestamp:     timestamp,
		Participants:  make(map[int64]*api.ParticipantDTO),
	}

	for _, p := range generateParticipants(numParticipants) {
		p.EventID = eventDTO.Id
		p.NameTS = timestamp
		p.ResponseTS = timestamp
		p.StatusTS = timestamp
		eventDTO.Participants[p.UserID] = p
	}

	eventsGenerated++

	return eventDTO
}

func generateEvents(numEvents int) []*api.EventDTO {

	events := make([]*api.EventDTO, 0, numEvents)

	for i := 0; i < numEvents; i++ {
		eventDTO := generateEvent(i % 10)
		eventDTO.Cancelled = i%4 == 0
		events = append(events, eventDTO)
	}

	return events
}

func generateParticipants(numParticipants int) []*api.ParticipantDTO {

	participants := make([]*api.ParticipantDTO, 0, numParticipants)

	for i := 0; i < numParticipants; i++ {
		timestamp := time.Now().UnixNano() / 1000
		userID := int64(i + 1)
		pDTO := &api.ParticipantDTO{
			UserID:           userID,
			Name:             fmt.Sprintf("User %v", userID),
			Response:         api.AttendanceResponse_NO_RESPONSE,
			InvitationStatus: api.InvitationStatus_SERVER_DELIVERED,
			ResponseTS:       timestamp,
			StatusTS:         timestamp,
		}

		participants = append(participants, pDTO)
	}

	return participants
}

func generateTimelineEntries(origin time.Time, numEntries int) []*api.TimeLineEntryDTO {

	entries := make([]*api.TimeLineEntryDTO, 0, numEntries)

	for i := 0; i < numEntries; i++ {

		newDate := origin.AddDate(i/3, i%12, i%30)

		dto := &api.TimeLineEntryDTO{
			EventID:  timeLinesGenerated + 1,
			Position: newDate,
		}

		entries = append(entries, dto)
		timeLinesGenerated++
	}

	return entries
}
