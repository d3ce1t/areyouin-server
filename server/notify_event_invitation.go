package main

import (
  core "peeple/areyouin/common"
  "peeple/areyouin/dao"
  "log"
)

// Task to notify guests/participants of an event that they have been invited. This
// task is used whenever a new event is created and participants have not been
// notified yet.
type NotifyEventInvitation struct {
	Event  *core.Event
	Target map[int64]*core.UserAccount // Users that will be invited to the event
  futures map[int64]chan bool
}

func (t *NotifyEventInvitation) Run(ex *TaskExecutor) {

	server := ex.server
	light_event := t.Event.GetEventWithoutParticipants()
	t.futures = make(map[int64]chan bool)

	if len(t.Target) == 0 {
		log.Println("NotifyEventInvitation: There aren't targetted participants to send notification")
		return
	}

	// Send event and its attendance status to all of the target participants
	for _, user := range t.Target {
    switch user.NetworkVersion {
    case 0, 1:
      if ok := t.notifyUserV1(user, light_event, server); !ok {
        continue
      }
    case 2:
        if ok := t.notifyUserV2(user, t.Event, server); !ok {
          continue
        }
    }
	} // End loop

	// Update invitation delivery status
	participants_changed := make([]int64, 0, len(t.Target))
	eventDAO := dao.NewEventDAO(ex.server.DbSession)

	for participant_id, c := range t.futures {

		ok := <-c // Blocks until ACK (true) or timeout (false) // TODO: Code may block here for 10 seconds. CHANGE IT!

		if ok {

      // ACK Received

			err := eventDAO.SetParticipantStatus(participant_id, t.Event.EventId, core.MessageStatus_CLIENT_DELIVERED) // participant changed
			if err == nil {
				t.Event.Participants[participant_id].Delivered = core.MessageStatus_CLIENT_DELIVERED
				// Add participant to changed set because delivery status has changed
				participants_changed = append(participants_changed, participant_id)
			} else {
				log.Println("NotifyEventInvitation Err:", err)
			}

		} else {

      // timeout or error

      user := t.Target[participant_id]
      if user.NetworkVersion >= 2 {
        ttl := uint32(t.Event.StartDate - core.GetCurrentTimeMillis()) / 1000
        sendGcmDataAvailableNotification(user.Id, user.IIDtoken, ttl)
      } else {
			  sendGcmNewEventNotification(user.Id, user.IIDtoken, t.Event)
      }
		}
	}

	// Notify changes to the rest of participants
	if len(participants_changed) > 0 {
		task := &NotifyParticipantChange{
			Event:               t.Event,
			ParticipantsChanged: participants_changed,
			Target:              core.GetParticipantKeys(t.Event.Participants),
		}

		task.Run(ex)
	}
}

func (t *NotifyEventInvitation) notifyUserV1(user *core.UserAccount, event *core.Event, server *Server) bool {

  // Notify participant about the invitation only if it's connected.
  session := server.GetSession(user.Id)

  if session == nil {
    sendGcmNewEventNotification(user.Id, user.IIDtoken, t.Event)
    return false
  }

  // Create InvitationReceived message
  // Keep this code for clients that uses v0 and v1

  // Filter event participants to protect privacy and create message
  filtered_participants := server.filterParticipantsMap(user.Id, t.Event.Participants)
  notify_msg := session.NewMessage().InvitationReceived(event)
  attendance_status_msg := session.NewMessage().AttendanceStatus(t.Event.EventId, filtered_participants)

  // Notify (use a channel because it is needed to know if message arrived)
  var future *Future

  if user.IIDtoken != "" {
    future = NewFuture(true)
  } else {
    future = NewFuture(false)
  }

  if ok := session.WriteAsync(future, notify_msg); ok {
    session.Write(attendance_status_msg)
    t.futures[user.Id] = future.C
    log.Printf("< (%v) SEND EVENT INVITATION (event_id=%v)\n", user.Id, t.Event.EventId)
  } else {
    sendGcmNewEventNotification(user.Id, user.IIDtoken, t.Event)
  }

  return true
}

func (t *NotifyEventInvitation) notifyUserV2(user *core.UserAccount, event *core.Event, server *Server) bool {

  // Notify participant about the invitation only if it's connected.
  session := server.GetSession(user.Id)

  if session == nil {
    ttl := uint32(event.StartDate - core.GetCurrentTimeMillis()) / 1000
    sendGcmDataAvailableNotification(user.Id, user.IIDtoken, ttl)
    return false
  }

  // From protocol v2 onward, invitation received message contains event info
  // and participants.

  eventCopy := &core.Event{}
  *eventCopy = *t.Event

  // Filter event participants to protect privacy and create message
  eventCopy.Participants = server.filterEventParticipants(user.Id, t.Event.Participants)
  notify_msg := session.NewMessage().InvitationReceived(eventCopy)

  // Notify (use a channel because it is needed to know if message arrived)
  var future *Future

  if user.IIDtoken != "" {
    future = NewFuture(true)
  } else {
    future = NewFuture(false)
  }

  if ok := session.WriteAsync(future, notify_msg); ok {
    t.futures[user.Id] = future.C // May block here upto 10 seconds
    log.Printf("< (%v) SEND EVENT INVITATION (event_id=%v)\n", user.Id, t.Event.EventId)
  } else {
    // Fallback
    ttl := uint32(event.StartDate - core.GetCurrentTimeMillis()) / 1000
    sendGcmDataAvailableNotification(user.Id, user.IIDtoken, ttl)
  }

  return true
}
