package main

import (
	"log"
	"peeple/areyouin/model"
	proto "peeple/areyouin/protocol"
)

type NotifyParticipantChange struct {
	Event               *model.Event
	ParticipantsChanged []int64 // Participants that has changed
	NumGuests           int32
	Target              []int64
}

func (t *NotifyParticipantChange) Run(ex *TaskExecutor) {

	if len(t.Target) == 0 {
		log.Println("NotifyParticipantChange. Task doesn't have any target")
		return
	}

	if len(t.ParticipantsChanged) == 0 {
		log.Println("NotifyParticipantChange. Task doesn't have changes to notify")
		return
	}

	server := ex.server

	// Build list with participants that have changed

	participantList := make([]*model.Participant, 0, len(t.ParticipantsChanged))
	for _, id := range t.ParticipantsChanged {
		participantList = append(participantList, t.Event.GetParticipant(id))
	}

	for _, target := range t.Target {

		// Send message to each participant

		session := server.GetSession(target)

		if session != nil {

			filteredParticipants := server.Model.Events.FilterParticipantsSlice(participantList, target)
			netFilteredParticipants := convParticipantList2Net(filteredParticipants)

			var msg *proto.AyiPacket

			// TODO: Why am I doing this?
			if t.NumGuests > 0 {
				msg = session.NewMessage().AttendanceStatusWithNumGuests(t.Event.Id(), netFilteredParticipants, t.NumGuests)
			} else {
				msg = session.NewMessage().AttendanceStatus(t.Event.Id(), netFilteredParticipants)
			}

			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED (%v participants changed)\n", target, t.Event.Id(), len(netFilteredParticipants))
			} else {
				log.Println("NotifyParticipantChange: Coudn't send notification to", target)
			}
		}
	}
}
