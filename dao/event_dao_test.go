package dao

import (
	"math/rand"
	core "peeple/areyouin/common"
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {

	user_dao := NewUserDAO(session)

	core.DeleteFakeusers(user_dao)
	core.ClearEvents(session)
	core.CreateFakeUsers(user_dao)

	author, _ := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name,
		core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if event.AuthorId != author.Id || event.AuthorName != author.Name ||
		event.Message != "test" || event.EventId == 0 {
		t.Fail()
	}
}

// Insert a valid event
func TestEventInsert1(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author, _ := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(),
		core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.InsertEventCAS(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}
}

// Insert an event that already exists
func TestEventInsert2(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author, _ := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(),
		core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.InsertEventCAS(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}

	if ok, _ := event_dao.InsertEventCAS(event); ok {
		t.Fatal("It was expected an event duplicate error but It wasn't triggered")
	}
}

// Test inserting event with participants
func TestEventAddOrUpdateParticipants(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author, _ := user_dao.LoadByEmail("user1@foo.com")

	if author == nil {
		t.Fail()
	}

	// Create event
	event_id := uint64(16364452597203970)
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(),
		core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.InsertEventCAS(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}

	// Create participants
	friends, _ := user_dao.LoadFriends(author.Id, 0)

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
	event_participants, _ := dao.LoadAllParticipants(uint64(16364452597203970))
	if len(event_participants) != 7 {
		t.FailNow()
	}
}

func TestLoadEvent(t *testing.T) {
	dao := NewEventDAO(session)
	events, err := dao.LoadEvent(uint64(16364452597203970))
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.FailNow()
	}
}

func TestLoadEventAndParticipants(t *testing.T) {

	dao := NewEventDAO(session)
	events, err := dao.LoadEventAndParticipants(uint64(16364452597203970))
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 || len(events[0].Participants) != 7 {
		t.FailNow()
	}
}

// Test AddEventToUserInbox
func TestAddEventToUserInbox(t *testing.T) {

	event_dao := NewEventDAO(session)
	user_dao := NewUserDAO(session)
	author, _ := user_dao.LoadByEmail("user1@foo.com")

	// Create event
	event_id := idgen.GenerateID()
	event := core.CreateNewEvent(event_id, author.Id, author.Name, core.GetCurrentTimeMillis(),
		core.GetCurrentTimeMillis(), core.GetCurrentTimeMillis(), "test")

	if ok, err := event_dao.InsertEventCAS(event); !ok {
		t.Fatal("Coudn't insert the event in the database because", err)
	}

	if err := event_dao.AddEventToUserInbox(author.Id, event,
		core.AttendanceResponse_NO_ASSIST); err != nil {
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

func BenchmarkInsertEventCAS(b *testing.B) {

	event_dao := NewEventDAO(session)
	date := core.GetCurrentTimeMillis()

	for n := 0; n < b.N; n++ {

		event_id := idgen.GenerateID()
		event := core.CreateNewEvent(event_id, 29150727158891520, "Test", date, date, date, "test")

		if _, err := event_dao.InsertEventCAS(event); err != nil {
			b.Fatal("Coudn't insert the event in the database because", err)
		}
	}
}

func BenchmarkInsertEvent(b *testing.B) {

	event_dao := NewEventDAO(session)
	date := core.GetCurrentTimeMillis()

	for n := 0; n < b.N; n++ {

		event_id := idgen.GenerateID()
		event := core.CreateNewEvent(event_id, 29150727158891520, "Test", date, date, date, "test")

		if err := event_dao.InsertEvent(event); err != nil {
			b.Fatal("Coudn't insert the event in the database because", err)
		}
	}
}

func BenchmarkInsertEventAndGuests1(b *testing.B) {

	event_dao := NewEventDAO(session)
	date := core.GetCurrentTimeMillis()
	author_id := uint64(29150727158891520)

	for n := 0; n < b.N; n++ {
		event_id := idgen.GenerateID()
		event := core.CreateNewEvent(event_id, author_id, "Test", date, date, date, "test")

		if err := event_dao.InsertEvent(event); err != nil {
			b.Fatal("Coudn't insert the event in the database because", err)
		}

		if err := event_dao.AddOrUpdateParticipants(event_id, participants100); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInsertEventAndGuests2(b *testing.B) {

	event_dao := NewEventDAO(session)
	date := core.GetCurrentTimeMillis()
	author_id := uint64(29150727158891520)

	for n := 0; n < b.N; n++ {

		event_id := idgen.GenerateID()
		event := core.CreateNewEvent(event_id, author_id, "Test", date, date, date, "test")
		event.Participants = participants100

		if err := event_dao.InsertEventAndParticipants(event); err != nil {
			b.Fatal("Coudn't insert the event in the database because", err)
		}
	}
}

func BenchmarkReadSameEvent(b *testing.B) {

	event_dao := NewEventDAO(session)

	for n := 0; n < b.N; n++ {
		if _, err := event_dao.LoadEvent(29170915539421088); err != nil {
			b.Fatal("Coudn't read event", err)
		}
	}
}

func BenchmarkReadRandomEvent(b *testing.B) {

	event_dao := NewEventDAO(session)

	for n := 0; n < b.N; n++ {
		index := rand.Intn(len(eventIds10000))
		if _, err := event_dao.LoadEvent(eventIds10000[index]); err != nil {
			b.Fatal("Coudn't read event", err)
		}
	}
}

func BenchmarkReadSameEventAndParticipants(b *testing.B) {

	event_dao := NewEventDAO(session)

	for n := 0; n < b.N; n++ {
		if _, err := event_dao.LoadEventAndParticipants(29170915539421088); err != nil {
			b.Fatal("Coudn't read event", err)
		}
	}
}

func BenchmarkReadRandomEventAndParticipants(b *testing.B) {

	rand.Seed(time.Now().UnixNano())
	event_dao := NewEventDAO(session)

	for n := 0; n < b.N; n++ {
		index := rand.Intn(len(eventIds10000))
		if _, err := event_dao.LoadEventAndParticipants(eventIds10000[index]); err != nil {
			b.Fatal("Coudn't read event", err)
		}
	}
}

func BenchmarkReadEventAndParticipants100(b *testing.B) {

	rand.Seed(time.Now().UnixNano())
	event_dao := NewEventDAO(session)

	for n := 0; n < b.N; n++ {
		if _, err := event_dao.LoadEventAndParticipants(eventIds10000[:100]...); err != nil {
			b.Fatal("Coudn't read event", err)
		}
	}
}

func BenchmarkReadRandomEventAndParticipants100(b *testing.B) {

	rand.Seed(time.Now().UnixNano())
	event_dao := NewEventDAO(session)

	for n := 0; n < b.N; n++ {
		if _, err := event_dao.LoadEventAndParticipants(eventIds10000unsorted[:100]...); err != nil {
			b.Fatal("Coudn't read event", err)
		}
	}
}
