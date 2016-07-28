package main

import (
	"log"
	"peeple/areyouin/model"
)

type NotifyEventChange struct {
	Event  *model.Event
	Target []int64
}

func (t *NotifyEventChange) Run(ex *TaskExecutor) {

	if len(t.Target) == 0 {
		log.Println("NotifyEventChange. Task doesn't have any target")
		return
	}

	// Send message to each participant
	server := ex.server
	light_event := convEvent2Net(t.Event.CloneEmptyParticipants())

	for _, participant_dst := range t.Target {

		session := server.getSession(participant_dst)

		if session != nil {

			msg := session.NewMessage().EventModified(light_event)

			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED\n", participant_dst, t.Event.Id())
			} else {
				log.Println("NotifyEventChange: Coudn't send notification to", participant_dst)
			}
		}
	}
}
