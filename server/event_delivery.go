package main

import (
	proto "areyouin/protocol"
)

type DeliverMessage struct {
	Event *proto.Event
	Dst   []uint64
}

func NewDeliverySystem() *DeliverySystem {
	ds := &DeliverySystem{}
	ds.queue = make(chan *DeliverMessage)
	ds.udb = udb           // from server.go global scope
	ds.edb = edb           // from server.go global scope
	ds.inbox = users_inbox // from server.go global scope
	return ds
}

type DeliverySystem struct {
	queue chan *DeliverMessage
	udb   *UsersDatabase
	edb   *EventsDatabase
	inbox map[uint64]*Inbox
}

func (ds *DeliverySystem) Submit(event *proto.Event, dst []uint64) {
	ds.queue <- &DeliverMessage{Event: event, Dst: dst}
}

func (ds *DeliverySystem) Run() {
	go func() {

		for {
			m := <-ds.queue
			event := m.Event
			dsts := m.Dst

			// Add event to the author events queue
			if ds.udb.ExistID(event.AuthorId) {
				ds.inbox[event.AuthorId].Add(event)
			}

			for _, user_id := range dsts {
				// Add event to the user events queue
				if ds.udb.ExistID(user_id) {
					ds.inbox[user_id].Add(event)
					// Send invitation to user
					notifyUser(user_id,
						proto.NewMessage().InvitationReceived(event).Marshal())
				}
			}

			// Send Event Created notification to the author
			notifyUser(event.AuthorId,
				proto.NewMessage().EventCreated(event).Marshal())
		}
	}()
}
