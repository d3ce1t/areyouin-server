package main

import (
	proto "areyouin/protocol"
	"log"
)

type DeliverMessage struct {
	Event *proto.Event
	Dst   []uint64
}

func NewDeliverySystem() *DeliverySystem {
	ds := &DeliverySystem{}
	ds.queue = make(chan *DeliverMessage, 100) // Buffered channel
	ds.udb = udb                               // from server.go global scope
	ds.edb = edb                               // from server.go global scope
	return ds
}

type DeliverySystem struct {
	queue chan *DeliverMessage
	udb   *UsersDatabase
	edb   *EventsDatabase
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

			// Add event to the author inbox
			if uac, ok := ds.udb.GetByID(event.AuthorId); ok {
				uac.inbox.Add(event)
				log.Println("Event", event.EventId, "delivered to author", uac.id)
			}

			for _, user_id := range dsts {
				// Add event to the user events queue
				if uac, ok := ds.udb.GetByID(user_id); ok {
					uac.inbox.Add(event)
					log.Println("Event", event.EventId, "delivered to user", uac.id)
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
