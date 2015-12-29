package main

import (
	"log"
	core "peeple/areyouin/common"
	proto "peeple/areyouin/protocol"
)

type NotifyParticipantChange struct {
	EventId        uint64 // Event in which the participant info has changed
	UserId         uint64 // User ID
	Name           string
	Response       core.AttendanceResponse
	Status         core.MessageStatus
	participantIds []uint64
}

func (task *NotifyParticipantChange) NumDsts() int {
	return len(task.participantIds)
}

func (task *NotifyParticipantChange) AddDst(userId uint64) {
	task.participantIds = append(task.participantIds, userId)
}

func (task *NotifyParticipantChange) AddParticipantsDst(participants []*core.EventParticipant) {

	if participants == nil {
		log.Println("AddParticipantsDst: Participants are nil")
		return
	}

	for _, p := range participants {
		task.AddDst(p.UserId)
	}
}

func (task *NotifyParticipantChange) Run(ex *TaskExecutor) {

	if len(task.participantIds) == 0 {
		log.Println("NotifyParticipantChange. Task doesn't have any target")
		return
	}

	server := ex.server

	participant_list := make([]*core.EventParticipant, 1)
	participant_list[0] = &core.EventParticipant{
		UserId:    task.UserId,
		Name:      task.Name,
		Response:  task.Response,
		Delivered: task.Status,
	}

	privacy_participant_list := make([]*core.EventParticipant, 1)
	privacy_participant_list[0] = participant_list[0].AsAnonym()

	msg := proto.NewMessage().AttendanceStatus(task.EventId, participant_list).Marshal()
	msg_privacy := proto.NewMessage().AttendanceStatus(task.EventId, privacy_participant_list).Marshal()

	for _, pId := range task.participantIds {
		if server.canSee(pId, participant_list[0]) {
			server.notifyUser(pId, msg, nil)
		} else {
			server.notifyUser(pId, msg_privacy, nil)
		}
	}
}
