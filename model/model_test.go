package model

import (
	"flag"
	"log"
	"os"
	"peeple/areyouin/cqldao"
	"testing"
	"time"
)

var testModel *AyiModel
var users []*UserAccount

func TestMain(m *testing.M) {

	// Create session
	session := cqldao.NewSession("areyouin_test", 4, "192.168.1.10")
	if err := session.Connect(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(-1)
	}

	// Create model and init server
	testModel = New(session, "default")
	testModel.StartBackgroundTasks()

	if err := createFakeUsers(testModel); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(-1)
	}

	flag.Parse()
	os.Exit(m.Run())
}

func createFakeUsers(testModel *AyiModel) error {

	if err := testModel.Accounts.userDAO.DeleteAll(); err != nil {
		return err
	}

	var tests = []struct {
		name     string
		email    string
		password string
	}{
		{"Test1", "test1@example.com", "12345"},
		{"Test2", "test2@example.com", "12345"},
		{"Test3", "test3@example.com", "12345"},
		{"Test4", "test4@example.com", "12345"},
	}

	for _, t := range tests {
		user, err := testModel.Accounts.CreateUserAccount(t.name, t.email, t.password, "", "", "")
		if err != nil {
			return err
		}

		users = append(users, user)
	}

	if err := testModel.Friends.MakeFriends(users[1], users[2]); err != nil {
		return err
	}

	if err := testModel.Friends.MakeFriends(users[1], users[3]); err != nil {
		return err
	}

	return nil
}

func generateEvents(numEvents int, testModel *AyiModel) ([]*Event, error) {

	events := make([]*Event, 0, numEvents)

	for i := 0; i < numEvents; i++ {
		userIndex := i % len(users)
		event, err := generateEvent(users[userIndex], testModel)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func generateEvent(author *UserAccount, testModel *AyiModel) (*Event, error) {

	createdDate := time.Now().UTC()
	startDate := createdDate.Add(30 * time.Minute)
	endDate := createdDate.Add(1 * time.Hour)

	event, err := testModel.Events.NewEvent(author, createdDate, startDate, endDate, "123456789012345", []int64{})
	if err != nil {
		return nil, err
	}

	return event, nil
}
