package main

import (
	"fmt"
	"strconv"
)

// make_friends user1 user2
func (shell *Shell) makeFriends(args []string) {

	friend_one_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	friend_two_id, err := strconv.ParseUint(args[2], 10, 64)
	manageShellError(err)

	server := shell.server
	userDAO := server.NewUserDAO()
	friendDAO := server.NewFriendDAO()

	user1, err := userDAO.Load(friend_one_id)
	manageShellError(err)

	user2, err := userDAO.Load(friend_two_id)
	manageShellError(err)

	err = friendDAO.MakeFriends(user1, user2)
	manageShellError(err)

	fmt.Fprintf(shell.io, "%v and %v are now friends\n", user1.Id, user2.Id)
}
