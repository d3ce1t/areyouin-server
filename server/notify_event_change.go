package main

import (
  core "peeple/areyouin/common"
  "log"
)

type NotifyEventChange struct {
	Event  *core.Event
	Target []int64
}

func (t *NotifyEventChange) Run(ex *TaskExecutor) {

	if len(t.Target) == 0 {
		log.Println("NotifyEventChange. Task doesn't have any target")
		return
	}

	// Send message to each participant
	server := ex.server
	light_event := t.Event.GetEventWithoutParticipants()

	for _, participant_dst := range t.Target {

		session := server.GetSession(participant_dst)

		if session != nil {

			msg := session.NewMessage().EventModified(light_event)

			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED\n", participant_dst, t.Event.EventId)
			} else {
				log.Println("NotifyEventChange: Coudn't send notification to", participant_dst)
			}
		}
	}
}
