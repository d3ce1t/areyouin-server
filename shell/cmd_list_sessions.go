package shell

import (
	"fmt"
	"time"

	"github.com/d3ce1t/areyouin-server/utils"
)

// list_sessions
type listSessionsCmd struct {
}

func (c *listSessionsCmd) Exec(shell *Shell, args []string) {

	activeSessions, err := shell.model.Accounts.GetActiveSessions(time.Now())
	manageShellError(err)

	for _, activeSession := range activeSessions {
		fmt.Fprintf(shell, "- %v %v\n", activeSession.UserID, utils.MillisToTimeUTC(activeSession.LastTime))
	}
}
