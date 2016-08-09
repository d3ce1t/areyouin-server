package shell

import (
	"fmt"
	"peeple/areyouin/cqldao"
	"peeple/areyouin/utils"
)

// list_users
func listUsers(shell *Shell, args []string) {

	userDAO := cqldao.NewUserDAO(shell.model.DbSession()).(*cqldao.UserDAO)

	users, err := userDAO.Int_LoadAllUserAccount()
	manageShellError(err)

	fmt.Fprintln(shell, rp("-", 105))
	fmt.Fprintf(shell, "| S | %-20s | %-15s | %-40s | %-16s |\n", "Id", "Name", "Email", "Last connection")
	fmt.Fprintln(shell, rp("-", 105))

	for _, user := range users {
		status_info := " "
		if valid, err := userDAO.Int_CheckUserConsistency(user); err != nil {
			status_info = "?"
		} else if !valid {
			status_info = "E"
		}

		fmt.Fprintf(shell, "| %v | %-20v | %-15v | %-40v | %-16v |\n",
			status_info, ff(user.Id, 17), ff(user.Name, 15), ff(user.Email, 40), ff(utils.UnixMillisToTime(user.LastConn), 16))
	}
	fmt.Fprintln(shell, rp("-", 105))

	fmt.Fprintln(shell, "Num. Users:", len(users))
}
