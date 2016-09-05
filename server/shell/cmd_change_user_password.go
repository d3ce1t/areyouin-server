package shell

import (
	"errors"
	"fmt"
	"strconv"
)

type changeUserPasswordCmd struct {
}

func (c *changeUserPasswordCmd) Exec(shell *Shell, args []string) {

	userID, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	if len(args) != 3 {
		manageShellError(errors.New("New password isn't provided"))
	}

	user, err := shell.model.Accounts.GetUserAccount(userID)
	manageShellError(err)

	err = shell.model.Accounts.ChangePassword(user, args[2])
	manageShellError(err)

	fmt.Fprint(shell, "Password changed\n")
}
