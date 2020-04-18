package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/d3ce1t/areyouin-server/shell"

	"golang.org/x/crypto/ssh"
)

func (server *Server) startShell() {

	defer func() {
		if r := recover(); r != nil {
			log.Println("StartSSHTerm Error:", r)
		}
	}()

	config := loadConfig()

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	addr := fmt.Sprintf("%v:%v", server.Config.SSHListenAddress(), server.Config.SSHListenPort())
	listener, err := net.Listen("tcp", addr)
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
		server.manageSSHSession(nConn, config)
		log.Println("Waiting for a new SSH connection...")
	}
}

func (server *Server) manageSSHSession(nConn net.Conn, config *ssh.ServerConfig) {

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

	sh := shell.NewShell(server, server.Model, channel, globalConfig)
	if server.Config.ShowTestModeWarning() {
		sh.OnStart = func(shell *shell.Shell) {
			fmt.Fprint(shell, "------------------------------------------\n")
			fmt.Fprint(shell, "! WARNING WARNING WARNING                !\n")
			fmt.Fprint(shell, "! You have connected to a testing server !\n")
			fmt.Fprint(shell, "! WARNING WARNING WARNING                !\n")
			fmt.Fprint(shell, "------------------------------------------\n")
		}
	}
	sh.Run()
}

func loadConfig() *ssh.ServerConfig {

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

	return config
}
