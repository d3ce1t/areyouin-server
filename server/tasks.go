package main

import (
	_ "image/jpeg"
	"log"
	"peeple/areyouin/model"
)

type NotifyDatasetChanged struct {
	Users []int64
}

func (t *NotifyDatasetChanged) Run(ex *TaskExecutor) {

}

type ImportFacebookFriends struct {
	TargetUser *model.UserAccount
	Fbtoken    string // Facebook User Access token
}

func (task *ImportFacebookFriends) Run(ex *TaskExecutor) {

	server := ex.server

	addedFriends, err := server.Model.Friends.ImportFacebookFriends(task.TargetUser)
	if err != nil {
		log.Println("ImportFacebookFriends Error:", err)
		return
	}

	log.Printf("ImportFacebookFriends: %v friends imported\n", len(addedFriends))

	// Loop through added friends in order to notify them
	for _, newFriend := range addedFriends {

		// Send friends to existing user
		ex.Submit(&SendUserFriends{UserId: newFriend.Id()})

		// Send new friends notification
		token := newFriend.PushToken()
		sendGcmDataAvailableNotification(newFriend.Id(), &token, GCM_MAX_TTL)
	}

	if len(addedFriends) > 0 {
		// Notify target user
		ex.Submit(&SendUserFriends{UserId: task.TargetUser.Id()})
	}
}

type SendUserFriends struct {
	UserId int64
}

func (t *SendUserFriends) Run(ex *TaskExecutor) {

	server := ex.server

	if session := server.getSession(t.UserId); session != nil {

		friends, err := server.Model.Friends.GetAllFriends(t.UserId)
		if err != nil {
			log.Println("SendUserFriends Error:", err)
			return
		}

		if len(friends) > 0 {
			packet := session.NewMessage().FriendsList(convFriendList2Net(friends))
			if session.Write(packet) {
				log.Printf("< (%v) SEND USER FRIENDS (num.friends: %v)\n", t.UserId, len(friends))
			}
		}

	}
}
