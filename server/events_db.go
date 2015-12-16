package main

import (
	proto "areyouin/protocol"
	"log"
)

func NewEventDatabase() *EventsDatabase {
	edb := &EventsDatabase{}
	edb.allevents = make(map[uint64]*proto.Event)
	edb.allparticipants = make(map[uint64]map[uint64]*proto.EventParticipant)
	return edb
}

type EventsDatabase struct {
	allevents       map[uint64]*proto.Event                       // Index by ID
	allparticipants map[uint64]map[uint64]*proto.EventParticipant // Index by EventID and User ID
}

func (edb *EventsDatabase) ExistEvent(id uint64) bool {
	_, ok := edb.allevents[id]
	return ok
}

func (edb *EventsDatabase) GetEvent(id uint64) (e *proto.Event, ok bool) {
	e, ok = edb.allevents[id]
	return
}

func (edb *EventsDatabase) InsertEvent(event *proto.Event) bool {

	if edb.ExistEvent(event.EventId) {
		log.Println("Given event (", event.EventId, ") already exists")
		return false
	}

	edb.allevents[event.EventId] = event
	edb.allparticipants[event.EventId] = make(map[uint64]*proto.EventParticipant)
	return true
}

func (edb *EventsDatabase) RemoveEvent(id uint64) {
	if edb.ExistEvent(id) {
		delete(edb.allevents, id)
		delete(edb.allparticipants, id)
	}
}

// Assume participant belongs to a registered user
func (edb *EventsDatabase) AddParticipant(event_id uint64, participant *proto.EventParticipant) bool {

	event_participants, exist := edb.allparticipants[event_id]

	if !exist {
		log.Println("addParticipant() attempt to add a participant to a non existent event", event_id)
		return false
	}

	event_participants[participant.UserId] = participant
	event, _ := edb.allevents[event_id]
	event.NumGuests = int32(len(event_participants))

	return true
}

func (edb *EventsDatabase) GetAllParticipants(event_id uint64) []*proto.EventParticipant {

	result := make([]*proto.EventParticipant, 0, len(edb.allparticipants[event_id]))

	for _, v := range edb.allparticipants[event_id] {
		result = append(result, v)
	}

	return result
}

func (edb *EventsDatabase) GetParticipantsMap(event_id uint64) (v map[uint64]*proto.EventParticipant, ok bool) {
	v, ok = edb.allparticipants[event_id]
	return
}
