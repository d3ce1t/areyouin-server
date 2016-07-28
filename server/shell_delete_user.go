package main

// delete_user $user_id --force
func (shell *Shell) deleteUser(args []string) {

	/*user_id, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	user, err := dao.Load(user_id)
	manageShellError(err)

	if len(args) == 2 {
		err = dao.Delete(user)

		if err != nil {
			fmt.Fprintln(shell.io, "Error:", err)
			fmt.Fprintln(shell.io, "Try command:")
			fmt.Fprintf(shell.io, "\tdelete_user %d --force\n", user_id)
			return
		}
	} else if len(args) > 2 {

		if args[2] != "--force" {
			manageShellError(ErrShellInvalidArgs)
		}

		// Try remove user account
		if err := dao.DeleteUserAccount(user_id); err == nil {
			fmt.Fprintln(shell.io, "User account removed")
		} else {
			fmt.Fprintln(shell.io, "Removing user account error:", err)
		}

		// Try remove e-mail credential
		if user.Email != "" {
			email_credential, err := dao.LoadEmailCredential(user.Email)
			if err == nil && email_credential.UserId == user.Id {
				if err := dao.DeleteEmailCredentials(user.Email); err == nil {
					fmt.Fprintln(shell.io, "E-mail credential removed")
				} else {
					fmt.Fprintln(shell.io, "Removing e-mail credential error:", err)
				}
			}
		}

		// Try remove facebook credential
		if user.Fbid != "" {
			facebook_credential, err := dao.LoadFacebookCredential(user.Fbid)
			if err == nil && facebook_credential.UserId == user.Id {
				if err := dao.DeleteFacebookCredentials(user.Fbid); err == nil {
					fmt.Fprintln(shell.io, "Facebook credential removed")
				} else {
					fmt.Fprintln(shell.io, "Removing facebook credential error:", err)
				}
			}
		}

	}

	fmt.Fprintf(shell.io, "User with id %d has been removed\n", user_id)*/
}
