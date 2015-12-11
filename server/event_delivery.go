package main

import (
	proto "areyouin/protocol"
	"log"
)

func NewDeliverySystem(server *Server) *DeliverySystem {
	ds := &DeliverySystem{}
	ds.queue = make(chan *proto.Event, 100) // Buffered channel
	ds.server = server
	return ds
}

type DeliverySystem struct {
	queue  chan *proto.Event
	server *Server
}

func (ds *DeliverySystem) Submit(event *proto.Event) {
	ds.queue <- event
}

func (ds *DeliverySystem) Run() {
	go func() {

		for {
			// Get pending event from queue
			event := <-ds.queue
			log.Println("Event", event.EventId, "has author", event.AuthorId, "and", event.NumberParticipants, "participants")

			// Add event to participants inboxes (author is also a participant)
			for _, participant := range event.Participants {

				puac, ok := ds.server.udb.GetByID(participant.UserId)

				if !ok {
					log.Println("Coudn't deliver event", event.EventId, "to someone because useraccount does not exist")
					continue
				}

				puac.inbox.Add(event) // FIXME: Clients coud read his/her events while still adding events to inbox
				participant.Delivered = proto.MessageStatus_SERVER_DELIVERED
				log.Println("Event", event.EventId, "delivered to user", participant.UserId)
			}

			// Notify participants
			event_copy := &proto.Event{}
			*event_copy = *event

			for _, participant := range event.Participants {
				// Send Event Created notification to the author
				if event.AuthorId == participant.UserId {
					ds.server.notifyUser(event.AuthorId,
						proto.NewMessage().EventCreated(event).Marshal(), func() {
							participant.Delivered = proto.MessageStatus_CLIENT_DELIVERED
						})
				} else { // Send invitation to user
					if puac, ok := ds.server.udb.GetByID(participant.UserId); ok {
						filterEventParticipants(puac, event_copy, event)
						notificationMsg := proto.NewMessage().InvitationReceived(event_copy).Marshal()
						ds.server.notifyUser(participant.UserId, notificationMsg, func() {
							participant.Delivered = proto.MessageStatus_CLIENT_DELIVERED
						})
					}
				}
			}

		} // For loop
	}() // Go func
}
