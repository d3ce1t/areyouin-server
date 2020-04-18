package main

/*import (
	"flag"
	"os"
	core "github.com/d3ce1t/areyouin-server/common"
	proto "github.com/d3ce1t/areyouin-server/protocol"
	"testing"
)

var server *Server

func TestMain(m *testing.M) {
	server = NewTestServer()
	core.DeleteFakeusers(server.NewUserDAO())
	core.ClearEvents(server.dbsession)
	core.CreateFakeUsers(server.NewUserDAO(), server.NewFriendDAO())
	flag.Parse()
	os.Exit(m.Run())
}

func TestOperationPermission(t *testing.T) {

	fakeCallback := func(packet_type proto.PacketType, message proto.Message, session *AyiSession) {
		checkAuthenticated(session)
	}

	server.RegisterCallback(proto.M_CREATE_EVENT, fakeCallback)

	session := NewSession(nil, server)

	packet := proto.NewPacket(proto.VERSION_2).CreateEvent("Test",
		core.GetCurrentTimeMillis(),
		core.GetCurrentTimeMillis()+3600*1000,
		[]uint64{1, 2, 3, 4, 5, 6, 7, 8})

	// Session no authenticated, so I expect ErrAuthRequired, if not test fail
	if err := server.serveMessage(packet, session); err != ErrAuthRequired {
		t.Fatal(err)
	}

	// Session is authenticated, so any error should happen
	session.IsAuth = true
	if err := server.serveMessage(packet, session); err != nil {
		t.Fatal(err)
	}
}*/

/*func TestDispatchEvent(t *testing.T) {

	user_dao := server.NewUserDAO()
	event_dao := server.NewEventDAO()
	author, err := user_dao.LoadByEmail("user1@foo.com")

	if err != nil {
		t.Fatal(err)
	}

	// Create event
	event_id := server.GetNewID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(),
		core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	// Prepare participants
	participants := server.createParticipantsFromFriends(author.Id)
	participant := author.AsParticipant()
	participant.SetFields(core.AttendanceResponse_ASSIST, core.MessageStatus_NO_DELIVERED)
	participants[participant.UserId] = participant
	event.SetParticipants(participants)

	// Insert event
	if ok, err := event_dao.InsertEventAndParticipants(event); !ok {
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
}*/

/*func TestPublishEvent(t *testing.T) {

	user_dao := server.NewUserDAO()
	author, _ := user_dao.LoadByEmail("user1@foo.com")

	// Create event
	event_id := server.GetNewID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	// Prepare participants
	participants_list := server.createParticipantsFromFriends(author.Id)
	participant := author.AsParticipant()
	participant.SetFields(core.AttendanceResponse_ASSIST, core.MessageStatus_NO_DELIVERED)
	participants_list = append(participants_list, participant)

	// Publish Event
	server.PublishEvent(event, participants_list)
}*/
