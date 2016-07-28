package main

import (
	"fmt"
	"log"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"

	gcm "github.com/google/go-gcm"
)

// GCM MESSAGES
const (
	GCM_NEW_EVENT_MESSAGE       = 1
	GCM_NEW_FRIEND_MESSAGE      = 2
	GCM_EVENT_CANCELLED_MESSAGE = 3
	GCM_NEW_DATA_AVAILABLE      = 4
)

func sendGcmEventNotification(user_id int64, token *model.IIDToken, start_date int64, data string) {

	if token == nil || token.Token() == "" {
		return
	}

	time_to_start := uint32(start_date-utils.GetCurrentTimeMillis()) / 1000
	ttl := utils.MinUint32(time_to_start, GCM_MAX_TTL) // Seconds

	gcm_message := gcm.HttpMessage{
		To:         token.Token(),
		Priority:   "high",
		TimeToLive: uint(ttl),
		Data: gcm.Data{
			"msg_type":    "packet",
			"packet_data": data,
		},
	}

	sendGcmMessage(user_id, gcm_message)
}

func sendGcmNewFriendNotification(user_id int64, token *model.IIDToken, new_friend *model.Friend) {

	if token == nil || token.Token() == "" {
		return
	}

	gcm_message := gcm.HttpMessage{
		To:       token.Token(),
		Priority: "high",
		Data: gcm.Data{
			"msg_type":      "notification",
			"notify_type":   GCM_NEW_FRIEND_MESSAGE,
			"friend_name":   new_friend.Name(),
			"friend_id":     new_friend.Id(),
			"friend_digest": fmt.Sprintf("%x", new_friend.PictureDigest()),
		},
	}

	sendGcmMessage(user_id, gcm_message)
}

func sendGcmNewEventNotification(user_id int64, token *model.IIDToken, event *model.Event) {

	if token == nil || token.Token() == "" {
		return
	}

	time_to_start := uint32(event.StartDate()-utils.GetCurrentTimeMillis()) / 1000
	ttl := utils.MinUint32(time_to_start, GCM_MAX_TTL) // Seconds

	gcm_message := gcm.HttpMessage{
		To:         token.Token(),
		TimeToLive: uint(ttl),
		Priority:   "high",
		Data: gcm.Data{
			"msg_type":    "notification",
			"notify_type": GCM_NEW_EVENT_MESSAGE,
			"event_id":    event.Id(),
		},
	}

	sendGcmMessage(user_id, gcm_message)
}

// Send-to-Sync PUSH Message
func sendGcmDataAvailableNotification(user_id int64, token *model.IIDToken, ttl uint32) {

	if token == nil || token.Token() == "" {
		return
	}

	gcm_ttl := utils.MinUint32(ttl, GCM_MAX_TTL) // Seconds

	gcm_message := gcm.HttpMessage{
		To:               token.Token(),
		TimeToLive:       uint(gcm_ttl),
		Priority:         "high",
		CollapseKey:      "send-to-sync",
		ContentAvailable: true, // For iOS
		Data: gcm.Data{
			"msg_type":     "notification",
			"notify_type":  GCM_NEW_DATA_AVAILABLE,
			"created_date": utils.GetCurrentTimeMillis(),
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
