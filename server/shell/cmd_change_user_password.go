package shell

import (
	"errors"
	"fmt"
	"strconv"
)

func changeUserPassword(shell *Shell, args []string) {

	userID, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	if len(args) != 3 {
		manageShellError(errors.New("New password isn't provided"))
	}

	var newPassword string = args[2]

	user, err := shell.model.Accounts.GetUserAccount(userID)
	manageShellError(err)

	err = shell.model.Accounts.ChangePassword(user, newPassword)
	manageShellError(err)

	fmt.Fprint(shell, "Password changed\n")
}
