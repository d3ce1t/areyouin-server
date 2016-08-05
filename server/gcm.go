package main

import (
	"log"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"

	gcm "github.com/google/go-gcm"
)

// GCM Settings
const (
	GcmAPIKey = "AIzaSyAf-h1zJCRWNDt-dI3liL1yx4NEYjOq5GQ"
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

	ttlSeconds := uint((event.StartDate() - utils.GetCurrentTimeMillis()) / 1000)

	if token.Platform() == "Android" {
		sendToSync(userID, token, ttlSeconds)
	} else {
		notification := createNewEventNotification(event, i18nLang(token.Lang()))
		notification.ttl = ttlSeconds
		sendNotification(userID, token.Token(), notification)
	}
}

func sendEventCancelledNotification(event *model.Event, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendEventCancelledNotification err: %v", err)
		return
	}

	ttlSeconds := uint((event.EndDate() - utils.GetCurrentTimeMillis()) / 1000)

	if token.Platform() == "Android" {
		sendToSync(userID, token, ttlSeconds)
	} else {
		notification := createEventCancelledNotification(event, i18nLang(token.Lang()))
		notification.ttl = ttlSeconds
		sendNotification(userID, token.Token(), notification)
	}
}

func sendEventResponseNotification(event *model.Event, participantID int64, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendEventResponseNotification err: %v", err)
		return
	}

	ttlSeconds := uint((event.EndDate() - utils.GetCurrentTimeMillis()) / 1000)

	if token.Platform() == "Android" {
		sendToSync(userID, token, ttlSeconds)
	} else {
		notification := createEventResponseNotification(event, participantID, i18nLang(token.Lang()))
		notification.ttl = ttlSeconds
		sendNotification(userID, token.Token(), notification)
	}
}

func sendFriendRequestNotification(friendName string, userID int64) {

	m := model.Get("default")

	token, err := m.Accounts.GetPushToken(userID)
	if err != nil {
		log.Printf("sendFriendRequestNotification err: %v", err)
		return
	}

	if token.Platform() == "Android" {
		sendToSync(userID, token, GcmMaxTTL)
	} else {
		notification := createFriendRequestdNotification(friendName, i18nLang(token.Lang()))
		notification.ttl = GcmMaxTTL
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

	if token.Platform() == "Android" {
		sendToSync(userID, token, GcmMaxTTL)
	} else {
		notification := createNewFriendNotification(friendName, i18nLang(token.Lang()))
		notification.ttl = GcmMaxTTL
		sendNotification(userID, token.Token(), notification)
	}
}

func sendNotification(userID int64, token string, notification *Notification) {

	message := gcm.HttpMessage{
		To:         token,
		TimeToLive: notification.ttl,
		Priority:   "high",
		Notification: gcm.Notification{
			Title: notification.title,
			Body:  notification.body,
			Icon:  "icon_notification_25dp", // Android only (drawable name)
			Sound: "default",
			Color: "#009688", // Android only
		},
	}

	sendGcmMessage(userID, message)
}

// Send-to-Sync PUSH Message
func sendToSync(userID int64, token *model.IIDToken, ttl uint) {

	if token == nil || token.Token() == "" {
		return
	}

	gcmTTL := utils.MinUint(ttl, GcmMaxTTL) // Seconds

	message := gcm.HttpMessage{
		To:               token.Token(),
		TimeToLive:       gcmTTL,
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
	response, err := gcm.SendHttp(GcmAPIKey, message)

	if err != nil && response != nil {
		log.Printf("* (%v) GCM Error: %v (resp.Error: %v)\n", userID, err, response.Error)
	} else if err != nil {
		log.Printf("* (%v) GCM Error: %v\n", userID, err)
	} else {
		log.Printf("* (%v) GCM Response: %v\n", userID, response)
	}
}
