package main

import (
  "fmt"
  "os"
  "peeple/areyouin/dao"
  "peeple/areyouin/model"
  proto "peeple/areyouin/protocol"
  imgserv "peeple/areyouin/images_server"
  "log"
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

  session := dao.NewSession(db_keyspace, db_address)
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

	server.RegisterCallback(proto.M_PING, onPing)
	server.RegisterCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
	server.RegisterCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
	server.RegisterCallback(proto.M_USER_AUTH, onUserAuthentication)
	server.RegisterCallback(proto.M_GET_ACCESS_TOKEN, onNewAccessToken)
	server.RegisterCallback(proto.M_CREATE_EVENT, onCreateEvent)
	server.RegisterCallback(proto.M_CANCEL_EVENT, onCancelEvent)
	server.RegisterCallback(proto.M_INVITE_USERS, onInviteUsers)
	server.RegisterCallback(proto.M_CONFIRM_ATTENDANCE, onConfirmAttendance)
	server.RegisterCallback(proto.M_GET_USER_FRIENDS, onGetUserFriends)
	server.RegisterCallback(proto.M_GET_USER_ACCOUNT, onGetUserAccount)
	server.RegisterCallback(proto.M_CHANGE_PROFILE_PICTURE, onChangeProfilePicture)
	server.RegisterCallback(proto.M_CLOCK_REQUEST, onClockRequest)
	server.RegisterCallback(proto.M_IID_TOKEN, onIIDTokenReceived)
	server.RegisterCallback(proto.M_CHANGE_EVENT_PICTURE, onChangeEventPicture)
	server.RegisterCallback(proto.M_SYNC_GROUPS, onSyncGroups)
	server.RegisterCallback(proto.M_GET_GROUPS, onGetGroups)
	server.RegisterCallback(proto.M_LIST_PRIVATE_EVENTS, onListPrivateEvents)
	server.RegisterCallback(proto.M_HISTORY_PRIVATE_EVENTS, onListEventsHistory)
	server.RegisterCallback(proto.M_CREATE_FRIEND_REQUEST, onFriendRequest)
	server.RegisterCallback(proto.M_GET_FRIEND_REQUESTS, onListFriendRequests)
	server.RegisterCallback(proto.M_CONFIRM_FRIEND_REQUEST, onConfirmFriendRequest)

  // Create images HTTP server and start
	if !maintenanceMode {
		images_server := imgserv.NewServer(session, model)
		go images_server.Run()
	}

	// Create shell and start listening in 2022 tcp port
	shell := NewShell(server)
	go shell.StartTermSSH()

	// start server loop
	server.Run()
}
