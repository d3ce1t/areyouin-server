package main

import (
	"errors"
	"fmt"
)

// fix_database
func (shell *Shell) fixDatabase(args []string) {

	server := shell.server
	stmt_select := `SELECT event_id, start_date, inbox_position FROM event`
	stmt_update := `UPDATE event SET inbox_position = ? WHERE event_id = ?`
	iter := server.DbSession().Query(stmt_select).Iter()

	var event_id uint64
	var start_date int64
	var inbox_position int64

	rows_processed := 0
	fixes := 0

	for iter.Scan(&event_id, &start_date, &inbox_position) {

		if inbox_position == 0 {
			q := server.DbSession().Query(stmt_update, start_date, event_id)
			if err := q.Exec(); err != nil {
				manageShellError(
					errors.New(
						fmt.Sprintf("Error %v (event_id: %v, fixes: %v, rows_processed: %v)",
							err.Error(), event_id, fixes, rows_processed)))
			}

			fixes++
		}

		rows_processed++

		if rows_processed%250 == 0 {
			fmt.Fprintf(shell.io, "Progress: (fixes: %v, rows_processed: %v)\n",
				fixes, rows_processed)
		}

	}

	manageShellError(iter.Close())

	fmt.Fprintf(shell.io, "Completed: (fixes: %v, rows_processed: %v)\n",
		fixes, rows_processed)
}
