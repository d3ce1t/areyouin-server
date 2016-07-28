package main

import (
	"errors"
	"fmt"
	"strconv"
)

func (shell *Shell) changeUserPassword(args []string) {

	userID, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	if len(args) != 3 {
		manageShellError(errors.New("New password isn't provided"))
	}

	var newPassword string = args[2]

	server := shell.server

	user, err := server.Model.Accounts.GetUserAccount(userID)
	manageShellError(err)

	err = server.Model.Accounts.ChangePassword(user, newPassword)
	manageShellError(err)

	fmt.Fprint(shell.io, "Password changed\n")
}
