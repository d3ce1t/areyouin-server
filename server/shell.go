package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	gcm "github.com/google/go-gcm"
	"golang.org/x/crypto/ssh"
	"image"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	core "peeple/areyouin/common"
	fb "peeple/areyouin/facebook"
	proto "peeple/areyouin/protocol"
	"sort"
	"strconv"
	"strings"
	"unicode"
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

// send_auth_error user_id
func (shell *Shell) sendAuthError(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	if session, ok := server.sessions.Get(user_id); ok {
		sendAuthError(session)
	}
}

// close_session user_id
func (shell *Shell) closeSession(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	if session, ok := server.sessions.Get(user_id); ok {
		session.Exit()
	}
}

// list_sessions
func (shell *Shell) listSessions(args []string) {

	server := shell.server

	keys := server.sessions.Keys()

	for _, k := range keys {
		session, _ := server.sessions.Get(k)
		fmt.Fprintf(shell.io, "- %v %v (%v %v %v)\n", k, session, session.ProtocolVersion, session.Platform, session.ClientVersion)
	}
}

// list_users
func (shell *Shell) listUserAccounts(args []string) {

	server := shell.server
	dao := server.NewUserDAO()
	users, err := dao.LoadAllUsers()
	manageShellError(err)

	fmt.Fprintln(shell.io, rp("-", 105))
	fmt.Fprintf(shell.io, "| S | %-17s | %-15s | %-40s | %-16s |\n", "Id", "Name", "Email", "Last connection")
	fmt.Fprintln(shell.io, rp("-", 105))

	for _, user := range users {
		status_info := " "
		if valid, err := dao.CheckValidAccountObject(user.Id, user.Email, user.Fbid, true); err != nil {
			log.Println("ListUserAccountsError:", err)
			status_info = "?"
		} else if !valid {
			status_info = "E"
		}

		fmt.Fprintf(shell.io, "| %v | %-17v | %-15v | %-40v | %-16v |\n", status_info, ff(user.Id, 17), ff(user.Name, 15), ff(user.Email, 40), ff(core.UnixMillisToTime(user.LastConnection), 16))
	}
	fmt.Fprintln(shell.io, rp("-", 105))

	fmt.Fprintln(shell.io, "Num. Users:", len(users))
}

// show_user
func (shell *Shell) showUser(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	dao := server.NewUserDAO()
	user, err := dao.Load(user_id)
	manageShellError(err)

	valid_user, _ := user.IsValid()
	valid_account, err := dao.CheckValidAccount(user_id, true)

	if err != nil {
		fmt.Fprintln(shell.io, "Error checking account:", err)
	}

	account_status := ""
	if !valid_user || !valid_account {
		account_status = "¡¡¡INVALID STATUS!!!"
	}

	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintf(shell.io, "User details (%v)\n", account_status)
	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintln(shell.io, "UserID:", user.Id)
	fmt.Fprintln(shell.io, "Name:", user.Name)
	fmt.Fprintln(shell.io, "Email:", user.Email)
	fmt.Fprintln(shell.io, "Email Verified:", user.EmailVerified)
	fmt.Fprintln(shell.io, "Created at:", core.UnixMillisToTime(user.CreatedDate))
	fmt.Fprintln(shell.io, "Last connection:", core.UnixMillisToTime(user.LastConnection))
	fmt.Fprintln(shell.io, "Authtoken:", user.AuthToken)
	fmt.Fprintln(shell.io, "Fbid:", user.Fbid)
	fmt.Fprintln(shell.io, "Fbtoken:", user.Fbtoken)

	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintln(shell.io, "E-mail credentials")
	fmt.Fprintln(shell.io, "---------------------------------")

	if email, err := dao.LoadEmailCredential(user.Email); err == nil {
		fmt.Fprintln(shell.io, "E-mail:", email.Email == user.Email)
		if email.Password == core.EMPTY_ARRAY_32B || email.Salt == core.EMPTY_ARRAY_32B {
			fmt.Fprintln(shell.io, "No password set")
		} else {
			fmt.Fprintf(shell.io, "Password: %x\n", email.Password)
			fmt.Fprintf(shell.io, "Salt: %x\n", email.Salt)
		}
		fmt.Fprintln(shell.io, "UserID Match:", email.UserId == user.Id)
	} else {
		fmt.Fprintln(shell.io, "Error:", err)
	}

	fmt.Fprintln(shell.io, "---------------------------------")
	fmt.Fprintln(shell.io, "Facebook credentials")
	fmt.Fprintln(shell.io, "---------------------------------")

	if user.HasFacebookCredentials() {
		facebook, err := dao.LoadFacebookCredential(user.Fbid)
		if err == nil {
			fmt.Fprintln(shell.io, "Fbid:", facebook.Fbid == user.Fbid)
			fmt.Fprintln(shell.io, "Fbtoken:", facebook.Fbtoken == user.Fbtoken)
			fmt.Fprintln(shell.io, "UserID Match:", facebook.UserId == user.Id)
		} else {
			fmt.Fprintln(shell.io, "Error:", err)
		}
	} else {
		fmt.Fprintln(shell.io, "There aren't credentials")
	}
	fmt.Fprintln(shell.io, "---------------------------------")

	if account_status != "" {
		fmt.Fprintf(shell.io, "\nACCOUNT INFO: %v\n", account_status)
	}
}

// delete_user $user_id --force
func (shell *Shell) deleteUser(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	dao := server.NewUserDAO()
	user, err := dao.Load(user_id)
	manageShellError(err)

	if len(args) == 2 {
		err = dao.Delete(user)

		if err != nil {
			fmt.Fprintln(shell.io, "Error:", err)
			fmt.Fprintln(shell.io, "Try command:")
			fmt.Fprintf(shell.io, "\tdelete_user %d --force\n", user_id)
			return
		}
	} else if len(args) > 2 {

		if args[2] != "--force" {
			manageShellError(ErrShellInvalidArgs)
		}

		// Try remove user account
		if err := dao.DeleteUserAccount(user_id); err == nil {
			fmt.Fprintln(shell.io, "User account removed")
		} else {
			fmt.Fprintln(shell.io, "Removing user account error:", err)
		}

		// Try remove e-mail credential
		if user.Email != "" {
			email_credential, err := dao.LoadEmailCredential(user.Email)
			if err == nil && email_credential.UserId == user.Id {
				if err := dao.DeleteEmailCredentials(user.Email); err == nil {
					fmt.Fprintln(shell.io, "E-mail credential removed")
				} else {
					fmt.Fprintln(shell.io, "Removing e-mail credential error:", err)
				}
			}
		}

		// Try remove facebook credential
		if user.Fbid != "" {
			facebook_credential, err := dao.LoadFacebookCredential(user.Fbid)
			if err == nil && facebook_credential.UserId == user.Id {
				if err := dao.DeleteFacebookCredentials(user.Fbid); err == nil {
					fmt.Fprintln(shell.io, "Facebook credential removed")
				} else {
					fmt.Fprintln(shell.io, "Removing facebook credential error:", err)
				}
			}
		}

	}

	fmt.Fprintf(shell.io, "User with id %d has been removed\n", user_id)
}

// ping client
func (shell *Shell) pingClient(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	var repeat_times uint64 = 1

	if len(args) >= 3 {
		repeat_times, err = strconv.ParseUint(args[2], 10, 32)
		manageShellError(err)
	}

	server := shell.server
	if session, ok := server.sessions.Get(user_id); ok {
		for i := uint64(0); i < repeat_times; i++ {
			session.Write(session.NewMessage().Ping())
		}
	}
}

// send_msg client
func (shell *Shell) sendMsg(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	if len(args) < 2 {
		manageShellError(ErrShellInvalidArgs)
	}

	server := shell.server
	userDAO := server.NewUserDAO()

	user_account, err := userDAO.Load(user_id)
	manageShellError(err)

	text_message := args[2]
	for i := 3; i < len(args); i++ {
		text_message += " " + args[i]
	}

	iid_token := user_account.IIDtoken
	gcm_message := gcm.HttpMessage{
		To:         iid_token,
		TimeToLive: 3600,
		Data: gcm.Data{
			"msg_type": uint8(proto.M_INVITATION_RECEIVED),
			"event_id": 0,
			"body":     text_message,
		},
	}

	fmt.Fprintf(shell.io, "Send Message %v\n", text_message)

	response, err := gcm.SendHttp(GCM_API_KEY, gcm_message)
	manageShellError(err)
	fmt.Fprintf(shell.io, "Response %v\n", response)
}

// reset_picture
func (shell *Shell) resetPicture(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	userDAO := server.NewUserDAO()

	// Load user
	user_account, err := userDAO.Load(user_id)
	manageShellError(err)

	// Get profile picture
	fbsession := fb.NewSession(user_account.Fbtoken)
	picture_bytes, err := fb.GetProfilePicture(fbsession)
	manageShellError(err)

	// Decode image
	original_image, _, err := image.Decode(bytes.NewReader(picture_bytes))
	manageShellError(err)

	// Resize image to 512x512
	picture_bytes, err = server.resizeImage(original_image, 512)
	manageShellError(err)

	// Compute digest and prepare image
	digest := sha256.Sum256(picture_bytes)

	picture := &core.Picture{
		RawData: picture_bytes,
		Digest:  digest[:],
	}

	// Save profile Picture
	err = server.saveProfilePicture(user_id, picture)
	manageShellError(err)

	fmt.Fprintf(shell.io, "Picture size %v bytes\n", len(picture_bytes))
}

// create_fake_user
func (shell *Shell) createFakeUser(args []string) {

	// Request random user data
	resp, err := http.Get("https://randomuser.me/api/")
	manageShellError(err)

	defer resp.Body.Close()

	json_data, err := ioutil.ReadAll(resp.Body)
	manageShellError(err)

	// Decode data
	var f interface{}
	err = json.Unmarshal(json_data, &f)
	manageShellError(err)

	m := f.(map[string]interface{})["results"].([]interface{})[0].(map[string]interface{})
	m = m["user"].(map[string]interface{})

	firstName := []rune(m["name"].(map[string]interface{})["first"].(string))
	lastName := []rune(m["name"].(map[string]interface{})["last"].(string))
	firstName[0] = unicode.ToUpper(firstName[0])
	lastName[0] = unicode.ToUpper(lastName[0])
	name := string(firstName) + " " + string(lastName)

	email := m["email"].(string)
	password := "12345" //m["password"].(string)

	pictures := m["picture"].(map[string]interface{})
	picture_url := pictures["large"].(string)

	// Download profile picture
	resp, err = http.Get(picture_url)
	manageShellError(err)

	defer resp.Body.Close()
	picture_bytes, err := ioutil.ReadAll(resp.Body)
	manageShellError(err)

	// Decode image
	original_image, _, err := image.Decode(bytes.NewReader(picture_bytes))
	manageShellError(err)

	// Resize image to 512xauto
	picture_bytes, err = shell.server.resizeImage(original_image, 512)
	manageShellError(err)

	// Create new user account
	user := core.NewUserAccount(shell.server.GetNewID(), name, email, password, "", "", "")
	err = shell.server.createUserAccount(user)
	manageShellError(err)

	// Success
	fmt.Fprint(shell.io, "Account created successfully\n")
	fmt.Fprintf(shell.io, "Name: %v\nEmail: %v\nPassword: %v\nPicture: %v\n",
		name, email, password, picture_url)

	// Save profile Picture
	digest := sha256.Sum256(picture_bytes)

	picture := &core.Picture{
		RawData: picture_bytes,
		Digest:  digest[:],
	}

	err = shell.server.saveProfilePicture(user.Id, picture)
	manageShellError(err)

	fmt.Fprintf(shell.io, "Profile picture changed\n")
}

// make_friends user1 user2
func (shell *Shell) makeFriends(args []string) {

	friend_one_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	friend_two_id, err := strconv.ParseUint(args[2], 10, 64)
	manageShellError(err)

	server := shell.server
	userDAO := server.NewUserDAO()
	friendDAO := server.NewFriendDAO()

	user1, err := userDAO.Load(friend_one_id)
	manageShellError(err)

	user2, err := userDAO.Load(friend_two_id)
	manageShellError(err)

	err = friendDAO.MakeFriends(user1, user2)
	manageShellError(err)

	fmt.Fprintf(shell.io, "%v and %v are now friends\n", user1.Id, user2.Id)
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
