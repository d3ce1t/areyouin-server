package main

import (
	"errors"
	"peeple/areyouin/api"
	"peeple/areyouin/facebook"
	"peeple/areyouin/model"
	proto "peeple/areyouin/protocol"
)

var (
	ErrSessionNotConnected        = errors.New("session not connected")
	ErrAuthRequired               = errors.New("auth required")
	ErrNoAuthRequired             = errors.New("no auth required")
	ErrForbidden                  = errors.New("forbidden access")
	ErrUnregisteredMessage        = errors.New("unregistered message")
	ErrUnregisteredFriendsIgnored = errors.New("ignored unregistered participants")
	ErrAuthorMismatch             = errors.New("author mismatch")
	ErrOperationFailed            = errors.New("operation failed")
	ErrFriendNotFound             = errors.New("Friend not found")
)

func getNetErrorCode(err error, default_code int32) int32 {

	var err_code int32

	switch err {

	case ErrAuthRequired:
		err_code = proto.E_AUTH_REQUIRED

	case ErrUnregisteredFriendsIgnored:
		err_code = proto.E_INVALID_PARTICIPANT

	case ErrForbidden:
		err_code = proto.E_FORBIDDEN_ACCESS

	case ErrAuthorMismatch:
		err_code = proto.E_EVENT_AUTHOR_MISMATCH

	case ErrFriendNotFound:
		err_code = proto.E_FRIEND_NOT_FOUND

	case model.ErrInvalidEmail:
		err_code = proto.E_INPUT_INVALID_EMAIL_ADDRESS

	case model.ErrInvalidName:
		err_code = proto.E_INPUT_INVALID_USER_NAME

	case model.ErrInvalidStartDate:
		err_code = proto.E_EVENT_INVALID_START_DATE

	case model.ErrInvalidEndDate:
		err_code = proto.E_EVENT_INVALID_END_DATE

	case model.ErrAccountNotLinkedToFacebook:
		err_code = proto.E_ACCOUNT_NOT_LINKED_TO_FACEBOOK

	case model.ErrEventOutOfCreationWindow:
		err_code = proto.E_EVENT_OUT_OF_CREATE_WINDOW

	case model.ErrParticipantNotFound:
		err_code = proto.E_INVALID_EVENT_OR_PARTICIPANT

	case model.ErrEmptyInbox:
		err_code = proto.E_EMPTY_LIST

	case model.ErrInvalidUserOrPassword:
		err_code = proto.E_INVALID_USER_OR_PASSWORD

	case model.ErrParticipantsRequired:
		err_code = proto.E_EVENT_PARTICIPANTS_REQUIRED

	case model.ErrEventNotWritable:
		err_code = proto.E_EVENT_CANNOT_BE_MODIFIED

	case model.ErrFriendRequestAlreadyExist:
		err_code = proto.E_FRIEND_REQUEST_ALREADY_SENT

	case model.ErrAlreadyFriends:
		err_code = proto.E_ALREADY_FRIENDS

	case api.ErrEmailAlreadyExists:
		err_code = proto.E_EMAIL_EXISTS

	case api.ErrFacebookAlreadyExists:
		err_code = proto.E_FB_EXISTS

	case facebook.ErrFacebookAccessForbidden:
		err_code = proto.E_FB_INVALID_ACCESS_TOKEN

	default:
		err_code = default_code
	}

	return err_code
}
