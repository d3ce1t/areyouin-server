package main

import (
	"strconv"
)

// send_auth_error user_id
func (shell *Shell) sendAuthError(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	if session, ok := server.sessions.Get(user_id); ok {
		sendAuthError(session)
	}
}
