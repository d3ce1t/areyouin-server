package main

import (
	"fmt"
	"log"
	core "peeple/areyouin/common"
)

// list_users
func (shell *Shell) listUserAccounts(args []string) {

	server := shell.server
	dao := server.NewUserDAO()
	users, err := dao.LoadAllUsers()
	manageShellError(err)

	fmt.Fprintln(shell.io, rp("-", 105))
	fmt.Fprintf(shell.io, "| S | %-17s | %-15s | %-40s | %-16s |\n", "Id", "Name", "Email", "Last connection")
	fmt.Fprintln(shell.io, rp("-", 105))

	for _, user := range users {
		status_info := " "
		if valid, err := dao.CheckValidAccountObject(user.Id, user.Email, user.Fbid, true); err != nil {
			log.Println("ListUserAccountsError:", err)
			status_info = "?"
		} else if !valid {
			status_info = "E"
		}

		fmt.Fprintf(shell.io, "| %v | %-17v | %-15v | %-40v | %-16v |\n",
			status_info, ff(user.Id, 17), ff(user.Name, 15), ff(user.Email, 40), ff(core.UnixMillisToTime(user.LastConnection), 16))
	}
	fmt.Fprintln(shell.io, rp("-", 105))

	fmt.Fprintln(shell.io, "Num. Users:", len(users))
}
