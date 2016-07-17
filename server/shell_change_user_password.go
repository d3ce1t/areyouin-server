package main

import (
	"strconv"
  "errors"
  "fmt"
	"peeple/areyouin/dao"
)

func (shell *Shell) changeUserPassword(args []string) {

  user_id, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

  if len(args) != 3 {
    manageShellError(errors.New("New password isn't provided"))
  }

  var newPassword string = args[2]

  server := shell.server
	userDAO := dao.NewUserDAO(server.DbSession)
	user, err := userDAO.Load(user_id)
	manageShellError(err)

  _, err = userDAO.ResetEmailCredentialPassword(user.Id, user.Email, newPassword)
  manageShellError(err)
  fmt.Fprint(shell.io, "Password changed\n")
}
