package main

import (
	"strconv"
)

// ping client
func (shell *Shell) pingClient(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	var repeat_times uint64 = 1

	if len(args) >= 3 {
		repeat_times, err = strconv.ParseUint(args[2], 10, 32)
		manageShellError(err)
	}

	server := shell.server
	if session, ok := server.sessions.Get(user_id); ok {
		for i := uint64(0); i < repeat_times; i++ {
			session.Write(session.NewMessage().Ping())
		}
	}
}
