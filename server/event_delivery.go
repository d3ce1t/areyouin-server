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
			log.Println("Event", event.EventId, "has author", event.AuthorId, "and", event.NumberParticipants, "participants")

			// Add event to participants inboxes (author is also a participant)
			for _, participant := range event.Participants {

				puac, ok := ds.udb.GetByID(participant.UserId)

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
					notifyUser(event.AuthorId,
						proto.NewMessage().EventCreated(event).Marshal(), func() {
							participant.Delivered = proto.MessageStatus_CLIENT_DELIVERED
						})
				} else { // Send invitation to user
					if puac, ok := ds.udb.GetByID(participant.UserId); ok {
						//event_copy.Participants = make([]*proto.EventParticipant, 2)
						//copy(event_copy.Participants, GetConfirmedParticipants(puac, event.Participants))
						event_copy.Participants = GetConfirmedParticipants(puac, event.Participants)
						notificationMsg := proto.NewMessage().InvitationReceived(event_copy).Marshal()
						notifyUser(participant.UserId, notificationMsg, func() {
							participant.Delivered = proto.MessageStatus_CLIENT_DELIVERED
						})
					}
				}

			}

		} // For loop
	}() // Go func
}

func GetConfirmedParticipants(user_account *UserAccount, participants []*proto.EventParticipant) []*proto.EventParticipant {

	result := make([]*proto.EventParticipant, 0, 10) // FIXME: Make constant

	for _, p := range participants {
		// If the participant is a confirmed user (yes or cannot assist answer has been given)
		if p.Response == proto.AttendanceResponse_ASSIST ||
			p.Response == proto.AttendanceResponse_CANNOT_ASSIST ||
			user_account.IsFriend(p.UserId) ||
			user_account.id == p.UserId { // self-user
			result = append(result, p)
		}
	}

	return result
}
