package main

import (
	"errors"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	proto "peeple/areyouin/protocol"
)

var (
	ErrSessionNotConnected        = errors.New("session not connected")
	ErrAuthRequired               = errors.New("auth required")
	ErrNoAuthRequired             = errors.New("no auth required")
	ErrUnhandledMessage           = errors.New("unhandled message")
	ErrUnknownMessage             = errors.New("unknown message")
	ErrNonFriendsIgnored          = errors.New("ignored non friends participants")
	ErrUnregisteredFriendsIgnored = errors.New("ignored unregistered participants")
	ErrParticipantsRequired       = errors.New("participants required")
	ErrAuthorDeliveryError        = errors.New("event coudn't be delivered to author")
	ErrShellInvalidArgs           = errors.New("Invalid args")
)

func getNetErrorCode(err error, default_code int32) int32 {

	var err_code int32

	switch err {
	case ErrAuthRequired:
		err_code = proto.E_AUTH_REQUIRED
	case core.ErrInvalidEmail:
		err_code = proto.E_INPUT_INVALID_EMAIL_ADDRESS
	case core.ErrInvalidName:
		err_code = proto.E_INPUT_INVALID_USER_NAME
	case core.ErrInvalidStartDate:
		err_code = proto.E_EVENT_INVALID_START_DATE
	case core.ErrInvalidEndDate:
		err_code = proto.E_EVENT_INVALID_END_DATE
	case dao.ErrEmailAlreadyExists:
		err_code = proto.E_EMAIL_EXISTS
	case dao.ErrFacebookAlreadyExists:
		err_code = proto.E_FB_EXISTS
	case dao.ErrGracePeriod:
		err_code = proto.E_OPERATION_FAILED
	case ErrParticipantsRequired:
		err_code = proto.E_EVENT_PARTICIPANTS_REQUIRED
	case ErrNonFriendsIgnored:
		err_code = proto.E_INVALID_PARTICIPANT
	case ErrUnregisteredFriendsIgnored:
		err_code = proto.E_INVALID_PARTICIPANT
	default:
		err_code = default_code
	}

	return err_code
}
