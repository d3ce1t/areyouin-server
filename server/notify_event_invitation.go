package main

import (
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"
)

// Task to notify guests/participants of an event that they have been invited. This
// task is used whenever a new event is created and participants have not been
// notified yet.
type NotifyEventInvitation struct {
	Event   *model.Event
	Target  map[int64]*model.UserAccount // Users that will be invited to the event
	futures map[int64]chan bool
}

func (t *NotifyEventInvitation) Run(ex *TaskExecutor) {

	server := ex.server
	t.futures = make(map[int64]chan bool)

	if len(t.Target) == 0 {
		log.Println("NotifyEventInvitation: There aren't targetted participants to send notification")
		return
	}

	// Send event and its attendance status to all of the target participants
	for _, user := range t.Target {
		if ok := t.notifyUser(user, t.Event, server); !ok {
			continue
		}

	} // End loop

	// Update invitation delivery status
	participants_changed := make([]int64, 0, len(t.Target))

	for participant_id, c := range t.futures {

		ok := <-c // Blocks until ACK (true) or timeout (false) // TODO: Code may block here for 10 seconds. CHANGE IT!

		if ok {

			// ACK Received

			applied, err := server.Model.Events.ChangeDeliveryState(t.Event, participant_id, api.InvitationStatus_CLIENT_DELIVERED)
			if err != nil {
				log.Println("NotifyEventInvitation Err:", err)
			}

			if applied {
				// Add participant to changed set because delivery status has changed
				participants_changed = append(participants_changed, participant_id)
			}

		} else {

			// timeout or error

			user := t.Target[participant_id]
			token := user.PushToken()
			ttl := uint32(t.Event.StartDate()-utils.GetCurrentTimeMillis()) / 1000
			sendGcmDataAvailableNotification(user.Id(), &token, ttl)
		}
	}

	// Notify changes to the rest of participants
	if len(participants_changed) > 0 {
		task := &NotifyParticipantChange{
			Event:               t.Event,
			ParticipantsChanged: participants_changed,
			Target:              t.Event.ParticipantIds(),
		}

		task.Run(ex)
	}
}

func (t *NotifyEventInvitation) notifyUser(user *model.UserAccount, event *model.Event, server *Server) bool {

	// Notify participant about the invitation only if it's connected.
	session := server.getSession(user.Id())
	token := user.PushToken()

	if session == nil {
		ttl := uint32(event.StartDate()-utils.GetCurrentTimeMillis()) / 1000
		sendGcmDataAvailableNotification(user.Id(), &token, ttl)
		return false
	}

	// From protocol v2 onward, invitation received message contains event info
	// and participants.

	// Copy event with participants filtered
	filteredEvent := server.Model.Events.GetFilteredEvent(event, user.Id())

	// Notify (use a channel because it is needed to know if message arrived)
	notify_msg := session.NewMessage().InvitationReceived(convEvent2Net(filteredEvent))
	var future *Future

	if token.Token() != "" {
		future = NewFuture(true)
	} else {
		future = NewFuture(false)
	}

	if ok := session.WriteAsync(future, notify_msg); ok {
		t.futures[user.Id()] = future.C // May block here upto 10 seconds
		log.Printf("< (%v) SEND EVENT INVITATION (event_id=%v)\n", user.Id(), t.Event.Id())
	} else {
		// Fallback
		ttl := uint32(event.StartDate()-utils.GetCurrentTimeMillis()) / 1000
		sendGcmDataAvailableNotification(user.Id(), &token, ttl)
	}

	return true
}
