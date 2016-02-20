package main

import (
	gcm "github.com/google/go-gcm"
	"log"
	core "peeple/areyouin/common"
	dao "peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
)

type NotifyParticipantChange struct {
	Event               *core.Event
	ParticipantsChanged []uint64 // Participants that has changed
	Target              []uint64
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

	for _, participant_dst := range t.Target {

		session := server.GetSession(participant_dst)

		if session != nil {
			privacy_participant_list := server.filterParticipantsSlice(participant_dst, participant_list)
			msg := session.NewMessage().AttendanceStatus(t.Event.EventId, privacy_participant_list)

			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED (%v participants changed)\n", participant_dst, t.Event.EventId, len(privacy_participant_list))
			} else {
				log.Println("NotifyParticipantChange: Coudn't send notificatino to", participant_dst)
			}
		}
	}
}

type ImportFacebookFriends struct {
	TargetUser core.UserFriend
	Fbtoken    string // Facebook User Access token
}

func (task *ImportFacebookFriends) Run(ex *TaskExecutor) {

	server := ex.server

	fbsession := fb.NewSession(task.Fbtoken)
	fbFriends, err := fb.GetFriends(fbsession)
	if err != nil {
		fb.LogError(err)
		return
	}

	friend_dao := server.NewFriendDAO()
	storedFriends, err := friend_dao.LoadFriendsIndex(task.TargetUser.GetUserId(), ALL_CONTACTS_GROUP)
	if err != nil {
		log.Println("ImportFacebookFriends Error:", err)
		return
	}

	log.Printf("ImportFacebookFriends: %v friends found\n", len(fbFriends))

	user_dao := server.NewUserDAO()
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

		friendUser, err := user_dao.Load(friend_id)
		if err != nil {
			log.Println("ImportFacebookFriends Error:", err)
		}

		log.Printf("ImportFacebookFriends: %v and %v are Facebook Friends\n", task.TargetUser.GetUserId(), friend_id)

		// Assume that if friend_id isn't in stored friends, then current user id isn't either
		// in the other user friends list
		if _, ok := storedFriends[friend_id]; !ok {
			friendUser.Name = fbFriend.Name // Use Facebook name because is familiar to user
			friend_dao.MakeFriends(task.TargetUser, friendUser)
			ex.Submit(&SendUserFriends{UserId: friend_id})
			task.sendGcmNotification(friendUser.Id, friendUser.IIDtoken, task.TargetUser.GetName())
			counter++
		}

	}

	if counter > 0 {
		ex.Submit(&SendUserFriends{UserId: task.TargetUser.GetUserId()})
	}
}

func (t *ImportFacebookFriends) sendGcmNotification(user_id uint64, token string, friend_name string) {

	gcm_message := gcm.HttpMessage{
		To: token,
		Data: gcm.Data{
			"msg_type":    "notification",
			"notify_type": GCM_NEW_FRIEND_MESSAGE,
			"friend_name": friend_name,
		},
	}

	sendGcmMessage(user_id, token, gcm_message)
}

type SendUserFriends struct {
	UserId uint64
}

func (task *SendUserFriends) Run(ex *TaskExecutor) {

	server := ex.server
	friend_dao := server.NewFriendDAO()

	friends, err := friend_dao.LoadFriends(task.UserId, ALL_CONTACTS_GROUP)

	if err != nil {
		log.Println("SendUserFriends Error:", err)
		return
	}

	if len(friends) > 0 {
		session := server.GetSession(task.UserId)
		if session != nil {
			packet := session.NewMessage().FriendsList(friends)
			if session.Write(packet) {
				log.Printf("< (%v) SEND USER FRIENDS\n", task.UserId)
			}
		}
	}
}

// Task to notify guests/participants of an event that they have been invited. This
// task is used whenever a new event is created and participants have not been
// notified yet.
type NotifyEventInvitation struct {
	Event  *core.Event
	Target map[uint64]*core.UserAccount // Users that will be invited to the event
}

func (t *NotifyEventInvitation) Run(ex *TaskExecutor) {

	server := ex.server
	light_event := t.Event.GetEventWithoutParticipants()
	futures := make(map[uint64]chan bool)

	if len(t.Target) == 0 {
		log.Println("NotifyEventInvitation: There aren't targetted participants to send notification")
		return
	}

	// Send event and its attendance status to all of the target participants
	for _, user := range t.Target {

		// Notify participant about the invitation only if it's connected.
		session := server.GetSession(user.Id)

		if session == nil {
			t.sendGcmNotification(user.Id, user.IIDtoken, t.Event)
			continue
		}

		// Create InvitationReceived message
		notify_msg := session.NewMessage().InvitationReceived(light_event)

		// Filter event participants to protect privacy and create message
		filtered_participants := server.filterParticipantsMap(user.Id, t.Event.Participants)
		attendance_status_msg := session.NewMessage().AttendanceStatus(t.Event.EventId, filtered_participants)

		// Notify (use a channel because it is needed to know if message arrived)
		var future *Future

		if user.IIDtoken != "" {
			future = NewFuture(true)
		} else {
			future = NewFuture(false)
		}

		if ok := session.WriteAsync(future, notify_msg, attendance_status_msg); ok {
			futures[user.Id] = future.C
			log.Printf("< (%v) SEND EVENT INVITATION (event_id=%v)\n", user.Id, t.Event.EventId)
		} else {
			t.sendGcmNotification(user.Id, user.IIDtoken, t.Event)
		}
	}

	// Update invitation delivery status
	participants_changed := make([]uint64, 0, len(t.Target))
	eventDAO := ex.server.NewEventDAO()

	for participant_id, c := range futures {

		ok := <-c

		if ok {

			err := eventDAO.SetParticipantStatus(participant_id, t.Event.EventId, core.MessageStatus_CLIENT_DELIVERED) // participant changed

			if err == nil {
				t.Event.Participants[participant_id].Delivered = core.MessageStatus_CLIENT_DELIVERED
				// Add participant to changed set because delivery status has changed
				participants_changed = append(participants_changed, participant_id)
			} else {
				log.Println("NotifyEventInvitation Err:", err)
			}

		} else { // timeout or error
			user := t.Target[participant_id]
			t.sendGcmNotification(user.Id, user.IIDtoken, t.Event)
		}
	}

	// Notify changes to the rest of participants
	if len(participants_changed) > 0 {
		task := &NotifyParticipantChange{
			Event:               t.Event,
			ParticipantsChanged: participants_changed,
			Target:              core.GetParticipantsIdSlice(t.Event.Participants),
		}

		task.Run(ex)
	}
}

func (t *NotifyEventInvitation) sendGcmNotification(user_id uint64, token string, event *core.Event) {

	time_to_start := uint32(event.StartDate-core.GetCurrentTimeMillis()) / 1000
	ttl := core.MinUint32(time_to_start, GCM_MAX_TTL) // Seconds

	gcm_message := gcm.HttpMessage{
		To:         token,
		TimeToLive: uint(ttl),
		Data: gcm.Data{
			"msg_type":    "notification",
			"notify_type": GCM_NEW_EVENT_MESSAGE,
			"event_id":    event.EventId,
		},
	}

	sendGcmMessage(user_id, token, gcm_message)
}

func sendGcmMessage(user_id uint64, token string, message gcm.HttpMessage) {

	if token == "" {
		return
	}

	log.Printf("Sending GCM notification to %v\n", user_id)
	response, err := gcm.SendHttp(GCM_API_KEY, message)

	if err != nil {
		log.Println("SendGCMNotification Error:", err)
		if response != nil {
			log.Println("SendGCMNotification Response Error:", response.Error)
		}
	} else {
		log.Printf("SendGCMNotifcation Response %v\n", response)
	}
}
