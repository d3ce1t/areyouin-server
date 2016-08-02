package model

type SignalType int

const (

	// Events

	// Event published
	SignalNewEvent SignalType = iota

	// Event cancelled
	SignalEventCancelled SignalType = iota

	// Event modified (picture, start date, ...)
	SignalEventInfoChanged SignalType = iota

	// Event participant list changed (added or removed participants)
	SignalEventParticipantsInvited SignalType = iota

	// Participant changed (response, invitationStatus)
	SignalParticipantChanged SignalType = iota

	// Users

	// New registered user
	SignalNewUserAccount SignalType = iota

	// Friends

	// Friends imported
	SignalNewFriendsImported SignalType = iota

	// Friend request sent
	SignalNewFriendRequest SignalType = iota

	// Friend request accepted
	SignalFriendRequestAccepted SignalType = iota

	// Friend request cancelled
	SignalFriendRequestCancelled SignalType = iota
)

type Signal struct {
	Type SignalType
	Data map[string]interface{}
}
