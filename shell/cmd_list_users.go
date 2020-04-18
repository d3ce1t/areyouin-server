package shell

import (
	"fmt"

	"github.com/d3ce1t/areyouin-server/cqldao"
	"github.com/d3ce1t/areyouin-server/utils"
)

// list_users
type listUsersCmd struct {
}

func (c *listUsersCmd) Exec(shell *Shell, args []string) {

	userDAO := cqldao.NewUserDAO(shell.model.DbSession()).(*cqldao.UserDAO)

	users, err := userDAO.Int_LoadAllUserAccount()
	manageShellError(err)

	fmt.Fprintln(shell, rp("-", 108))
	fmt.Fprintf(shell, "| S | %-20s | %-15s | %-40s | %-16s |\n", "Id", "Name", "Email", "Last connection")
	fmt.Fprintln(shell, rp("-", 108))

	for _, user := range users {
		status_info := " "
		if valid, err := userDAO.Int_CheckUserConsistency(user); err != nil {
			status_info = "?"
		} else if !valid {
			status_info = "E"
		}

		fmt.Fprintf(shell, "| %v | %-20v | %-15v | %-40v | %-16v |\n",
			status_info, ff(user.Id, 20), ff(user.Name, 15), ff(user.Email, 40), ff(utils.MillisToTimeUTC(user.LastConn), 16))
	}
	fmt.Fprintln(shell, rp("-", 108))

	fmt.Fprintln(shell, "Num. Users:", len(users))
}
