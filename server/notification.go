package main

import (
	"fmt"
	"peeple/areyouin/api"
	"peeple/areyouin/model"
)

type Notification struct {
	title string
	body  string
	ttl   uint
}

func createNewEventNotification(event *model.Event, lang i18nLang) *Notification {

	notification := &Notification{
		title: T(lang, NotificationNewEventTitle),
		body:  fmt.Sprintf(T(lang, NotificationNewEventBody), event.AuthorName()),
	}

	return notification
}

func createEventCancelledNotification(event *model.Event, lang i18nLang) *Notification {

	notification := &Notification{
		title: fmt.Sprintf(T(lang, NotificationEventCancelledTitle), event.Title()),
		body:  fmt.Sprintf(T(lang, NotificationEventCancelledBody), event.AuthorName()),
	}

	return notification
}

func createEventResponseNotification(event *model.Event, participantID int64, lang i18nLang) *Notification {

	participant := event.GetParticipant(participantID)

	var title, body string

	switch participant.Response() {
	case api.AttendanceResponse_ASSIST:
		title = T(lang, NotificationEventResponseAssistTitle)
		body = T(lang, NotificationEventResponseAssistBody)
	case api.AttendanceResponse_MAYBE:
		title = T(lang, NotificationEventResponseMaybeTitle)
		body = T(lang, NotificationEventResponseMaybeBody)
	case api.AttendanceResponse_NO_ASSIST:
		title = T(lang, NotificationEventResponseNoAssistTitle)
		body = T(lang, NotificationEventResponseNoAssistBody)
	}

	notification := &Notification{
		title: fmt.Sprintf(title, event.Title()),
		body:  fmt.Sprintf(body, participant.Name()),
	}

	return notification
}

func createFriendRequestdNotification(friendName string, lang i18nLang) *Notification {
	notification := &Notification{
		title: T(lang, NotificationNewFriendRequestTitle),
		body:  fmt.Sprintf(T(lang, NotificationNewFriendRequestBody), friendName),
	}
	return notification
}

func createNewFriendNotification(friendName string, lang i18nLang) *Notification {
	notification := &Notification{
		title: T(lang, NotificationNewFriendTitle),
		body:  fmt.Sprintf(T(lang, NotificationNewFriendBody), friendName),
	}
	return notification
}
