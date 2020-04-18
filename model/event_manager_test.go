package model

import (
	"testing"
	"time"

	"github.com/d3ce1t/areyouin-server/api"
	"github.com/d3ce1t/areyouin-server/utils"
)

func TestNewEvent_TimeRules(t *testing.T) {

	cd := utils.CreateDate

	var tests = []struct {
		createdDate time.Time
		startDate   time.Time
		endDate     time.Time
		expected    error
	}{
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 0), ErrInvalidStartDate},
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 30), cd(2016, 1, 1, 12, 0), ErrInvalidEndDate},
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 30), cd(2016, 1, 1, 12, 30), ErrInvalidEndDate},
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 30), cd(2016, 1, 1, 13, 0), nil},
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 0).Add(366 * 24 * time.Hour), cd(2016, 1, 1, 13, 0), ErrInvalidStartDate},
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 0).Add(365 * 24 * time.Hour), cd(2016, 1, 1, 13, 0), ErrInvalidEndDate},
		{cd(2016, 1, 1, 12, 0), cd(2016, 12, 31, 12, 0), cd(2016, 1, 1, 13, 0), ErrInvalidEndDate},
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 13, 0), cd(2016, 1, 8, 14, 0), ErrInvalidEndDate},
		{cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 13, 0), cd(2016, 1, 8, 13, 0), nil},
	}

	author := users[1]

	for i, test := range tests {
		_, err := testModel.Events.NewEvent(author,
			test.createdDate, test.startDate, test.endDate, "this is a test description", nil)
		if err != test.expected {
			t.Fatalf("test %v: Expected '%v' but got '%v'", i, test.expected, err)
		}
	}
}

func TestNewEvent_AuthorAndDescription(t *testing.T) {

	author := users[1]

	// Create an INVAID user account object
	invalidAuthor := &UserAccount{}

	// Create a valid user account object (not persisted)
	notPersistedUser, err := NewUserAccount("Foo", "foo@example.com", "12345", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	cd := utils.CreateDate

	var tests = []struct {
		author      *UserAccount
		description string
		expected    error
	}{
		{nil, "", ErrInvalidAuthor},
		{invalidAuthor, "", ErrInvalidAuthor},
		{notPersistedUser, "", ErrInvalidAuthor},
		{author, "", ErrInvalidDescription},
		{author, "Short", ErrInvalidDescription},
		{author, "12345678901234", ErrInvalidDescription},
		{author, "123456789012345", nil},
	}

	for i, test := range tests {
		_, err := testModel.Events.NewEvent(test.author,
			cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 30), cd(2016, 1, 1, 13, 0), test.description, []int64{})
		if err != test.expected {
			t.Fatalf("test %v: Expected '%v' but got '%v'", i, test.expected, err)
		}
	}
}

func TestNewEvent_Participants(t *testing.T) {

	cd := utils.CreateDate

	authorNoFriends := users[0]
	authorWithFriends := users[1]

	friends, err := testModel.Friends.GetAllFriends(authorWithFriends.id)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		author          *UserAccount
		participants    []int64
		numParticipants int
	}{
		{authorNoFriends, nil, 1},   // Author can create event even when he/she has no friends
		{authorWithFriends, nil, 1}, // Author can create event with no friends
		{authorWithFriends, FriendKeys(friends), len(friends) + 1},
		{authorNoFriends, []int64{1, 2, 3, 4}, 1}, // Not found participants are ignored
		{authorNoFriends, FriendKeys(friends), 1}, // Cannot invite other user's friends
	}

	for i, test := range tests {

		event, err := testModel.Events.NewEvent(test.author,
			cd(2016, 1, 1, 12, 0), cd(2016, 1, 1, 12, 30), cd(2016, 1, 1, 13, 0),
			"123456789012345", test.participants)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if event.NumGuests() != test.numParticipants {
			t.Fatalf("test %v: Expected '%v' participants but got '%v' participants",
				i, test.numParticipants, event.NumGuests())
		}
	}
}

func TestNewEvent_SaveEvent_New(t *testing.T) {

	events, err := generateEvents(5, testModel)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Prepare data

	// Persist first event
	if err := testModel.Events.SaveEvent(events[0]); err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Change created date to be out of creation window
	events[2].createdDate = utils.GetCurrentTimeUTC().
		Truncate(time.Second).Add(-2 * time.Minute).Add(-time.Second)
	events[3].createdDate = utils.GetCurrentTimeUTC().
		Truncate(time.Second).Add(2 * time.Minute).Add(time.Second)

	// Change start date to a date in the past
	events[4].startDate = utils.GetCurrentTimeUTC().Truncate(time.Second).Add(-6 * time.Hour)
	events[4].endDate = events[4].startDate.Add(1 * time.Hour)

	// Describe test

	var tests = []struct {
		event    *Event
		expected error
	}{
		{nil, ErrInvalidEvent},      // Invalid event
		{&Event{}, ErrInvalidEvent}, // Zero event
		{events[0], nil},            // Persisted event, do not fail but gets ignored
		{events[1], nil},            // Valid event
		{events[2], ErrEventOutOfCreationWindow},
		{events[3], ErrEventOutOfCreationWindow},
		// SaveEvent assumes an event is valid despite it has a start date in the past
		{events[4], nil},
	}

	// Run test

	for i, test := range tests {
		err := testModel.Events.SaveEvent(test.event)
		if err != test.expected {
			t.Fatalf("test %v: Expected '%v' but got '%v'", i, test.expected, err)
		}
	}

	for i, writtenEvent := range []*Event{events[0], events[1], events[4]} {
		loadedEvent, err := testModel.Events.LoadEvent(writtenEvent.id)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !loadedEvent.Equal(writtenEvent) {
			t.Logf("test %v: Event mismatch", i)
			t.Logf("Loaded Event: %v", loadedEvent)
			t.Logf("Written Event: %v", writtenEvent)
			t.FailNow()
		}
	}
}

func TestNewEvent_SaveEvent_Modified(t *testing.T) {

	// Prepare data
	events, err := generateEvents(8, testModel)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	for i, event := range events {

		if i == 4 {
			// Simulate that oldEvent is already started.
			event.startDate = utils.GetCurrentTimeUTC().Truncate(time.Second).Add(-6 * time.Hour)
			event.endDate = event.startDate.Add(1 * time.Hour)
		}

		if i > 0 {
			if err := testModel.Events.SaveEvent(event); err != nil {
				t.Fatalf("Error: %v", err)
			}
		}

		// Modify event
		b := testModel.Events.NewEventModifier(event, event.authorID).
			SetDescription("123456789012345 hello world")

		if i == 2 {
			b.SetModifiedDate(utils.GetCurrentTimeUTC().
				Truncate(time.Second).
				Add(-2 * time.Minute).Add(-time.Second))
		} else if i == 3 {
			b.SetModifiedDate(utils.GetCurrentTimeUTC().
				Truncate(time.Second).
				Add(2 * time.Minute).Add(time.Second))
		} else if i == 5 {
			b.ParticipantAdder().AddUserID(users[2].id).
				AddUserID(users[3].id)
		} else if i == 6 {
			b.SetCancelled(true)
		} else if i == 7 {
			currentDate := utils.GetCurrentTimeUTC()
			b.SetStartDate(currentDate.Add(48 * time.Hour)).
				SetEndDate(currentDate.Add(96 * time.Hour))
		}

		events[i], err = b.Build()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

	}

	// Describe test
	var tests = []struct {
		event    *Event
		expected error
	}{
		{events[0], nil}, // A modified event not previously persisted
		{events[1], nil}, // A event that modifies a persistent one
		{events[2], ErrEventOutOfCreationWindow},
		{events[3], ErrEventOutOfCreationWindow},
		{events[4], ErrEventNotWritable},
		{events[5], nil}, // Event with new participants
		{events[6], nil}, // Cancelled event
		{events[7], nil}, // Event with modified dates

	}

	// Run test
	for i, test := range tests {
		err := testModel.Events.SaveEvent(test.event)
		if err != test.expected {
			t.Fatalf("test %v: Expected '%v' but got '%v'", i, test.expected, err)
		}
	}

	for i, writtenEvent := range []*Event{events[0], events[1], events[5], events[6], events[7]} {
		loadedEvent, err := testModel.Events.LoadEvent(writtenEvent.id)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !loadedEvent.Equal(writtenEvent) {
			t.Logf("test %v: Event mismatch", i)
			t.FailNow()
		}
	}
}

func TestNewEvent_ChangeParticipantResponse(t *testing.T) {

	events, err := generateEvents(2, testModel)
	if err != nil {
		t.Fatal(err)
	}

	b := testModel.Events.NewEventModifier(events[1], users[1].id)
	b.ParticipantAdder().AddUserAccount(users[2]).
		AddUserAccount(users[3])
	events[1], err = b.Build()
	if err != nil {
		t.Fatal(err)
	}

	if err := testModel.Events.SaveEvent(events[1]); err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		userID   int64
		response api.AttendanceResponse
		event    *Event
		expected error
	}{
		{users[2].id, api.AttendanceResponse_MAYBE, nil, ErrInvalidEvent},       // nil event
		{users[2].id, api.AttendanceResponse_MAYBE, &Event{}, ErrInvalidEvent},  // zero event
		{users[2].id, api.AttendanceResponse_MAYBE, events[0], ErrInvalidEvent}, // not persisted event
		{users[2].id, api.AttendanceResponse_ASSIST, events[1], nil},
		{users[3].id, api.AttendanceResponse_MAYBE, events[1], nil},
		{0, api.AttendanceResponse_NO_ASSIST, events[1], ErrParticipantNotFound},
	}

	for i, test := range tests {
		_, err := testModel.Events.ChangeParticipantResponse(test.event.id, test.userID, test.response)
		if err != test.expected {
			t.Fatalf("test %v: Expected '%v' but got '%v'", i, test.expected, err)
		}
	}
}

func TestNewEvent_ChangeDeliveryState(t *testing.T) {

	events, err := generateEvents(2, testModel)
	if err != nil {
		t.Fatal(err)
	}

	b := testModel.Events.NewEventModifier(events[1], users[1].id)
	b.ParticipantAdder().AddUserAccount(users[2]).
		AddUserAccount(users[3])
	events[1], err = b.Build()
	if err != nil {
		t.Fatal(err)
	}

	if err := testModel.Events.SaveEvent(events[1]); err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		userID   int64
		state    api.InvitationStatus
		event    *Event
		expected error
	}{
		{users[2].id, api.InvitationStatus_SERVER_DELIVERED, nil, ErrInvalidEvent},       // nil event
		{users[2].id, api.InvitationStatus_SERVER_DELIVERED, &Event{}, ErrInvalidEvent},  // zero event
		{users[2].id, api.InvitationStatus_CLIENT_DELIVERED, events[0], ErrInvalidEvent}, // not persisted event
		{users[2].id, api.InvitationStatus_CLIENT_DELIVERED, events[1], nil},
		{users[3].id, api.InvitationStatus_CLIENT_DELIVERED, events[1], nil},
		{0, api.InvitationStatus_CLIENT_DELIVERED, events[1], ErrParticipantNotFound},
	}

	for i, test := range tests {
		_, err := testModel.Events.ChangeDeliveryState(test.userID, test.state, test.event)
		if err != test.expected {
			t.Fatalf("test %v: Expected '%v' but got '%v'", i, test.expected, err)
		}
	}
}

func TestNewEvent_GetRecentEvents(t *testing.T) {

	// Clean data
	if err := testModel.Events.deleteAllEvents(); err != nil {
		t.Fatal(err)
	}

	// Generate 10 events as users[1]
	generatedEvents = 0
	events := make([]*Event, 0, 10)
	for i := 0; i < 10; i++ {
		event, err := generateEvent(users[1], testModel)
		if err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}

	// Publish events and index
	writtenEventsIndex := make(map[int64]*Event)

	for i, event := range events {
		b := testModel.Events.NewEventModifier(event, users[1].id)
		b.ParticipantAdder().AddUserAccount(users[2]).
			AddUserAccount(users[3])
		modifiedEvent, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}
		if err := testModel.Events.SaveEvent(modifiedEvent); err != nil {
			t.Fatal(err)
		}
		writtenEventsIndex[modifiedEvent.id] = modifiedEvent
		events[i] = modifiedEvent
	}

	// Check inboxes
	for i := 1; i < len(users); i++ {

		inbox, err := testModel.Events.GetRecentEvents(users[i].id)
		if err != nil {
			t.Fatalf("Test %v: %v", i, err)
		}

		if len(inbox) != len(events) {
			t.Fatalf("Test %v: Num. events mismatch (%v != %v)", i, len(inbox), len(events))
		}

		for j, event := range inbox {
			writtenEvent := writtenEventsIndex[event.id]
			if !event.Equal(writtenEvent) {
				t.Logf("Test %v: event %v is not equal", i, j)
				t.Logf("Loaded Event: %v", event)
				t.Logf("Participant List: %v", *event.Participants)
				for _, p := range event.Participants.participants {
					t.Log(p)
				}
				t.Logf("WrittenEvent: %v", writtenEvent)
				t.Logf("Participant List: %v", *writtenEvent.Participants)
				for _, p := range writtenEvent.Participants.participants {
					t.Log(p)
				}
				t.FailNow()
			}
		}
	}

}
