package main

type i18nKey int

// i18n string keys
const (
	NotificationNewEventTitle i18nKey = iota
	NotificationNewEventBody
	NotificationEventCancelledTitle
	NotificationEventCancelledBody
	NotificationEventResponseAssistTitle
	NotificationEventResponseAssistBody
	NotificationEventResponseMaybeTitle
	NotificationEventResponseMaybeBody
	NotificationEventResponseNoAssistTitle
	NotificationEventResponseNoAssistBody
	NotificationNewFriendRequestTitle
	NotificationNewFriendRequestBody
	NotificationNewFriendTitle
	NotificationNewFriendBody
)

type i18nLang int

// Supported languages
const (
	EN i18nLang = iota // English
	ES                 // Spanish
)

var (
	language = map[i18nLang]map[i18nKey]string{
		ES: {
			// New event notification
			NotificationNewEventTitle: "Nuevo evento",
			NotificationNewEventBody:  "%v te ha invitado a un evento",
			// Event cancelled notification
			NotificationEventCancelledTitle: "%v",
			NotificationEventCancelledBody:  "%v ha cancelado el evento",
			// Participant will assist to event
			NotificationEventResponseAssistTitle: "%v",
			NotificationEventResponseAssistBody:  "%v asistirá al evento",
			// Participant may assist to event
			NotificationEventResponseMaybeTitle: "%v",
			NotificationEventResponseMaybeBody:  "%v quizá asista al evento",
			// Participant will assist to event
			NotificationEventResponseNoAssistTitle: "%v",
			NotificationEventResponseNoAssistBody:  "%v no asistirá al evento",
			// Friend request notification
			NotificationNewFriendRequestTitle: "Solicitud de amistad",
			NotificationNewFriendRequestBody:  "%v quiere ser tu amigo",
			// New friend notification
			NotificationNewFriendTitle: "Nuevo amigo",
			NotificationNewFriendBody:  "%v y tú ahora sois amigos",
		},
		EN: {
			// New event notification
			NotificationNewEventTitle: "New event",
			NotificationNewEventBody:  "%v has invited you to an event",
			// Event cancelled notification
			NotificationEventCancelledTitle: "%v",
			NotificationEventCancelledBody:  "%v has cancelled the event",
			// Participant will assist to event
			NotificationEventResponseAssistTitle: "%v",
			NotificationEventResponseAssistBody:  "%v will assist to event",
			// Participant may assist to event
			NotificationEventResponseMaybeTitle: "%v",
			NotificationEventResponseMaybeBody:  "%v may assist to event",
			// Participant will assist to event
			NotificationEventResponseNoAssistTitle: "%v",
			NotificationEventResponseNoAssistBody:  "%v will not assist to event",
			// Friend request notification
			NotificationNewFriendRequestTitle: "Friend request",
			NotificationNewFriendRequestBody:  "%v wants to be your friend",
			// New friend notification
			NotificationNewFriendTitle: "New friend",
			NotificationNewFriendBody:  "%v and you are friends now",
		},
	}
)

func T(lang i18nLang, key i18nKey) string {
	return language[lang][key]
}
