package main

import (
	"log"
	core "peeple/areyouin/common"
	proto "peeple/areyouin/protocol"
)

const SIZE_QUEUE = 100

func NewDeliverySystem(server *Server) *DeliverySystem {
	ds := &DeliverySystem{}
	ds.queue = make(chan *core.Event, SIZE_QUEUE) // Buffered channel
	ds.server = server
	return ds
}

type DeliverySystem struct {
	queue  chan *core.Event
	server *Server
}

// TODO: DeliverySystem Submit must be persistent in order to continue the job
// in case of failure
func (ds *DeliverySystem) Submit(event *core.Event) {
	ds.queue <- event
}

/*
 Puts an event to participants' inbox, updates event delivery information and notify
 participants
*/
func (ds *DeliverySystem) Run() {
	go func() {

		dao := ds.server.NewEventDAO()

		for {
			// Get pending event from queue
			event := <-ds.queue
			log.Println("Event", event.EventId, "has author", event.AuthorId, "and", event.NumGuests, "guests")

			// Dispatch event to each participant
			event_participants := dao.LoadAllParticipants(event.EventId)

			if len(event_participants) > 0 {

				for _, participant := range event_participants {
					if err := ds.dispatchEvent(event, participant); err == nil {
						log.Println("Event", event.EventId, "delivered to user", participant.UserId)
					} else {
						log.Println("Coudn't deliver event", event.EventId, err)
					}
				}

				// Send notifications
				for _, participant := range event_participants {
					ds.sendNotifications(event, event_participants, participant)
				}
			} else {
				log.Println("DeliverySystem: Event", event.EventId, "has no participants (nothing to do)")
			}
		}
	}() // Go func
}

func (ds *DeliverySystem) dispatchEvent(event *core.Event, participant *core.EventParticipant) error {

	// FIXME: These two writes should be a batch
	// Add event to participant inbox (author is also a participant)
	dao := ds.server.NewEventDAO()
	if err := dao.AddEventToUserInbox(participant.UserId, event, participant.Response); err != nil {
		return err
	}

	participant.Delivered = core.MessageStatus_SERVER_DELIVERED
	if err := dao.SetParticipantStatus(participant.UserId, event.EventId, participant.Delivered); err != nil {
		return err
	}

	return nil
}

func (ds *DeliverySystem) sendNotifications(event *core.Event, event_participants []*core.EventParticipant,
	participant *core.EventParticipant) {

	var msg []byte

	// Create message of event creation or InvitationReceived
	if event.AuthorId == participant.UserId {
		msg = proto.NewMessage().EventCreated(event).Marshal()
	} else { // Send invitation to user
		msg = proto.NewMessage().InvitationReceived(event).Marshal()
	}

	// Filter event participants to protect privacy
	event_participants = ds.server.filterParticipants(participant.UserId, event_participants)

	// Append attendance status msg
	attendanceStatus := proto.NewMessage().AttendanceStatus(event.EventId, event_participants).Marshal()
	msg = append(msg, attendanceStatus...)

	// Notify
	ds.server.notifyUser(participant.UserId, msg, ds.callback(event, participant))
}

// FIXME: Callback called from handleSession goroutine
func (ds *DeliverySystem) callback(event *core.Event, participant *core.EventParticipant) func() {
	e := event
	p := participant
	return func() {
		dao := ds.server.NewEventDAO()
		p.Delivered = core.MessageStatus_CLIENT_DELIVERED
		if err := dao.SetParticipantStatus(p.UserId, event.EventId, p.Delivered); err != nil {
			log.Println("DeliverySystem:Callback:", err)
		}
		ds.onParticipantChanged(e, p)
	}
}

// Notify all of the event's participants about a participant status changes
func (ds *DeliverySystem) onParticipantChanged(event *core.Event, changed_participant *core.EventParticipant) {

	dao := ds.server.NewEventDAO()

	// Prepare message with only one participant
	participant_list := make([]*core.EventParticipant, 1)
	participant_list[0] = changed_participant
	message := proto.NewMessage().AttendanceStatus(event.EventId, participant_list).Marshal()

	event_participants := dao.LoadAllParticipants(event.EventId)

	// Only notify to those participants that can see the changed_participant
	for _, participant := range event_participants {
		if ds.server.canSee(participant.UserId, changed_participant) {
			ds.server.notifyUser(participant.UserId, message, nil)
		}
	}
}
