package main

import (
	"log"
	"time"

	"github.com/d3ce1t/areyouin-server/api"
	"github.com/d3ce1t/areyouin-server/idgen"
	"github.com/d3ce1t/areyouin-server/utils"
)

// Test write workload to create an event
func testCreateEvent(testNumber int) (time.Duration, error) {

	pID := idgen.NewID()
	pName := "Bench User"

	event := &api.EventDTO{
		Id:            idgen.NewID(),
		AuthorId:      pID,
		AuthorName:    pName,
		Description:   "This is a test event with a few words only",
		CreatedDate:   utils.GetCurrentTimeMillis(),
		InboxPosition: utils.GetCurrentTimeMillis(),
		StartDate:     utils.GetCurrentTimeMillis(),
		EndDate:       utils.GetCurrentTimeMillis(),
	}

	participant := &api.ParticipantDTO{
		UserID:           pID,
		Name:             pName,
		Response:         api.AttendanceResponse_NO_ASSIST,
		InvitationStatus: api.InvitationStatus_SERVER_DELIVERED,
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
		err = eventDAO.InsertParticipant(participant)
		if err != nil {
			log.Printf("TestCreateEvent %v Error: %v", testNumber, err)
			return 0, err
		}
	}

	return time.Now().Sub(startTime), nil
}
