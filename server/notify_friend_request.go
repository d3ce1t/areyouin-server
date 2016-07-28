package main

import (
	"log"
	"peeple/areyouin/model"
)

type NotifyFriendRequest struct {
	UserId        int64
	FriendRequest *model.FriendRequest
}

func (t *NotifyFriendRequest) Run(ex *TaskExecutor) {

	server := ex.server
	session := server.GetSession(t.UserId)

	if session != nil {
		session.Write(session.NewMessage().FriendRequestReceived(convFriendRequest2Net(t.FriendRequest)))
		log.Printf("< (%v) FRIEND REQUEST RECEIVED: %v\n", session, t.FriendRequest)
	} else {

		iid_token, err := server.Model.Accounts.GetPushToken(t.UserId)
		if err != nil {
			log.Printf("* NotifyFriendRequest error (userId %v) %v\n", t.UserId, err)
			return
		}

		sendGcmDataAvailableNotification(t.UserId, iid_token, GCM_MAX_TTL)
	}
}
