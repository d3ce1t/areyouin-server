package main

import (
	_ "image/jpeg"
	"log"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
)

type NotifyDatasetChanged struct {
	Users []int64
}

func (t *NotifyDatasetChanged) Run(ex *TaskExecutor) {

}

type ImportFacebookFriends struct {
	TargetUser *core.UserAccount
	Fbtoken    string // Facebook User Access token
}

func (task *ImportFacebookFriends) Run(ex *TaskExecutor) {

	server := ex.server

	addedFriends, err := server.Model.Accounts.ImportFacebookFriends(task.TargetUser)
	if err != nil {
		log.Println("ImportFacebookFriends Error:", err)
		return
	}

	log.Printf("ImportFacebookFriends: %v friends imported\n", len(addedFriends))

	// Loop through added friends in order to notify them
	for _, newFriend := range addedFriends {

		// Send friends to existing user
		ex.Submit(&SendUserFriends{UserId: newFriend.Id})

		// Send new friends notification
		if newFriend.NetworkVersion == 0 || newFriend.NetworkVersion == 1 {
			sendGcmNewFriendNotification(newFriend.Id, newFriend.IIDtoken, task.TargetUser)
		} else {
			sendGcmDataAvailableNotification(newFriend.Id, newFriend.IIDtoken, GCM_MAX_TTL)
		}
	}

	if len(addedFriends) > 0 {
		// Notify target user
		ex.Submit(&SendUserFriends{UserId: task.TargetUser.GetUserId()})
	}
}

type SendUserFriends struct {
	UserId int64
}

func (t *SendUserFriends) Run(ex *TaskExecutor) {

	server := ex.server

	if session := server.GetSession(t.UserId); session != nil {

		friend_dao := dao.NewFriendDAO(server.DbSession)

		friends, err := friend_dao.LoadFriends(t.UserId, 0)
		if err != nil {
			log.Println("SendUserFriends Error:", err)
			return
		}

		if len(friends) > 0 {
			packet := session.NewMessage().FriendsList(friends)
			if session.Write(packet) {
				log.Printf("< (%v) SEND USER FRIENDS (num.friends: %v)\n", t.UserId, len(friends))
			}
		}

	}
}
