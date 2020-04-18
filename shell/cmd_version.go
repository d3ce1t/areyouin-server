package shell

import "fmt"

// version
type versionCmd struct {
}

func (c *versionCmd) Exec(shell *Shell, args []string) {
	fmt.Fprintf(shell, "Version %v Build %v\n", shell.server.Version(), shell.server.BuildTime())
}
