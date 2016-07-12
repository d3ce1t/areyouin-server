package main

import (
  core "peeple/areyouin/common"
  gcm "github.com/google/go-gcm"
  "log"
  "fmt"
)

// GCM MESSAGES
const (
	GCM_NEW_EVENT_MESSAGE       = 1
	GCM_NEW_FRIEND_MESSAGE      = 2
	GCM_EVENT_CANCELLED_MESSAGE = 3
  GCM_NEW_DATA_AVAILABLE      = 4
)

func sendGcmEventNotification(user_id int64, token string, start_date int64, data string) {

  if token == "" {
    return
  }

	time_to_start := uint32(start_date-core.GetCurrentTimeMillis()) / 1000
	ttl := core.MinUint32(time_to_start, GCM_MAX_TTL) // Seconds

	gcm_message := gcm.HttpMessage{
		To:         token,
		Priority:   "high",
		TimeToLive: uint(ttl),
		Data: gcm.Data{
			"msg_type":    "packet",
			"packet_data": data,
		},
	}

	sendGcmMessage(user_id, gcm_message)
}

func sendGcmNewFriendNotification(user_id int64, token string, new_friend core.UserFriend) {

  if token == "" {
    return
  }

	gcm_message := gcm.HttpMessage{
		To:       token,
		Priority: "high",
		Data: gcm.Data{
			"msg_type":      "notification",
			"notify_type":   GCM_NEW_FRIEND_MESSAGE,
			"friend_name":   new_friend.GetName(),
			"friend_id":     new_friend.GetUserId(),
			"friend_digest": fmt.Sprintf("%x", new_friend.GetPictureDigest()),
		},
	}

	sendGcmMessage(user_id, gcm_message)
}

func sendGcmNewEventNotification(user_id int64, token string, event *core.Event) {

  if token == "" {
    return
  }

	time_to_start := uint32(event.StartDate - core.GetCurrentTimeMillis()) / 1000
	ttl := core.MinUint32(time_to_start, GCM_MAX_TTL) // Seconds

	gcm_message := gcm.HttpMessage{
		To:         token,
		TimeToLive: uint(ttl),
		Priority:   "high",
		Data: gcm.Data{
			"msg_type":    "notification",
			"notify_type": GCM_NEW_EVENT_MESSAGE,
			"event_id":    event.EventId,
		},
	}

	sendGcmMessage(user_id, gcm_message)
}

// Send-to-Sync PUSH Message
func sendGcmDataAvailableNotification(user_id int64, token string, ttl uint32) {

  if token == "" {
    return
  }

  gcm_ttl := core.MinUint32(ttl, GCM_MAX_TTL) // Seconds

  gcm_message := gcm.HttpMessage{
		To:         token,
		TimeToLive: uint(gcm_ttl),
		Priority:   "high",
    CollapseKey: "send-to-sync",
    ContentAvailable: true, // For iOS
		Data: gcm.Data{
			"msg_type":    "notification",
			"notify_type": GCM_NEW_DATA_AVAILABLE,
      "created_date": core.GetCurrentTimeMillis(),
		},
	}

  sendGcmMessage(user_id, gcm_message)
}

func sendGcmMessage(user_id int64, message gcm.HttpMessage) {

	log.Printf("< (%v) Send GCM notification\n", user_id)
	response, err := gcm.SendHttp(GCM_API_KEY, message)

	if err != nil && response != nil {
		log.Printf("* (%v) GCM Error: %v (resp.Error: %v)\n", user_id, err, response.Error)
	} else if err != nil {
		log.Printf("* (%v) GCM Error: %v\n", user_id, err)
	} else {
		log.Printf("* (%v) GCM Response: %v\n", user_id, response)
	}
}
