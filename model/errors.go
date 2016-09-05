package model

import (
	"errors"
	"peeple/areyouin/api"
)

var (
	ErrNotFound = api.ErrNotFound

	// User Account
	ErrInvalidEmail    = errors.New("invalid e-mail address")
	ErrInvalidName     = errors.New("invalid user name")
	ErrInvalidPassword = errors.New("password is too short")
	ErrNoCredentials   = errors.New("no credentials")

	// Event
	ErrInvalidStartDate = errors.New("invalid start date")
	ErrInvalidEndDate   = errors.New("invalid end date")
	ErrInvalidEventData = errors.New("invalidad event data")

	ErrModelAlreadyExist         = errors.New("cannot register model because it already exists")
	ErrModelNotFound             = errors.New("model not found")
	ErrModelInconsistency        = errors.New("Model has an inconsistency that requires admin fixes")
	ErrImageOutOfBounds          = errors.New("image is out of bounds")
	ErrInvalidUserOrPassword     = errors.New("invalid user or password")
	ErrInvalidAuthor             = errors.New("invalid author")
	ErrParticipantsRequired      = errors.New("participants required")
	ErrEventOutOfCreationWindow  = errors.New("event out of allowed creationg window")
	ErrAuthorDeliveryError       = errors.New("event coudn't be delivered to author")
	ErrEventNotWritable          = errors.New("event isn't writable")
	ErrParticipantNotFound       = errors.New("participant not found")
	ErrEmptyInbox                = errors.New("user inbox is empty")
	ErrAlreadyFriends            = errors.New("already friends")
	ErrFriendRequestAlreadyExist = errors.New("friend request already exists")

	ErrAccountNotLinkedToFacebook = errors.New("account isn't linked to facebook")
)
