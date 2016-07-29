package main

import (
	"fmt"
	"log"
	"os"
	"peeple/areyouin/cqldao"
	imgserv "peeple/areyouin/images_server"
	"peeple/areyouin/model"
	proto "peeple/areyouin/protocol"
	"peeple/areyouin/server/shell"
	"time"
)

func main() {

	// Flags

	testingMode := true
	maintenanceMode := false

	// Process args

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--enable-maintenance" {
		maintenanceMode = true
	}

	// Initialisation

	var db_keyspace string
	var db_address string

	if !testingMode {
		db_keyspace = "areyouin"
		db_address = "192.168.1.2"
	} else {

		fmt.Println("----------------------------------------")
		fmt.Println("! WARNING WARNING WARNING              !")
		fmt.Println("! You have started a testing server    !")
		fmt.Println("! WARNING WARNING WARNING              !")
		fmt.Println("----------------------------------------")

		db_keyspace = "areyouin"
		db_address = "192.168.1.10"
	}

	session := cqldao.NewSession(db_keyspace, db_address)
	model := model.New(session, "default")

	// Connect to database

	err := session.Connect()

	for err != nil {
		log.Println(err)
		time.Sleep(5 * time.Second)
		err = session.Connect()
	}

	log.Println("Connected to Cassandra successfully")

	// Create and init server

	server := NewServer(session, model)
	server.MaintenanceMode = maintenanceMode
	server.Testing = testingMode

	// Register callbacks

	server.registerCallback(proto.M_PING, onPing)
	server.registerCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
	server.registerCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
	server.registerCallback(proto.M_USER_AUTH, onUserAuthentication)
	server.registerCallback(proto.M_GET_ACCESS_TOKEN, onNewAccessToken)
	server.registerCallback(proto.M_CREATE_EVENT, onCreateEvent)
	server.registerCallback(proto.M_CANCEL_EVENT, onCancelEvent)
	server.registerCallback(proto.M_INVITE_USERS, onInviteUsers)
	server.registerCallback(proto.M_CONFIRM_ATTENDANCE, onConfirmAttendance)
	server.registerCallback(proto.M_GET_USER_FRIENDS, onGetUserFriends)
	server.registerCallback(proto.M_GET_USER_ACCOUNT, onGetUserAccount)
	server.registerCallback(proto.M_CHANGE_PROFILE_PICTURE, onChangeProfilePicture)
	server.registerCallback(proto.M_CLOCK_REQUEST, onClockRequest)
	server.registerCallback(proto.M_IID_TOKEN, onIIDTokenReceived)
	server.registerCallback(proto.M_CHANGE_EVENT_PICTURE, onChangeEventPicture)
	server.registerCallback(proto.M_SYNC_GROUPS, onSyncGroups)
	server.registerCallback(proto.M_GET_GROUPS, onGetGroups)
	server.registerCallback(proto.M_LIST_PRIVATE_EVENTS, onListPrivateEvents)
	server.registerCallback(proto.M_HISTORY_PRIVATE_EVENTS, onListEventsHistory)
	server.registerCallback(proto.M_CREATE_FRIEND_REQUEST, onFriendRequest)
	server.registerCallback(proto.M_GET_FRIEND_REQUESTS, onListFriendRequests)
	server.registerCallback(proto.M_CONFIRM_FRIEND_REQUEST, onConfirmFriendRequest)

	// Create images HTTP server and start
	if !maintenanceMode {
		images_server := imgserv.NewServer(session, model)
		go images_server.Run()
	}

	// Create shell and start listening in 2022 tcp port
	go shell.StartSSH(model)

	// start server loop
	server.run()
}
