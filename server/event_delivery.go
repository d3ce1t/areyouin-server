package main

import (
	proto "areyouin/protocol"
	"log"
)

func NewDeliverySystem() *DeliverySystem {
	ds := &DeliverySystem{}
	ds.queue = make(chan *proto.Event, 100) // Buffered channel
	ds.udb = udb                            // from server.go global scope
	ds.edb = edb                            // from server.go global scope
	return ds
}

type DeliverySystem struct {
	queue chan *proto.Event
	udb   *UsersDatabase
	edb   *EventsDatabase
}

func (ds *DeliverySystem) Submit(event *proto.Event) {
	ds.queue <- event
}

func (ds *DeliverySystem) Run() {
	go func() {

		for {
			// Get pending event from queue
			event := <-ds.queue

			// Add event to participants inboxes (author is also a participant)
			notificationMsg := proto.NewMessage().InvitationReceived(event).Marshal()

			for _, pstatus := range event.Participants {

				puac, ok := ds.udb.GetByID(pstatus.UserId)

				if !ok {
					log.Println("Coudn't deliver event", event.EventId, "to someone because useraccount does not exist")
					continue
				}

				puac.inbox.Add(event)
				pstatus.Delivered = proto.MessageStatus_SERVER_DELIVERED

				if event.AuthorId == pstatus.UserId {
					log.Println("Event", event.EventId, "delivered to author", puac.id)
					// Send Event Created notification to the author
					notifyUser(event.AuthorId,
						proto.NewMessage().EventCreated(event).Marshal(), func() {
							pstatus.Delivered = proto.MessageStatus_CLIENT_DELIVERED
						})
				} else {
					log.Println("Event", event.EventId, "delivered to user", puac.id)
					// Send invitation to user
					notifyUser(pstatus.UserId, notificationMsg, func() {
						pstatus.Delivered = proto.MessageStatus_CLIENT_DELIVERED
					})
				}
			}
		}
	}()
}
