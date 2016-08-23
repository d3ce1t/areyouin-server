package main

import (
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/idgen"
	"peeple/areyouin/utils"
	"time"
)

// Test write workload to create an event
func testCreateEvent(testNumber int) (time.Duration, error) {

	participant := &api.ParticipantDTO{
		UserId:           idgen.NewID(),
		Name:             "Bench User",
		Response:         api.AttendanceResponse_NO_ASSIST,
		InvitationStatus: api.InvitationStatus_SERVER_DELIVERED,
	}

	event := &api.EventDTO{
		Id:            idgen.NewID(),
		AuthorId:      participant.UserId,
		AuthorName:    participant.Name,
		Description:   "This is a test event with a few words only",
		CreatedDate:   utils.GetCurrentTimeMillis(),
		InboxPosition: utils.GetCurrentTimeMillis(),
		StartDate:     utils.GetCurrentTimeMillis(),
		EndDate:       utils.GetCurrentTimeMillis(),
	}

	startTime := time.Now()

	// Insert event
	err := eventDAO.Insert(event)
	if err != nil {
		log.Printf("TestCreateEvent %v Error: %v", testNumber, err)
		return 0, err
	}

	// Insert participants
	for i := 0; i < 10; i++ {
		err = eventDAO.AddParticipantToEvent(participant, event)
		if err != nil {
			log.Printf("TestCreateEvent %v Error: %v", testNumber, err)
			return 0, err
		}
	}

	// Update num guests
	if _, err := eventDAO.SetNumGuests(event.Id, 10); err != nil {
		log.Printf("TestCreateEvent %v Error: %v", testNumber, err)
		return 0, err
	}

	return time.Now().Sub(startTime), nil
}
