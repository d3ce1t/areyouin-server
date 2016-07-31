package shell

import (
	"fmt"
	"peeple/areyouin/utils"
	"time"
)

// list_sessions
func listSessions(shell *Shell, args []string) {

	activeSessions, err := shell.model.Accounts.GetActiveSessions(time.Now())
	manageShellError(err)

	for _, activeSession := range activeSessions {
		fmt.Fprintf(shell, "- %v %v\n", activeSession.UserID, utils.UnixMillisToTime(activeSession.LastTime))
	}
}
