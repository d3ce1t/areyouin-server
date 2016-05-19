package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sort"
	"strings"
)

func NewShell(server *Server) *Shell {
	return &Shell{server: server}
}

type Shell struct {
	server   *Server
	welcome  string
	prompt   string
	commands map[string]Command
	io       io.ReadWriter
}

type Command func([]string)

func (shell *Shell) StartTermSSH() {

	defer func() {
		if r := recover(); r != nil {
			log.Println("StartTermSSH Error:", r)
		}
	}()

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() == "admin" && string(pass) == "admin" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := ioutil.ReadFile("cert/server_rsa")
	if err != nil {
		panic("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		panic("failed to listen for connection")
	}

	defer listener.Close()

	// Manage incoming connections
	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("SSH Terminal: failed to accept incoming connection (%v)\n", err)
		}
		shell.manageSshSession(nConn, config)
		log.Println("Waiting for a new SSH connection...")
	}
}

func (shell *Shell) manageSshSession(nConn net.Conn, config *ssh.ServerConfig) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Shell Session Error:", r)
			nConn.Close()
		}
	}()

	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	serverConn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	defer serverConn.Close()

	if err != nil {
		panic("failed to handshake")
	}
	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	var newChannel ssh.NewChannel

	for newChannel = range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		break
	}

	channel, requests, err := newChannel.Accept()
	defer channel.Close()

	if err != nil {
		panic("could not accept channel.")
	}

	// Sessions have out-of-band requests such as "shell",
	// "pty-req" and "env".  Here we handle only the
	// "shell" request.
	go func(in <-chan *ssh.Request) {
		for req := range in {
			ok := false
			switch req.Type {
			case "shell":
				ok = true
				if len(req.Payload) > 0 {
					// We don't accept any
					// commands, only the
					// default shell.
					ok = false
				}
			}
			req.Reply(ok, nil)
		}
	}(requests)

	shell.Execute(channel)
}

// Shell wrapper to manage errors
func (shell *Shell) Execute(channel io.ReadWriter) {

	shell.io = channel

	if shell.server.Testing {
		fmt.Fprint(shell.io, "----------------------------------------\n")
		fmt.Fprint(shell.io, "! WARNING WARNING WARNING              !\n")
		fmt.Fprint(shell.io, "! You have started a testing server    !\n")
		fmt.Fprint(shell.io, "! WARNING WARNING WARNING              !\n")
		fmt.Fprint(shell.io, "----------------------------------------\n")
	}

	shell.welcome = "Welcome to AreYouIN server shell\n"
	shell.prompt = "areyouin$>"
	shell.init()
	fmt.Fprintf(channel, "\n%s\n\n", shell.welcome)
	exit := false

	for !exit {
		exit = shell.executeShell()
	}

	fmt.Fprintln(channel, "Good bye")
	log.Println("Shell session terminated")
}

func (shell *Shell) init() {
	shell.commands = map[string]Command{
		"help":             shell.help,
		"list_sessions":    shell.listSessions,
		"list_users":       shell.listUserAccounts,
		"delete_user":      shell.deleteUser,
		"send_auth_error":  shell.sendAuthError,
		"send_msg":         shell.sendMsg,
		"close_session":    shell.closeSession,
		"show_user":        shell.showUser,
		"ping":             shell.pingClient,
		"reset_picture":    shell.resetPicture,
		"create_fake_user": shell.createFakeUser,
		"make_friends":     shell.makeFriends,
		"fix_database":     shell.fixDatabase,
		"change_user_password": shell.changeUserPassword,
	}
}

func (shell *Shell) executeShell() (exit bool) {

	// Defer recovery
	defer func() {
		if r := recover(); r != nil {

			err, ok := r.(error)

			if ok {
				if err == io.EOF {
					exit = true
				} else {
					exit = false
					fmt.Fprintf(shell.io, "Error: %v\r\n", err)
				}
			} else {
				exit = true
			}
			log.Printf("Shell Error: %v\n", err)
		}
	}()

	var args []string
	in := bufio.NewReader(shell.io)

	for { // Execute until main goroutine finish
		// Show prompt
		fmt.Fprint(shell.io, shell.prompt+" ")

		// Read command
		line, err := in.ReadString('\n')
		manageShellError(err)
		line = strings.TrimSpace(line)
		args = strings.Split(line, " ")

		if args[0] == "exit" {
			return true
		}

		if command, ok := shell.commands[args[0]]; ok {
			command(args)
		} else {
			fmt.Fprintf(shell.io, "Command %s does not exist\r\n", args[0])
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

	keys := make([]string, 0, len(shell.commands))

	for k, _ := range shell.commands {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, str := range keys {
		fmt.Fprintf(shell.io, "- %v\n", str)
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
