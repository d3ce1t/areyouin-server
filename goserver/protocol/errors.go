package protocol

const (
	E_NO_ERROR int32 = iota
	E_INVALID_USER
	E_USER_EXISTS
	E_FB_MISSING_DATA
	E_FB_INVALID_TOKEN
	E_MALFORMED_MESSAGE
)

const (
	OK_AUTH int32 = iota
)
