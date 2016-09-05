package shell

import (
	"flag"
	"fmt"
	"strings"
)

// send_msg client
type sendNotificationCmd struct {
}

func (c *sendNotificationCmd) Exec(shell *Shell, args []string) {

	var userID int64

	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	cmd.SetOutput(shell)
	cmd.Usage = func() {
		fmt.Fprintf(shell, "Usage of %s:\n", args[0])
		cmd.PrintDefaults()
	}

	cmd.Int64Var(&userID, "user-id", 0, "ID of the user you want to send a notification")

	cmd.Parse(args[1:])

	if userID == 0 {
		cmd.Usage()
		return
	}

	var message string

	for _, arg := range cmd.Args() {
		message += " " + arg
	}

	message = strings.TrimSpace(message)

	if message == "" {
		cmd.Usage()
		return
	}

	fmt.Fprintf(shell, "Send message to %v\n%v\n", userID, message)

	/*user_id, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	if len(args) < 2 {
		manageShellError(ErrShellInvalidArgs)
	}

	server := shell.server
	userDAO := dao.NewUserDAO(server.DbSession)
	user_account, err := userDAO.Load(user_id)
	manageShellError(err)

	sendGcmDataAvailableNotification(user_account.Id, user_account.IIDtoken, 3600)
	fmt.Fprintf(shell.io, "Message Sent\n")*/
}

/*func (c *sendNotificationCmd) createNotification(title string, message string) *gcm.Notification {

	bodyArgs, _ := json.Marshal([]string{friendName})

	notification := &gcm.Notification{
		TitleLocKey: "notification.friend.new.title",
		BodyLocKey:  "notification.friend.new.body",
		BodyLocArgs: string(bodyArgs),
		Icon:        "icon_notification_25dp", // Android only (drawable name)
		Sound:       "default",
		Color:       "#009688", // Android only
	}

	return notification
}*/
