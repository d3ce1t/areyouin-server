package main

import (
	"errors"

	"github.com/d3ce1t/areyouin-server/api"
	"github.com/d3ce1t/areyouin-server/facebook"
	"github.com/d3ce1t/areyouin-server/model"
	proto "github.com/d3ce1t/areyouin-server/protocol"
)

var (
	ErrSessionNotConnected        = errors.New("session not connected")
	ErrForbidden                  = errors.New("forbidden")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrUnregisteredMessage        = errors.New("unregistered message")
	ErrUnregisteredFriendsIgnored = errors.New("ignored unregistered participants")
	ErrAuthorMismatch             = errors.New("author mismatch")
	ErrOperationFailed            = errors.New("operation failed")
	ErrFriendNotFound             = errors.New("friend not found")
)

func getNetErrorCode(err error, default_code int32) int32 {

	var err_code int32

	switch err {

	case ErrUnauthorized:
		err_code = proto.E_UNAUTHORIZED

	case ErrUnregisteredFriendsIgnored:
		err_code = proto.E_INVALID_PARTICIPANT

	case ErrForbidden:
		err_code = proto.E_FORBIDDEN

	case ErrAuthorMismatch:
		err_code = proto.E_EVENT_AUTHOR_MISMATCH

	case ErrFriendNotFound:
		err_code = proto.E_FRIEND_NOT_FOUND

	case model.ErrInvalidEmail:
		err_code = proto.E_INPUT_INVALID_EMAIL_ADDRESS

	case model.ErrInvalidName:
		err_code = proto.E_INPUT_INVALID_USER_NAME

	/* DISABLED, it causes crash in iOS<=1.0.10
	case model.ErrInvalidPassword:
		err_code = proto.E_INPUT_INVALID_PASSWORD
	*/

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
		err_code = proto.E_EVENT_NOT_WRITABLE

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
