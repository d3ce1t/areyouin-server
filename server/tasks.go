package main

import (
	"log"
	core "peeple/areyouin/common"
	dao "peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
	proto "peeple/areyouin/protocol"
)

type NotifyParticipantChange struct {
	Event               *core.Event
	ParticipantsChanged []uint64 // Participants that has changed
}

func (t *NotifyParticipantChange) Run(ex *TaskExecutor) {

	if len(t.Event.Participants) == 0 {
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

	for _, participant_dst := range t.Event.Participants {

		session := server.GetSession(participant_dst.UserId)

		if session != nil {
			privacy_participant_list := server.filterParticipantsSlice(participant_dst.UserId, participant_list)
			msg := proto.NewMessage().AttendanceStatus(t.Event.EventId, privacy_participant_list).Marshal()

			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED (%v participants changed)\n", participant_dst.UserId, t.Event.EventId, len(privacy_participant_list))
			} else {
				log.Println("NotifyParticipantChange: Coudn't send notificatino to", participant_dst.UserId)
			}
		}
	}
}

type ImportFacebookFriends struct {
	UserId  uint64
	Name    string
	Fbtoken string // Facebook User Access token
}

func (task *ImportFacebookFriends) Run(ex *TaskExecutor) {

	server := ex.server
	fbsession := fb.NewSession(task.Fbtoken)
	fbFriends, err := fb.GetFriends(fbsession)

	if err != nil {
		fb.LogError(err)
		return
	}

	user_dao := server.NewUserDAO()
	storedFriends, err := user_dao.LoadFriendsIndex(task.UserId, ALL_CONTACTS_GROUP)

	if err != nil {
		log.Println("ImportFacebookFriends Error:", err)
		return
	}

	log.Printf("ImportFacebookFriends: %v friends found\n", len(fbFriends))

	// Create a friend object with the info from the user that initiated the import
	current_user := &core.Friend{
		UserId: task.UserId,
		Name:   task.Name,
	}

	counter := 0

	for _, fbFriend := range fbFriends {

		friend_id, err := user_dao.GetIDByFacebookID(fbFriend.Id)

		if err != nil {
			if err == dao.ErrNotFound {
				log.Printf("ImportFacebookFriends Error: Facebook friend %v has the App but it's not registered\n", fbFriend.Id)
			} else {
				log.Println("ImportFacebookFriends Error:", err)
			}
			continue
		}

		log.Printf("ImportFacebookFriends: %v and %v are Facebook Friends\n", current_user.GetUserId(), friend_id)

		// Assume that if friend_id isn't in stored friends, then current user id isn't either
		// in the other user friends list
		if _, ok := storedFriends[friend_id]; !ok {
			user_dao.MakeFriends(current_user, &core.Friend{UserId: friend_id, Name: fbFriend.Name})
			ex.Submit(&SendUserFriends{UserId: friend_id})
			counter++
		}
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
		log.Println("SendUserFriends Error:", err)
		return
	}

	if len(friends) > 0 {
		reply := proto.NewMessage().FriendsList(friends).Marshal()
		if server.SendMessage(task.UserId, reply) {
			log.Printf("< (%v) SEND USER FRIENDS\n", task.UserId)
		}
	}
}

// Task to notify guests/participants of an event that have been invited. This
// task is used whenever a new event is created and participants have not been
// notified yet.
type NotifyEventInvitation struct {
	Event              *core.Event
	TargetParticipants map[uint64]*core.EventParticipant
}

func (t *NotifyEventInvitation) Run(ex *TaskExecutor) {

	server := ex.server
	light_event := t.Event.GetEventWithoutParticipants()
	futures := make(map[uint64]chan bool)
	participants_changed := make([]uint64, 0, len(t.TargetParticipants))

	if len(t.TargetParticipants) == 0 {
		log.Println("NotifyEventInvitation: There aren't targetted participants to send notification")
		return
	}

	// Send event and its attendance status to all of the target participants
	for _, participant := range t.TargetParticipants {

		// Add participant to participants changed because if NotifyEventInvitation is being
		// called, it is because the participant didn't exist before in the participant list,
		// so it can be considered that this participant has changed from the perspective
		// of the existing participants
		participants_changed = append(participants_changed, participant.UserId)

		// Notify participant about the invitation only if it's connected.
		session := server.GetSession(participant.UserId)

		if session == nil {
			continue
		}

		var msg []byte

		// Create message of event creation or InvitationReceived
		if t.Event.AuthorId == participant.UserId {
			msg = proto.NewMessage().EventCreated(light_event).Marshal()
		} else { // Send invitation to user
			msg = proto.NewMessage().InvitationReceived(light_event).Marshal()
		}

		// Filter event participants to protect privacy
		filtered_participants := server.filterParticipantsMap(participant.UserId, t.Event.Participants)

		// Append attendance status msg
		attendanceStatus := proto.NewMessage().AttendanceStatus(t.Event.EventId, filtered_participants).Marshal()
		msg = append(msg, attendanceStatus...)

		// Notify
		c := make(chan bool)
		if ok := session.WriteAsync(msg, c); ok {
			futures[participant.UserId] = c
		}
	}

	// Update invitation delivery status
	eventDAO := ex.server.NewEventDAO()

	for participant_id, c := range futures {
		ok := <-c
		if ok {
			log.Printf("< (%v) SEND NEW EVENT %v\n", participant_id, t.Event.EventId)
			err := eventDAO.SetParticipantStatus(participant_id, t.Event.EventId, core.MessageStatus_CLIENT_DELIVERED) // participant changed
			if err == nil {
				t.Event.Participants[participant_id].Delivered = core.MessageStatus_CLIENT_DELIVERED
			} else {
				log.Println("NotifyEventInvitation Err:", err)
			}
		}
	}

	// Notify changes to the rest of participants
	task := &NotifyParticipantChange{
		Event:               t.Event,
		ParticipantsChanged: participants_changed,
	}

	task.Run(ex)
}
