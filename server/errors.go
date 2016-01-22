package main

import (
	"errors"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	proto "peeple/areyouin/protocol"
)

var (
	ErrAuthRequired               = errors.New("auth required")
	ErrUnhandledMessage           = errors.New("unhandled message")
	ErrUnknownMessage             = errors.New("unknown message")
	ErrNonFriendsIgnored          = errors.New("ignored non friends participants")
	ErrUnregisteredFriendsIgnored = errors.New("ignored unregistered participants")
)

func getNetErrorCode(err error, default_code int32) int32 {

	var err_code int32

	switch err {
	case core.ErrInvalidEmail:
		err_code = proto.E_INPUT_INVALID_EMAIL_ADDRESS
	case core.ErrInvalidName:
		err_code = proto.E_INPUT_INVALID_USER_NAME
	case dao.ErrEmailAlreadyExists:
		err_code = proto.E_EMAIL_EXISTS
	case dao.ErrFacebookAlreadyExists:
		err_code = proto.E_FB_EXISTS
	case dao.ErrGracePeriod:
		err_code = proto.E_OPERATION_FAILED
	default:
		err_code = default_code
	}

	return err_code
}
