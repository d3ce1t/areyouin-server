package shell

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/model"
	"strings"
)

type Command interface {
	Exec(shell *Shell, args []string)
}

type Shell struct {
	io.ReadWriter
	welcome  string
	prompt   string
	commands map[string]Command
	model    *model.AyiModel
	server   api.Server
	OnStart  func(*Shell)
	config   api.Config
}

func NewShell(server api.Server, model *model.AyiModel, io io.ReadWriter, cfg api.Config) *Shell {
	shell := &Shell{
		welcome:    "Welcome to AreYouIN server shell\n",
		prompt:     "areyouin$>",
		model:      model,
		server:     server,
		config:     cfg,
		ReadWriter: io}

	shell.init()
	return shell
}

// Shell wrapper to manage errors
func (s *Shell) Run() {

	if s.OnStart != nil {
		s.OnStart(s)
	}

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
		"help":          new(helpCmd),
		"list_sessions": new(listSessionsCmd),
		"list_users":    new(listUsersCmd),
		"delete_user":   new(deleteUserCmd),
		//"send_auth_error":      sendAuthError,
		"notify": new(sendNotificationCmd),
		//"close_session":        closeSession,
		"show_user": new(showUserCmd),
		//"ping":                 pingClient,
		"reset_picture":    new(resetPictureCmd),
		"create_fake_user": new(createFakeUserCmd),
		//"make_friends":     makeFriends,
		//"fix_database":         fixDatabase,
		"change_user_password": new(changeUserPasswordCmd),
		"version":              new(versionCmd),
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
			command.Exec(s, args)
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

func ff(text interface{}, length int) string {
	s := fmt.Sprintf("%v", text)
	if len(s) > length {
		s = s[:length]
	}
	return s
}

func rp(str string, length int) string {
	var repeat string
	for i := 0; i < length; i++ {
		repeat += str
	}
	return repeat
}
