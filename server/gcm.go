package main

import (
	"log"

	"github.com/d3ce1t/areyouin-server/model"
	"github.com/d3ce1t/areyouin-server/utils"

	gcm "github.com/google/go-gcm"
)

// GCM Settings
const (
	GcmMaxTTL = 2419200
)

// GCM Messages
const (
	GcmNewDataAvailable = 4
)

func sendNewEventNotification(event *model.Event, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendNewEventNotification err: %v", err)
		return
	}

	ttlSeconds := uint(event.StartDate().Sub(utils.GetCurrentTimeUTC()).Seconds())

	if token.Version() <= 2 {
		sendToSync(userID, token.Token(), ttlSeconds)
	} else {

		// Send notification
		gcmTTL := utils.MinUint(ttlSeconds, GcmMaxTTL) // ttlSeconds

		sendGcmMessage(userID, gcm.HttpMessage{
			To:               token.Token(),
			TimeToLive:       &gcmTTL,
			Priority:         "high",
			Notification:     createNewEventNotification(event),
			ContentAvailable: true, // For iOS
		})

		if token.Platform() == PLATFORM_ANDROID {
			// Android push composed of notification + data isn't received directly by
			// app. So send a second push with send-to-sync data.
			sendToSync(userID, token.Token(), ttlSeconds)
		}
	}
}

func sendEventCancelledNotification(event *model.Event, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendEventCancelledNotification err: %v", err)
		return
	}

	ttlSeconds := uint(event.EndDate().Sub(utils.GetCurrentTimeUTC()).Seconds())

	if token.Version() <= 2 {
		sendToSync(userID, token.Token(), ttlSeconds)
	} else {

		// Send notification

		gcmTTL := utils.MinUint(ttlSeconds, GcmMaxTTL) // ttlSeconds

		sendGcmMessage(userID, gcm.HttpMessage{
			To:               token.Token(),
			TimeToLive:       &gcmTTL,
			Priority:         "high",
			Notification:     createEventCancelledNotification(event),
			ContentAvailable: true, // For iOS, like send to sync
		})

		if token.Platform() == PLATFORM_ANDROID {
			// Android push composed of notification + data isn't received directly by
			// app. So send a second push with send-to-sync data.
			sendToSync(userID, token.Token(), ttlSeconds)
		}
	}
}

func sendEventResponseNotification(event *model.Event, participantID int64, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendEventResponseNotification err: %v", err)
		return
	}

	ttlSeconds := uint(event.EndDate().Sub(utils.GetCurrentTimeUTC()).Seconds())

	if token.Version() <= 2 {
		sendToSync(userID, token.Token(), ttlSeconds)
	} else {
		notification := createEventResponseNotification(event, participantID)
		sendNotificationWithTTL(userID, token.Token(), notification, ttlSeconds)
	}
}

func sendFriendRequestNotification(friendName string, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendFriendRequestNotification err: %v", err)
		return
	}

	if token.Version() <= 2 {
		sendToSync(userID, token.Token(), GcmMaxTTL)
	} else {
		notification := createFriendRequestdNotification(friendName)
		sendNotification(userID, token.Token(), notification)
	}
}

func sendNewFriendNotification(friendName string, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendNewFriendNotification err: %v", err)
		return
	}

	if token.Version() <= 2 {
		sendToSync(userID, token.Token(), GcmMaxTTL)
	} else {
		notification := createNewFriendNotification(friendName)
		sendNotification(userID, token.Token(), notification)
	}
}

func sendNotification(userID int64, token string, notification *gcm.Notification) {

	if notification == nil {
		log.Println("* WARNING: sendNotification -> Skip notification because nil")
		return
	}

	message := gcm.HttpMessage{
		To:           token,
		Priority:     "high",
		Notification: notification,
	}

	sendGcmMessage(userID, message)
}

func sendNotificationWithTTL(userID int64, token string, notification *gcm.Notification, ttl uint) {

	if notification == nil {
		log.Println("* WARNING: sendNotificationWithTTL -> Skip notification because nil")
		return
	}

	gcmTTL := utils.MinUint(ttl, GcmMaxTTL) // Seconds

	message := gcm.HttpMessage{
		To:           token,
		TimeToLive:   &gcmTTL,
		Priority:     "high",
		Notification: notification,
	}

	sendGcmMessage(userID, message)
}

// Send-to-Sync PUSH Message
func sendToSync(userID int64, token string, ttl uint) {

	if token == "" {
		return
	}

	gcmTTL := utils.MinUint(ttl, GcmMaxTTL) // Seconds

	message := gcm.HttpMessage{
		To:               token,
		TimeToLive:       &gcmTTL,
		Priority:         "high",
		CollapseKey:      "send-to-sync",
		ContentAvailable: true, // For iOS
		Data: gcm.Data{
			"msg_type":     "notification",
			"notify_type":  GcmNewDataAvailable,
			"created_date": utils.GetCurrentTimeMillis(),
		},
	}

	sendGcmMessage(userID, message)
}

func sendGcmMessage(userID int64, message gcm.HttpMessage) {

	log.Printf("< (%v) Send GCM notification\n", userID)
	response, err := gcm.SendHttp(globalConfig.FirebaseAPIKey(), message)

	if err != nil && response != nil {
		log.Printf("* (%v) GCM Error: %v (resp.Error: %v)\n", userID, err, response.Error)
	} else if err != nil {
		log.Printf("* (%v) GCM Error: %v\n", userID, err)
	} else {
		log.Printf("* (%v) GCM Response: %v\n", userID, response)
	}
}
