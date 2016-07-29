package shell

// list_sessions
/*func listSessions(shell *Shell, args []string) {

	server := shell.server

	keys := server.sessions.Keys()

	for _, k := range keys {
		session, _ := server.sessions.Get(k)
		fmt.Fprintf(shell.io, "- %v %v (protocolVersion=%v, platform=%v, clientVersion=%v)\n", k, session.Conn.RemoteAddr().String(),
			session.ProtocolVersion, session.Platform, session.ClientVersion)
	}
}*/
