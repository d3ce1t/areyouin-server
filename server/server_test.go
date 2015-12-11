package main

import (
	"flag"
	//"fmt"
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
	//user2, _ := udb.GetByEmail("user2@foo.com")
	//events := user2.GetAllEvents()
	//fmt.Println(events)
	//filterEventParticipants(user2, dst_event*proto.Event, src_event*proto.Event)
}
