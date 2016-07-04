package main

import (
  core "peeple/areyouin/common"
  "log"
)

type NotifyFriendRequest struct {
	UserId 				int64
	FriendRequest *core.FriendRequest
}

func (t *NotifyFriendRequest) Run(ex *TaskExecutor) {

	server := ex.server
	session := server.GetSession(t.UserId)

	if session != nil {
		session.Write(session.NewMessage().FriendRequestReceived(t.FriendRequest))
		log.Printf("< (%v) FRIEND REQUEST RECEIVED: %v\n", session, t.FriendRequest)
	} else {

    userDAO := server.NewUserDAO()
    iid_token, err := userDAO.GetIIDToken(t.UserId)
    if err != nil {
      log.Printf("* NotifyFriendRequest error (userId %v) %v\n", t.UserId, err)
      return
    }

    sendGcmDataAvailableNotification(t.UserId, iid_token.Token, GCM_MAX_TTL)
  }
}
