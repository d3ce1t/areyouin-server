package main

import (
	"flag"
	"os"
	"testing"
)

//var udb *UsersDatabase
//var edb *EventsDatabase
var server *Server

func TestMain(m *testing.M) {
	server = NewServer()
	initFakeUsers(server)
	initFakeEvents(server)
	flag.Parse()
	os.Exit(m.Run())
}

func TestFilterEventParticipants(t *testing.T) {

	user2, _ := server.udb.GetByEmail("user2@foo.com")
	events := user2.GetAllEvents()

	for _, v := range events {
		if v == nil {
			t.Fatal("Event is nil!")
		}
	}
}
