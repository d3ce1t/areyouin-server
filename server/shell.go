package main

import (
	"bufio"
	"fmt"
	"os"
	core "peeple/areyouin/common"
	proto "peeple/areyouin/protocol"
	"sort"
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

	shell.welcome = "Welcome to AreYouIN server shell"
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
		"list_users":      shell.listUserAccounts,
		"delete_user":     shell.deleteUser,
		"send_auth_error": shell.sendAuthError,
		"show_user":       shell.showUser,
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

	keys := make([]string, 0, len(shell.commands))

	for k, _ := range shell.commands {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, str := range keys {
		fmt.Printf("- %v\n", str)
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

// list_users
func (shell *Shell) listUserAccounts(args []string) {

	server := shell.Server
	dao := server.NewUserDAO()
	users, err := dao.LoadAllUsers()
	manageShellError(err)

	fmt.Println(rp("-", 105))
	fmt.Printf("| S | %-17s | %-15s | %-40s | %-16s |\n", "Id", "Name", "Email", "Last connection")
	fmt.Println(rp("-", 105))

	for _, user := range users {
		status_info := " "
		if valid, err := dao.CheckValidCredentials(user.Id, user.Email, user.Fbid); err != nil {
			status_info = "?"
		} else if !valid {
			status_info = "E"
		}

		fmt.Printf("| %v | %-17v | %-15v | %-40v | %-16v |\n", status_info, ff(user.Id, 17), ff(user.Name, 15), ff(user.Email, 40), ff(core.UnixMillisToTime(user.LastConnection), 16))
	}
	fmt.Println(rp("-", 105))

	fmt.Println("Num. Users:", len(users))
}

// show_user
func (shell *Shell) showUser(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.Server
	dao := server.NewUserDAO()
	user, err := dao.Load(user_id)
	manageShellError(err)

	valid_user, _ := user.IsValid()
	valid_account, err := dao.CheckValidAccount(user_id, true)

	if err != nil {
		fmt.Println("Error checking account:", err)
	}

	account_status := ""
	if !valid_user || !valid_account {
		account_status = "¡¡¡INVALID STATUS!!!"
	}

	fmt.Println("---------------------------------")
	fmt.Printf("User details (%v)\n", account_status)
	fmt.Println("---------------------------------")
	fmt.Println("UserID:", user.Id)
	fmt.Println("Name:", user.Name)
	fmt.Println("Email:", user.Email)
	fmt.Println("Email Verified:", user.EmailVerified)
	fmt.Println("Created at:", core.UnixMillisToTime(user.CreatedDate))
	fmt.Println("Last connection:", core.UnixMillisToTime(user.LastConnection))
	fmt.Println("Authtoken:", user.AuthToken)
	fmt.Println("Fbid:", user.Fbid)
	fmt.Println("Fbtoken:", user.Fbtoken)

	fmt.Println("---------------------------------")
	fmt.Println("E-mail credentials")
	fmt.Println("---------------------------------")

	if email, err := dao.LoadEmailCredential(user.Email); err == nil {
		fmt.Println("E-mail:", email.Email == user.Email)
		if email.Password == core.EMPTY_ARRAY_32B || email.Salt == core.EMPTY_ARRAY_32B {
			fmt.Println("No password set")
		} else {
			fmt.Printf("Password: %x\n", email.Password)
			fmt.Printf("Salt: %x\n", email.Salt)
		}
		fmt.Println("UserID Match:", email.UserId == user.Id)
	} else {
		fmt.Println("Error:", err)
	}

	fmt.Println("---------------------------------")
	fmt.Println("Facebook credentials")
	fmt.Println("---------------------------------")

	if user.HasFacebookCredentials() {
		facebook, err := dao.LoadFacebookCredential(user.Fbid)
		if err == nil {
			fmt.Println("Fbid:", facebook.Fbid == user.Fbid)
			fmt.Println("Fbtoken:", facebook.Fbtoken == user.Fbtoken)
			fmt.Println("UserID Match:", facebook.UserId == user.Id)
		} else {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("There aren't credentials")
	}
	fmt.Println("---------------------------------")

	if account_status != "" {
		fmt.Printf("\nACCOUNT INFO: %v\n", account_status)
	}
}

// delete_user $user_id --force
func (shell *Shell) deleteUser(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.Server
	dao := server.NewUserDAO()
	user, err := dao.Load(user_id)
	manageShellError(err)

	if len(args) == 2 {
		err = dao.Delete(user)

		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("Try command:")
			fmt.Printf("\tdelete_user %d --force\n", user_id)
			return
		}
	} else if len(args) > 2 {

		// Try remove user account
		if err := dao.DeleteUserAccount(user_id); err != nil {
			fmt.Println("Removing user account error:", err)
		} else {
			fmt.Println("User account removed")
		}

		// Try remove e-mail credential
		if err := dao.DeleteEmailCredentials(user.Email); err != nil {
			fmt.Println("Removing e-mail credential error:", err)
		} else {
			fmt.Println("E-mail credential removed")
		}

		// Try remove facebook credential
		if err := dao.DeleteFacebookCredentials(user.Fbid); err != nil {
			fmt.Println("Removing facebook credential error:", err)
		} else {
			fmt.Println("Facebook credential removed")
		}

	}

	fmt.Printf("User with id %d has been removed\n", user_id)
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

	server := shell.Server
	if session, ok := server.sessions[user_id]; ok {
		for i := uint64(0); i < repeat_times; i++ {
			ping_msg := proto.NewMessage().Ping().Marshal()
			session.WriteReply(ping_msg)
		}
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
