package main

import (
  core "peeple/areyouin/common"
  proto "peeple/areyouin/protocol"
  "log"
)

type NotifyParticipantChange struct {
	Event               *core.Event
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

	// Build list with participants that have changed
	participant_list := make([]*core.EventParticipant, 0, len(t.ParticipantsChanged))

	for _, id := range t.ParticipantsChanged {
		participant_list = append(participant_list, t.Event.Participants[id])
	}

	// Send message to each participant
	server := ex.server

	for _, participant_dst := range t.Target {

		session := server.GetSession(participant_dst)

		if session != nil {
			privacy_participant_list := server.filterParticipantsSlice(participant_dst, participant_list)

			var msg *proto.AyiPacket

			if t.NumGuests > 0 {
				msg = session.NewMessage().AttendanceStatusWithNumGuests(t.Event.EventId, privacy_participant_list, t.NumGuests)
			} else {
				msg = session.NewMessage().AttendanceStatus(t.Event.EventId, privacy_participant_list)
			}

			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED (%v participants changed)\n", participant_dst, t.Event.EventId, len(privacy_participant_list))
			} else {
				log.Println("NotifyParticipantChange: Coudn't send notification to", participant_dst)
			}
		}
	}
}
