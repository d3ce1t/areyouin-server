package shell

// send_auth_error user_id
/*func sendAuthError(shell *Shell, args []string) {

	user_id, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	if session, ok := server.sessions.Get(user_id); ok {
		session.Write(session.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD))
		fmt.Fprintln(shell.io, "Send invalid user or password")
	}
}*/
