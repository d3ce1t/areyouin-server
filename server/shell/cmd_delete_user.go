package shell

import (
	"fmt"
	"peeple/areyouin/cqldao"
	"strconv"
)

type deleteUserCmd struct {
}

// delete_user $user_id --force
func (c *deleteUserCmd) Exec(shell *Shell, args []string) {

	userID, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	userDAO := cqldao.NewUserDAO(shell.model.DbSession()).(*cqldao.UserDAO)

	if len(args) == 2 {

		user, err := userDAO.Load(userID)
		manageShellError(err)

		if err := userDAO.Delete(user); err != nil {
			fmt.Fprintln(shell, "Error:", err)
			fmt.Fprintln(shell, "Try command:")
			fmt.Fprintf(shell, "\tdelete_user %d --force\n", userID)
			return
		}
	} else if len(args) > 2 {

		if args[2] != "--force" {
			manageShellError(ErrShellInvalidArgs)
		}

		user, err := userDAO.Int_LoadUserAccount(userID)
		manageShellError(err)

		if err := userDAO.Delete(user); err != nil {
			fmt.Fprintln(shell, "Error:", err)
			fmt.Fprintln(shell, "Try command:")
			fmt.Fprintf(shell, "\tdelete_user %d --force\n", userID)
			return
		}

	}

	fmt.Fprintf(shell, "User with id %d has been removed\n", userID)
}
