package dao

import (
	core "peeple/areyouin/common"
	proto "peeple/areyouin/protocol"
	"testing"
)

func TestCreateEvent(t *testing.T) {

	user_dao := NewUserDAO(session)

	core.ClearUserAccounts(session)
	core.ClearEvents(session)
	core.CreateFakeUsers(user_dao)

	author := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if event.AuthorId != author.Id || event.AuthorName != author.Name ||
		event.Message != "test" || event.EventId == 0 {
		t.Fail()
	}
}

// Insert a valid event
func TestEventInsert1(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.Insert(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}
}

// Insert an event that already exists
func TestEventInsert2(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.Insert(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}

	if ok, _ := event_dao.Insert(event); ok {
		t.Fatal("It was expected an event duplicate error but It wasn't triggered")
	}
}

// Test inserting event with participants
func TestEventAddOrUpdateParticipants(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := uint64(16364452597203970)
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.Insert(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}

	// Create participants
	friends := user_dao.LoadFriends(author.Id, 0)

	if friends != nil {
		participants := core.CreateParticipantsFromFriends(author.Id, friends)
		if err := event_dao.AddOrUpdateParticipants(event_id, participants); err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatal("No friends or error")
	}
}

func TestLoadParticipants(t *testing.T) {
	dao := NewEventDAO(session)
	event_participants := dao.LoadAllParticipants(uint64(16364452597203970))
	if len(event_participants) != 7 {
		t.FailNow()
	}
}

// Test AddEventToUserInbox
func TestAddEventToUserInbox(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author := user_dao.LoadByEmail("user1@foo.com")

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.Insert(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}

	if err := event_dao.AddEventToUserInbox(author.Id, event,
		proto.AttendanceResponse_NO_ASSIST); err != nil {
		t.Fatal(err)
	}

	if events, err := event_dao.LoadUserEvents(author.Id, 0); err == nil {

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
		t.Fatal(err)
	}
}
