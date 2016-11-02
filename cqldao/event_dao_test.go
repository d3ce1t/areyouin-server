package cqldao

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"testing"
	"time"
)

func TestEventDAO_Insert(t *testing.T) {
	events := generateEvents(100)
	d := NewEventDAO(session)
	for _, eventDTO := range events {
		if err := d.Insert(eventDTO); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEventDAO_Replace(t *testing.T) {

	// Insert 100 events
	events := generateEvents(100)
	d := NewEventDAO(session)
	for _, eventDTO := range events {
		if err := d.Insert(eventDTO); err != nil {
			t.Fatal(err)
		}
	}

	// Replace
	for i, eventDTO := range events {

		copy := eventDTO.Clone()
		copy.Timestamp = time.Now().UnixNano() / 1000

		if i%10 == 0 {
			// Change dates
			currentTime := time.Now().UTC()
			startDate := currentTime.Add(35 * time.Minute)
			endDate := currentTime.Add(1 * time.Hour)
			copy.StartDate = utils.TimeToMillis(startDate)
			copy.EndDate = utils.TimeToMillis(endDate)
		} else if i%6 == 0 {
			// Add participants
			for _, p := range generateParticipants(20) {
				p.EventID = copy.Id
				p.ResponseTS = copy.Timestamp
				p.StatusTS = copy.Timestamp
				copy.Participants[p.UserID] = p
			}
		} else if i%5 == 0 {
			// Cancel event
			copy.Cancelled = true
			copy.InboxPosition = utils.GetCurrentTimeMillis()
		}

		if err := d.Replace(eventDTO, copy); err != nil {
			t.Fatalf("Error replacing event: %v (i=%v)", err, i)
		}
	}
}

func TestEventDAO_InsertParticipant(t *testing.T) {

	// Insert 1 event
	event := generateEvent(1)
	d := NewEventDAO(session)
	if err := d.Insert(event); err != nil {
		t.Fatal(err)
	}

	// Update participant
	p := event.Participants[1]
	p.Response = api.AttendanceResponse_ASSIST
	p.ResponseTS = time.Now().UnixNano() / 1000

	if err := d.InsertParticipant(p); err != nil {
		t.Fatal(err)
	}

	// Check
	loadedParticipant, err := d.LoadParticipant(p.UserID, p.EventID)
	if err != nil {
		t.Fatal(err)
	}

	if !api.EqualParticipantDTO(p, loadedParticipant) {
		t.Fatal("Written participant and loaded don't match")
	}
}

func TestEventDAO_LoadEvents(t *testing.T) {

	// Insert 100 events
	events := generateEvents(100)
	d := NewEventDAO(session)
	for _, eventDTO := range events {
		if err := d.Insert(eventDTO); err != nil {
			t.Fatal(err)
		}
	}

	// get ids
	eventIds := make([]int64, 0, len(events))
	for _, eventDTO := range events {
		eventIds = append(eventIds, eventDTO.Id)
	}

	// Read written events
	storedEvents, err := d.LoadEvents(eventIds...)
	if err != nil {
		t.Fatalf("Error loading events: %v", err)
	}

	// Check count
	if len(storedEvents) == 0 || len(storedEvents) != len(events) {
		t.Fatal("Loaded events count mismatch")
	}

	// Compare
	for i, readEvent := range storedEvents {
		if !api.EqualEventDTO(readEvent, events[i]) {

			t.Log("Event read from database is not equal to the one that was written")
			t.Log("Read Event:")
			for _, p := range readEvent.Participants {
				t.Logf("%v", *p)
			}
			t.Log("Written Event:")
			for _, p := range events[i].Participants {
				t.Logf("%v", *p)
			}
			t.FailNow()
		}
	}
}

func TestEventDAO_LoadParticipant(t *testing.T) {

	// Insert one event
	event := generateEvent(1)
	d := NewEventDAO(session)
	if err := d.Insert(event); err != nil {
		t.Fatal(err)
	}

	// Read participant
	participant, err := d.LoadParticipant(1, event.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !api.EqualParticipantDTO(participant, event.Participants[1]) {
		t.Fatal("Read participant is different than the written one")
	}
}

func TestEventDAO_Event_Timestamp(t *testing.T) {

	// Insert one event with one participant
	event := generateEvent(1)
	d := NewEventDAO(session)
	if err := d.Insert(event); err != nil {
		t.Fatal(err)
	}

	// Read timestamp
	var writtenTS, ts int64
	stmt := `SELECT DISTINCT event_timestamp, writetime(author_id) as ts FROM event WHERE event_id = ?`
	q := session.Query(stmt, event.Id)
	if err := q.Scan(&writtenTS, &ts); err != nil {
		t.Fatal(err)
	}

	// Check
	if event.Timestamp != writtenTS || writtenTS != ts {
		t.Fatal("Timestamp doesn't match")
	}
}

func TestEventDAO_Participant_Timestamp(t *testing.T) {

	// Insert one event with one participant
	event := generateEvent(1)
	d := NewEventDAO(session)
	if err := d.Insert(event); err != nil {
		t.Fatal(err)
	}

	// Read timestamp
	var respTS, statusTS int64
	p := event.Participants[1]
	stmt := `SELECT writetime(guest_response) as respTS, writetime(guest_status) as statusTS
		FROM event WHERE event_id = ? AND guest_id = ?`
	q := session.Query(stmt, event.Id, p.UserID)
	if err := q.Scan(&respTS, &statusTS); err != nil {
		t.Fatal(err)
	}

	// Check
	if event.Timestamp != respTS || event.Timestamp != statusTS {
		t.Fatal("Timestamp doesn't match")
	}
}

/*func BenchmarkInsertEventCAS(b *testing.B) {

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
*/
