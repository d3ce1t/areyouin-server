package main

import (
	proto "areyouin/protocol"
	"log"
)

func NewEventDatabase() *EventsDatabase {
	edb := &EventsDatabase{}
	edb.allevents = make(map[uint64]*proto.Event)
	//edb.userevents = make(map[uint64][]*Event)
	return edb
}

type EventsDatabase struct {
	allevents map[uint64]*proto.Event // Index by ID
	//userevents map[uint64][]*Event // Index events by user ID
}

func (edb *EventsDatabase) ExistEvent(id uint64) bool {
	_, ok := edb.allevents[id]
	return ok
}

func (edb *EventsDatabase) Get(id uint64) (e *proto.Event, ok bool) {
	e, ok = edb.allevents[id]
	return
}

func (edb *EventsDatabase) Insert(event *proto.Event) bool {

	if edb.ExistEvent(event.EventId) {
		log.Println("Given event (", event.EventId, ") already exists")
		return false
	}

	edb.allevents[event.EventId] = event

	return true
}

func (edb *EventsDatabase) Remove(id uint64) {
	if edb.ExistEvent(id) {
		delete(edb.allevents, id)
	}
}
