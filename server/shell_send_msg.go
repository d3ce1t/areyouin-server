package main

import (
	"fmt"
	gcm "github.com/google/go-gcm"
	proto "peeple/areyouin/protocol"
	"strconv"
)

// send_msg client
func (shell *Shell) sendMsg(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	if len(args) < 2 {
		manageShellError(ErrShellInvalidArgs)
	}

	server := shell.server
	userDAO := server.NewUserDAO()

	user_account, err := userDAO.Load(user_id)
	manageShellError(err)

	text_message := args[2]
	for i := 3; i < len(args); i++ {
		text_message += " " + args[i]
	}

	iid_token := user_account.IIDtoken
	gcm_message := gcm.HttpMessage{
		To:         iid_token,
		TimeToLive: 3600,
		Data: gcm.Data{
			"msg_type": uint8(proto.M_INVITATION_RECEIVED),
			"event_id": 0,
			"body":     text_message,
		},
	}

	fmt.Fprintf(shell.io, "Send Message %v\n", text_message)

	response, err := gcm.SendHttp(GCM_API_KEY, gcm_message)
	manageShellError(err)
	fmt.Fprintf(shell.io, "Response %v\n", response)
}
