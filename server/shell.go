package main

import (
	"bufio"
	"fmt"
	"os"
	proto "peeple/areyouin/protocol"
	"strconv"
	"strings"
)

type Shell struct {
	Server   *Server
	welcome  string
	prompt   string
	commands map[string]Command
}

type Command func([]string)

// Shell wrapper to manage errors
func (shell *Shell) Execute() {

	shell.welcome = "Welcome to AreYouIN server SHELL"
	shell.prompt = "areyouin$>"
	shell.init()

	fmt.Printf("\n%s\n\n\n", shell.welcome)
	exit := false

	for !exit {
		exit = shell.executeShell()
	}

	fmt.Println("Good bye")
}

func (shell *Shell) init() {
	shell.commands = map[string]Command{
		"help":            shell.help,
		"list_sessions":   shell.listSessions,
		"send_auth_error": shell.sendAuthError,
		"ping":            shell.pingClient,
	}
}

func (shell *Shell) executeShell() (exit bool) {

	// Defer recovery
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println("Error:", err)
			exit = false
		}
	}()

	var args []string

	in := bufio.NewReader(os.Stdin)

	for { // Execute until main goroutine finish
		// Show prompt
		fmt.Print(shell.prompt + " ")

		// Read command
		line, err := in.ReadString('\n')
		manageShellError(err)
		line = strings.TrimSpace(line)
		args = strings.Split(line, " ")

		if command, ok := shell.commands[args[0]]; ok {
			command(args)
		} else {
			fmt.Printf("Command %s does not exist\n", args[0])
		}
	} // Loop
}

func manageShellError(err error) {
	if err != nil {
		panic(err)
	}
}

// help
func (shell *Shell) help(args []string) {
	for k, _ := range shell.commands {
		fmt.Printf("- %v\n", k)
	}
}

// send_auth_error user_id
func (shell *Shell) sendAuthError(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.Server
	if session, ok := server.sessions[user_id]; ok {
		sendAuthError(session)
	}
}

// list_sessions
func (shell *Shell) listSessions(args []string) {

	server := shell.Server

	for k, session := range server.sessions {
		fmt.Printf("- %v %v\n", k, session)
	}
}

// ping client
func (shell *Shell) pingClient(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.Server
	if session, ok := server.sessions[user_id]; ok {
		ping_msg := proto.NewMessage().Ping().Marshal()
		session.WriteReply(ping_msg)
	}
}
