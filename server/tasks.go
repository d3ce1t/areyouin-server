package main

import (
	"log"
	core "peeple/areyouin/common"
	dao "peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
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

// FIXME: It's only adding so far. It has to sync, i.e. remove friends that are not facebook friends
type SyncFacebookFriends struct {
	UserId  uint64
	Name    string
	Fbtoken string // Facebook User Access token
}

func (task *SyncFacebookFriends) Run(ex *TaskExecutor) {

	server := ex.server
	fbsession := fb.NewSession(task.Fbtoken)
	friends, err := fb.GetFriends(fbsession)

	if err != nil {
		fb.LogError(err)
		return
	}

	user_dao := server.NewUserDAO()
	counter := 0

	for _, friend := range friends {

		friend_id, err := user_dao.GetIDByFacebookID(friend.Id)

		if err != nil {
			if err != dao.ErrNotFound {
				log.Println("SyncFacebookFriends Error:", err)
			} else {
				log.Println("SyncFacebookFriends: Facebook friend has the App but it's not registered")
			}
			continue
		}

		user_dao.MakeFriends(
			&core.Friend{UserId: task.UserId, Name: task.Name},
			&core.Friend{UserId: friend_id, Name: friend.Name},
		)

		ex.Submit(&SendUserFriends{UserId: friend_id})
		counter++
	}

	if counter > 0 {
		ex.Submit(&SendUserFriends{UserId: task.UserId})
	}
}

type SendUserFriends struct {
	UserId uint64
}

func (task *SendUserFriends) Run(ex *TaskExecutor) {

	server := ex.server
	dao := server.NewUserDAO()

	friends, err := dao.LoadFriends(task.UserId, ALL_CONTACTS_GROUP)

	if err != nil {
		log.Println("NotifyFriendChange Error:", err)
		return
	}

	if len(friends) > 0 {
		reply := proto.NewMessage().FriendsList(friends).Marshal()
		server.notifyUser(task.UserId, reply, func() {
			log.Println("SEND USER FRIENDS to", task.UserId)
		})
	}
}
