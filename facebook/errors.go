package facebook

import (
	"errors"
)

var (
	ErrMissingFields = errors.New("missing fields in result")
	ErrFacebookAccessForbidden = errors.New("no facebook access")
)
