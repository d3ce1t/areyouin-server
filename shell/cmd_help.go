package shell

import (
	"fmt"
	"sort"
)

type helpCmd struct {
}

// help
func (c *helpCmd) Exec(shell *Shell, args []string) {

	keys := make([]string, 0, len(shell.commands))

	for k, _ := range shell.commands {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, str := range keys {
		fmt.Fprintf(shell, "- %v\n", str)
	}
}
