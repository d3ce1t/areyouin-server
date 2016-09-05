package shell

import (
	"fmt"
	"peeple/areyouin/utils"
	"time"
)

// list_sessions
type listSessionsCmd struct {
}

func (c *listSessionsCmd) Exec(shell *Shell, args []string) {

	activeSessions, err := shell.model.Accounts.GetActiveSessions(time.Now())
	manageShellError(err)

	for _, activeSession := range activeSessions {
		fmt.Fprintf(shell, "- %v %v\n", activeSession.UserID, utils.UnixMillisToTime(activeSession.LastTime))
	}
}
