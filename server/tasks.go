package main

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"log"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
)

type NotifyDatasetChanged struct {
	Users []int64
}

func (t *NotifyDatasetChanged) Run(ex *TaskExecutor) {

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

	friend_dao := dao.NewFriendDAO(server.DbSession)
	storedFriends, err := friend_dao.LoadFriendsMap(task.TargetUser.GetUserId())
	if err != nil {
		log.Println("ImportFacebookFriends Error:", err)
		return
	}

	log.Printf("ImportFacebookFriends: %v friends found\n", len(fbFriends))

	user_dao := dao.NewUserDAO(server.DbSession)
	counter := 0

	// Loop Facebook friends in order to get AyiID
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
			continue
		}

		log.Printf("ImportFacebookFriends: %v and %v are Facebook Friends\n", task.TargetUser.GetUserId(), friend_id)

		// Assume that if friend_id isn't in stored friends, then current user id isn't either
		// in the other user friends list
		if _, ok := storedFriends[friendUser.Id]; !ok {
			friendUser.Name = fbFriend.Name // Use Facebook name because is familiar to user
			friend_dao.MakeFriends(task.TargetUser, friendUser)
			log.Printf("ImportFacebookFriends: %v and %v are now AreYouIN friends\n", task.TargetUser.GetUserId(), friendUser.Id)

			// Send friends to existing user
			ex.Submit(&SendUserFriends{UserId: friend_id})

			// Send new friends notification
			if friendUser.NetworkVersion == 0 || friendUser.NetworkVersion == 1 {
				sendGcmNewFriendNotification(friendUser.Id, friendUser.IIDtoken, task.TargetUser)
				counter++
			} else {
				sendGcmDataAvailableNotification(friendUser.Id, friendUser.IIDtoken, GCM_MAX_TTL)
			}
		}

	}

	if counter > 0 {
		// Send friends to user that initiated import
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

type LoadFacebookProfilePicture struct {
	User    *core.UserAccount
	Fbtoken string
}

func (task *LoadFacebookProfilePicture) Run(ex *TaskExecutor) {

	server := ex.server

	// Get profile picture
	fbsession := fb.NewSession(task.Fbtoken)
	picture_bytes, err := fb.GetProfilePicture(fbsession)
	if err != nil {
		log.Println("LoadFacebookProfilePicture: ", err)
		return
	}

	// Decode image
	original_image, _, err := image.Decode(bytes.NewReader(picture_bytes))
	if err != nil {
		log.Println("LoadFacebookProfilePicture: ", err)
		return
	}

	// Resize image to 512x512
	picture_bytes, err = core.ResizeImage(original_image, core.PROFILE_PICTURE_MAX_WIDTH)
	if err != nil {
		log.Println("LoadFacebookProfilePicture: ", err)
		return
	}

	// Change profile picture
	if err = server.Accounts.ChangeProfilePicture(task.User, picture_bytes); err != nil {
		log.Println("LoadFacebookProfilePicture: ", err)
		return
	}

	log.Printf("LoadFacebookProfilePicture: Profile picture updated (digest=%x)\n", task.User.PictureDigest)

	session := server.GetSession(task.User.Id)
	if session != nil {
		if session.ProtocolVersion < 2 {
			task.User.Picture = picture_bytes
			session.Write(session.NewMessage().UserAccount(task.User))
			log.Printf("< (%v) SEND USER ACCOUNT INFO (%v bytes)\n", session.UserId, len(task.User.Picture))
		} else {
			session.Write(session.NewMessage().UserAccount(task.User))
			log.Printf("< (%v) SEND USER ACCOUNT INFO\n", session.UserId)
		}
	}
}
