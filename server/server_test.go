package main

import (
	core "areyouin/common"
	proto "areyouin/protocol"
	"flag"
	"os"
	"testing"
)

var server *Server

func TestMain(m *testing.M) {
	server = NewTestServer()
	core.ClearUserAccounts(server.dbsession)
	core.ClearEvents(server.dbsession)
	core.CreateFakeUsers(server.NewUserDAO())
	flag.Parse()
	os.Exit(m.Run())
}

func TestDispatchEvent(t *testing.T) {

	user_dao := server.NewUserDAO()
	event_dao := server.NewEventDAO()
	author := user_dao.LoadByEmail("user1@foo.com")

	// Create event
	event_id := server.GetNewID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	// Prepare participants
	participants_list := server.createParticipantsFromFriends(author.Id)
	participant := author.AsParticipant()
	participant.SetFields(proto.AttendanceResponse_ASSIST, proto.MessageStatus_NO_DELIVERED)
	participants_list = append(participants_list, participant)

	// Insert event
	if ok, err := event_dao.Insert(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}

	// Dispatch: Put event to participants' inbox
	for _, p := range participants_list {
		if err := server.ds.dispatchEvent(event, p); err != nil {
			t.Fatal(err)
		}
	}

	// Check users' inbox
	for _, p := range participants_list {

		if events, err := event_dao.LoadUserEvents(p.UserId, 0); err == nil {

			t.Log("Participant", p.UserId, "has", len(events))

			found := false
			i := 0

			for !found && i < len(events) {
				if events[i].EventId == event.EventId {
					found = true
				}
				i++
			}

			if !found {
				t.Fatal("Event", event.EventId, "cannot be found on users inbox")
			}

		} else {
			t.FailNow()
		}
	} // for
}

func TestPublishEvent(t *testing.T) {

	user_dao := server.NewUserDAO()
	author := user_dao.LoadByEmail("user1@foo.com")

	// Create event
	event_id := server.GetNewID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	// Prepare participants
	participants_list := server.createParticipantsFromFriends(author.Id)
	participant := author.AsParticipant()
	participant.SetFields(proto.AttendanceResponse_ASSIST, proto.MessageStatus_NO_DELIVERED)
	participants_list = append(participants_list, participant)

	// Publish Event
	server.PublishEvent(event, participants_list)
}
