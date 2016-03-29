package main

import (
	"errors"
	"fmt"
)

// fix_database
func (shell *Shell) fixDatabase(args []string) {
	switch args[1] {
	case "--event":
		shell.fixEvent()
	case "--import-old-friends-table":
		shell.importFriendsFromOldFormat()
	}
}

// Fill evet.inbox_position with event.start_date if the first one is not set
func (shell *Shell) fixEvent() {

	server := shell.server
	stmt_select := `SELECT event_id, start_date, inbox_position FROM event`
	stmt_update := `UPDATE event SET inbox_position = ? WHERE event_id = ?`

	var event_id uint64
	var start_date int64
	var inbox_position int64

	rows_processed := 0
	fixes := 0

	iter := server.DbSession().Query(stmt_select).Iter()

	for iter.Scan(&event_id, &start_date, &inbox_position) {

		rows_processed++

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

		if rows_processed%250 == 0 {
			fmt.Fprintf(shell.io, "Progress: (fixes: %v, rows_processed: %v)\n",
				fixes, rows_processed)
		}

	}

	manageShellError(iter.Close())

	fmt.Fprintf(shell.io, "Completed: (fixes: %v, rows_processed: %v)\n",
		fixes, rows_processed)
}

// Copy user_friends to friends_by_user. It only takes into account ALL_CONTACTS group.
func (shell *Shell) importFriendsFromOldFormat() {

	server := shell.server
	stmt_select := `SELECT user_id, group_id, friend_id, name, picture_digest FROM user_friends`
	stmt_update := `INSERT INTO friends_by_user (user_id, friend_id, friend_name, picture_digest)
	 	VALUES (?, ?, ?, ?)`

	var user_id uint64
	var group_id int32
	var friend_id int64
	var name string
	var picture_digest []byte

	rows_processed := 0
	fixes := 0

	iter := server.DbSession().Query(stmt_select).Iter()

	for iter.Scan(&user_id, &group_id, &friend_id, &name, &picture_digest) {

		rows_processed++

		if group_id != 0 {
			continue
		}

		q := server.DbSession().Query(stmt_update, user_id, friend_id, name, picture_digest)
		if err := q.Exec(); err != nil {
			manageShellError(
				errors.New(
					fmt.Sprintf("Error %v (user_id: %v, friend_id: %v, fixes: %v, rows_processed: %v)",
						err.Error(), user_id, friend_id, fixes, rows_processed)))
		}

		fixes++

		if rows_processed%250 == 0 {
			fmt.Fprintf(shell.io, "Progress: (fixes: %v, rows_processed: %v)\n",
				fixes, rows_processed)
		}
	}

	manageShellError(iter.Close())

	fmt.Fprintf(shell.io, "Completed: (fixes: %v, rows_processed: %v)\n",
		fixes, rows_processed)
}
