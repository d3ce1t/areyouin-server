package main

import (
	core "peeple/areyouin/common"
	proto "peeple/areyouin/protocol"
)

type NotifyParticipantChange struct {
	EventId  uint64 // Event in which the participant info has changed
	UserId   uint64 // User ID
	Name     string
	Response core.AttendanceResponse
	Status   core.MessageStatus
}

func (task *NotifyParticipantChange) Run(ex *TaskExecutor) {

	server := ex.server
	event_dao := server.NewEventDAO()
	participants := event_dao.LoadAllParticipants(task.EventId)

	if participants == nil {
		return
	}

	participant_list := make([]*core.EventParticipant, 1)
	participant_list[0] = &core.EventParticipant{
		UserId:    task.UserId,
		Name:      task.Name,
		Response:  task.Response,
		Delivered: task.Status,
	}

	msg := proto.NewMessage().AttendanceStatus(task.EventId, participant_list).Marshal()

	for _, p := range participants {
		if server.canSee(p.UserId, participant_list[0]) {
			server.notifyUser(p.UserId, msg, nil)
		}
	}
}
