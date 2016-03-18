package main

import (
	"strconv"
)

// close_session user_id
func (shell *Shell) closeSession(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	if session, ok := server.sessions.Get(user_id); ok {
		session.Exit()
	}
}
