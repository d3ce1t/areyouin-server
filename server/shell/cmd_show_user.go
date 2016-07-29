package shell

import (
	"fmt"
	"peeple/areyouin/cqldao"
	"peeple/areyouin/utils"
	"strconv"
)

// show_user
func showUser(shell *Shell, args []string) {

	userID, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	userDAO := cqldao.NewUserDAO(shell.model.DbSession()).(*cqldao.UserDAO)
	user, err := userDAO.Int_LoadUserAccount(userID)
	manageShellError(err)

	validAccount, err := userDAO.Int_CheckUserConsistency(user)
	manageShellError(err)

	accountStatus := ""
	if !validAccount {
		accountStatus = "(¡¡¡INVALID STATUS!!!)"
	}

	fmt.Fprintln(shell, "---------------------------------")
	fmt.Fprintf(shell, "User details %v\n", accountStatus)
	fmt.Fprintln(shell, "---------------------------------")
	fmt.Fprintln(shell, "UserID:", user.Id)
	fmt.Fprintln(shell, "Name:", user.Name)
	fmt.Fprintln(shell, "Email:", user.Email)
	fmt.Fprintln(shell, "Email Verified:", user.EmailVerified)
	fmt.Fprintln(shell, "Created at:", utils.UnixMillisToTime(user.CreatedDate))
	fmt.Fprintln(shell, "Last connection:", utils.UnixMillisToTime(user.LastConn))
	fmt.Fprintln(shell, "Authtoken:", user.AuthToken)
	fmt.Fprintln(shell, "Fbid:", user.FbId)
	fmt.Fprintln(shell, "Fbtoken:", user.FbToken)

	fmt.Fprintln(shell, "---------------------------------")
	fmt.Fprintln(shell, "E-mail credentials")
	fmt.Fprintln(shell, "---------------------------------")

	if emailCred, err := userDAO.Int_LoadEmailCredential(user.Email); err == nil {
		fmt.Fprintln(shell, "E-mail:", emailCred.Email == user.Email)
		if emailCred.Password == cqldao.EMPTY_ARRAY_32B || emailCred.Salt == cqldao.EMPTY_ARRAY_32B {
			fmt.Fprintln(shell, "No password set")
		} else {
			fmt.Fprintf(shell, "Password: %x\n", emailCred.Password)
			fmt.Fprintf(shell, "Salt: %x\n", emailCred.Salt)
		}
		fmt.Fprintln(shell, "UserID Match:", emailCred.UserId == user.Id)
	} else {
		fmt.Fprintln(shell, "Error:", err)
	}

	fmt.Fprintln(shell, "---------------------------------")
	fmt.Fprintln(shell, "Facebook credentials")
	fmt.Fprintln(shell, "---------------------------------")

	if user.FbId != "" {
		fbCred, err := userDAO.Int_LoadFacebookCredential(user.FbId)
		if err == nil {
			fmt.Fprintln(shell, "FbId:", fbCred.FbId == user.FbId)
			fmt.Fprintln(shell, "FbToken:", fbCred.FbToken == user.FbToken)
			fmt.Fprintln(shell, "UserID Match:", fbCred.UserId == user.Id)
		} else {
			fmt.Fprintln(shell, "Error:", err)
		}
	} else {
		fmt.Fprintln(shell, "There isn't Facebook")
	}
	fmt.Fprintln(shell, "---------------------------------")

	if accountStatus != "" {
		fmt.Fprintf(shell, "\nACCOUNT INFO: %v\n", accountStatus)
	}
}
