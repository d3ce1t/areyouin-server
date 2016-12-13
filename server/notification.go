package main

import (
	"encoding/json"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/model"

	gcm "github.com/google/go-gcm"
)

func createNewEventNotification(event *model.Event) *gcm.Notification {

	bodyArgs, _ := json.Marshal([]string{event.AuthorName()})

	notification := &gcm.Notification{
		TitleLocKey: "notification.event.new.title",
		BodyLocKey:  "notification.event.new.body",
		BodyLocArgs: string(bodyArgs),
		Icon:        "icon_notification_25dp", // Android only (drawable name)
		Sound:       "default",
		Color:       "#009688", // Android only
	}

	return notification
}

func createEventCancelledNotification(event *model.Event) *gcm.Notification {

	titleArgs, _ := json.Marshal([]string{event.Title()})
	bodyArgs, _ := json.Marshal([]string{event.AuthorName()})

	notification := &gcm.Notification{
		TitleLocKey:  "notification.event.cancelled.title",
		TitleLocArgs: string(titleArgs),
		BodyLocKey:   "notification.event.cancelled.body",
		BodyLocArgs:  string(bodyArgs),
		Icon:         "icon_notification_25dp", // Android only (drawable name)
		Sound:        "default",
		Color:        "#009688", // Android only
	}

	return notification
}

func createEventResponseNotification(event *model.Event, participantID int64) *gcm.Notification {

	participant, _ := event.Participants.Get(participantID)
	titleArgs, _ := json.Marshal([]string{event.Title()})
	bodyArgs, _ := json.Marshal([]string{participant.Name()})

	var titleKey, bodyKey string

	switch participant.Response() {
	case api.AttendanceResponse_ASSIST:
		titleKey = "notification.event.response.assist.title"
		bodyKey = "notification.event.response.assist.body"
	case api.AttendanceResponse_MAYBE:
		titleKey = "notification.event.response.maybe.title"
		bodyKey = "notification.event.response.maybe.body"
	case api.AttendanceResponse_NO_ASSIST:
		titleKey = "notification.event.response.no_assist.title"
		bodyKey = "notification.event.response.no_assist.body"
	case api.AttendanceResponse_NO_RESPONSE:
		log.Println("* WARNING: createEventResponseNotification with NO_RESPONSE value")
		return nil
	}

	notification := &gcm.Notification{
		TitleLocKey:  titleKey,
		TitleLocArgs: string(titleArgs),
		BodyLocKey:   bodyKey,
		BodyLocArgs:  string(bodyArgs),
		Icon:         "icon_notification_25dp", // Android only (drawable name)
		Sound:        "default",
		Color:        "#009688", // Android only
	}

	return notification
}

func createFriendRequestdNotification(friendName string) *gcm.Notification {

	bodyArgs, _ := json.Marshal([]string{friendName})

	notification := &gcm.Notification{
		TitleLocKey: "notification.friend_request.new.title",
		BodyLocKey:  "notification.friend_request.new.body",
		BodyLocArgs: string(bodyArgs),
		Icon:        "icon_notification_25dp", // Android only (drawable name)
		Sound:       "default",
		Color:       "#009688", // Android only
	}

	return notification
}

func createNewFriendNotification(friendName string) *gcm.Notification {

	bodyArgs, _ := json.Marshal([]string{friendName})

	notification := &gcm.Notification{
		TitleLocKey: "notification.friend.new.title",
		BodyLocKey:  "notification.friend.new.body",
		BodyLocArgs: string(bodyArgs),
		Icon:        "icon_notification_25dp", // Android only (drawable name)
		Sound:       "default",
		Color:       "#009688", // Android only
	}

	return notification
}

/*func createFriendJoinedNotification(friendName string) *gcm.Notification {

}*/
