package shell

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"peeple/areyouin/model"
	"strings"
)

type Command func(*Shell, []string)

type Shell struct {
	io.ReadWriter
	welcome  string
	prompt   string
	commands map[string]Command
	model    *model.AyiModel
	OnStart  func(*Shell)
}

func NewShell(model *model.AyiModel, io io.ReadWriter) *Shell {
	shell := &Shell{
		welcome:    "Welcome to AreYouIN server shell\n",
		prompt:     "areyouin$>",
		model:      model,
		ReadWriter: io}
	shell.init()
	return shell
}

// Shell wrapper to manage errors
func (s *Shell) Run() {

	if s.OnStart != nil {
		s.OnStart(s)
	}

	/*if shell.server.Testing {
		fmt.Fprint(shell.io, "------------------------------------------\n")
		fmt.Fprint(shell.io, "! WARNING WARNING WARNING                !\n")
		fmt.Fprint(shell.io, "! You have connected to a testing server !\n")
		fmt.Fprint(shell.io, "! WARNING WARNING WARNING                !\n")
		fmt.Fprint(shell.io, "------------------------------------------\n")
	}*/

	fmt.Fprintf(s, "\n%s\n\n", s.welcome)
	exit := false

	for !exit {
		exit = s.executeShell()
	}

	fmt.Fprintln(s, "Good bye")
	log.Println("Shell session terminated")
}

func (s *Shell) init() {
	s.commands = map[string]Command{
		"help": help,
		//"list_sessions":        listSessions,
		"list_users":  listUserAccounts,
		"delete_user": deleteUser,
		//"send_auth_error":      sendAuthError,
		//"send_msg":             sendMsg,
		//"close_session":        closeSession,
		"show_user": showUser,
		//"ping":                 pingClient,
		"reset_picture":        resetPicture,
		"create_fake_user":     createFakeUser,
		"make_friends":         makeFriends,
		"fix_database":         fixDatabase,
		"change_user_password": changeUserPassword,
	}
}

func (s *Shell) executeShell() (exit bool) {

	// Defer recovery
	defer func() {
		if r := recover(); r != nil {

			err, ok := r.(error)

			if ok {
				if err == io.EOF {
					exit = true
				} else {
					exit = false
					fmt.Fprintf(s, "Error: %v\r\n", err)
				}
			} else {
				exit = true
			}
			log.Printf("Shell Error: %v\n", err)
		}
	}()

	var args []string
	in := bufio.NewReader(s)

	// Execute until main goroutine finish

	for {
		// Show prompt
		fmt.Fprint(s, s.prompt+" ")

		// Read command
		line, err := in.ReadString('\n')
		manageShellError(err)
		line = strings.TrimSpace(line)
		args = strings.Split(line, " ")

		if args[0] == "exit" {
			return true
		}

		if command, ok := s.commands[args[0]]; ok {
			command(s, args)
		} else {
			fmt.Fprintf(s, "Command %s does not exist\r\n", args[0])
		}

	} // Loop
}

func manageShellError(err error) {
	if err != nil {
		panic(err)
	}
}

func ff(text interface{}, lenght int) string {
	s := fmt.Sprintf("%v", text)
	if len(s) > lenght {
		s = s[:lenght]
	}
	return s
}

func rp(str string, lenght int) string {
	var repeat string
	for i := 0; i < lenght; i++ {
		repeat += str
	}
	return repeat
}
