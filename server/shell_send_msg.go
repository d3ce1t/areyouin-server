package main

import (
	"fmt"
	"strconv"
)

// send_msg client
func (shell *Shell) sendMsg(args []string) {

	user_id, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	if len(args) < 2 {
		manageShellError(ErrShellInvalidArgs)
	}

	server := shell.server
	userDAO := server.NewUserDAO()

	user_account, err := userDAO.Load(user_id)
	manageShellError(err)

	sendGcmDataAvailableNotification(user_account.Id, user_account.IIDtoken, 3600)
	fmt.Fprintf(shell.io, "Message Sent\n")
}
