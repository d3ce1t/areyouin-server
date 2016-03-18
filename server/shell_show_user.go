package main

import (
	"fmt"
	core "peeple/areyouin/common"
	"strconv"
)

// show_user
func (shell *Shell) showUser(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	dao := server.NewUserDAO()
	user, err := dao.Load(user_id)
	manageShellError(err)

	valid_user, _ := user.IsValid()
	valid_account, err := dao.CheckValidAccount(user_id, true)

	if err != nil {
		fmt.Fprintln(shell.io, "Error checking account:", err)
	}

	account_status := ""
	if !valid_user || !valid_account {
		account_status = "¡¡¡INVALID STATUS!!!"
	}

	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintf(shell.io, "User details (%v)\n", account_status)
	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintln(shell.io, "UserID:", user.Id)
	fmt.Fprintln(shell.io, "Name:", user.Name)
	fmt.Fprintln(shell.io, "Email:", user.Email)
	fmt.Fprintln(shell.io, "Email Verified:", user.EmailVerified)
	fmt.Fprintln(shell.io, "Created at:", core.UnixMillisToTime(user.CreatedDate))
	fmt.Fprintln(shell.io, "Last connection:", core.UnixMillisToTime(user.LastConnection))
	fmt.Fprintln(shell.io, "Authtoken:", user.AuthToken)
	fmt.Fprintln(shell.io, "Fbid:", user.Fbid)
	fmt.Fprintln(shell.io, "Fbtoken:", user.Fbtoken)

	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintln(shell.io, "E-mail credentials")
	fmt.Fprintln(shell.io, "---------------------------------")

	if email, err := dao.LoadEmailCredential(user.Email); err == nil {
		fmt.Fprintln(shell.io, "E-mail:", email.Email == user.Email)
		if email.Password == core.EMPTY_ARRAY_32B || email.Salt == core.EMPTY_ARRAY_32B {
			fmt.Fprintln(shell.io, "No password set")
		} else {
			fmt.Fprintf(shell.io, "Password: %x\n", email.Password)
			fmt.Fprintf(shell.io, "Salt: %x\n", email.Salt)
		}
		fmt.Fprintln(shell.io, "UserID Match:", email.UserId == user.Id)
	} else {
		fmt.Fprintln(shell.io, "Error:", err)
	}

	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintln(shell.io, "Facebook credentials")
	fmt.Fprintln(shell.io, "---------------------------------")

	if user.HasFacebookCredentials() {
		facebook, err := dao.LoadFacebookCredential(user.Fbid)
		if err == nil {
			fmt.Fprintln(shell.io, "Fbid:", facebook.Fbid == user.Fbid)
			fmt.Fprintln(shell.io, "Fbtoken:", facebook.Fbtoken == user.Fbtoken)
			fmt.Fprintln(shell.io, "UserID Match:", facebook.UserId == user.Id)
		} else {
			fmt.Fprintln(shell.io, "Error:", err)
		}
	} else {
		fmt.Fprintln(shell.io, "There aren't credentials")
	}
	fmt.Fprintln(shell.io, "---------------------------------")

	if account_status != "" {
		fmt.Fprintf(shell.io, "\nACCOUNT INFO: %v\n", account_status)
	}
}
