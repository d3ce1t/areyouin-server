package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/d3ce1t/areyouin-server/api"
	"github.com/d3ce1t/areyouin-server/cqldao"
	imgserv "github.com/d3ce1t/areyouin-server/images_server"
	"github.com/d3ce1t/areyouin-server/model"
	proto "github.com/d3ce1t/areyouin-server/protocol"
)

// Server global configuration
var globalConfig api.Config

func main() {

	fmt.Printf("AreYouIN Server version %v Build %v\n", SERVER_VERSION, BUILD_TIME)

	// Load config from file
	cfg, err := loadConfigFromFile("areyouin.yaml")
	if err != nil {
		log.Fatalf("Couldn't load areyouin.yaml config file: %v", err)
	}

	globalConfig = cfg

	// Process args
	var bootStrap bool

	flag.BoolVar(&cfg.data.MaintenanceMode, "enable-maintenance", cfg.data.MaintenanceMode, "Enable maintenance mode")
	flag.BoolVar(&bootStrap, "bootstrap", false, "Bootstrap server")
	flag.Parse()

	if bootStrap {
		cfg.data.MaintenanceMode = true
	}

	// Initialisation

	if cfg.ShowTestModeWarning() {
		fmt.Println("----------------------------------------")
		fmt.Println("! WARNING WARNING WARNING              !")
		fmt.Println("! You have started a testing server    !")
		fmt.Println("! WARNING WARNING WARNING              !")
		fmt.Println("----------------------------------------")
	}

	// Connect to database

	session := cqldao.NewSession(cfg.DbKeyspace(), cfg.DbCQLVersion(), cfg.DbAddress()...)
	err = session.Connect()
	for err != nil {
		log.Println(err)
		time.Sleep(5 * time.Second)
		err = session.Connect()
	}

	log.Println("Connected to Cassandra successfully")

	// Create model and init server
	model := model.New(session, "default")
	server := NewServer(session, model, cfg)

	if !cfg.MaintenanceMode() {

		// Register callbacks

		server.registerCallback(proto.M_PING, onPing)
		server.registerCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
		server.registerCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
		server.registerCallback(proto.M_USER_AUTH, onUserAuthentication)
		server.registerCallback(proto.M_GET_ACCESS_TOKEN, onNewAccessToken)
		server.registerCallback(proto.M_CREATE_EVENT, onCreateEvent)
		server.registerCallback(proto.M_MODIFY_EVENT, onModifyEvent)
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
		server.registerCallback(proto.M_READ_EVENT, onReadEvent)
		server.registerCallback(proto.M_LIST_PRIVATE_EVENTS, onListPrivateEvents)
		server.registerCallback(proto.M_HISTORY_PRIVATE_EVENTS, onListEventsHistory)
		server.registerCallback(proto.M_CREATE_FRIEND_REQUEST, onFriendRequest)
		server.registerCallback(proto.M_GET_FRIEND_REQUESTS, onListFriendRequests)
		server.registerCallback(proto.M_CONFIRM_FRIEND_REQUEST, onConfirmFriendRequest)
		server.registerCallback(proto.M_GET_FACEBOOK_FRIENDS, onGetFacebookFriends)
		server.registerCallback(proto.M_USER_LINK_ACCOUNT, onLinkAccount)
		server.registerCallback(proto.M_IMPORT_FACEBOOK_FRIENDS, onImportFacebookFriends)
		server.registerCallback(proto.M_SET_FACEBOOK_ACCESS_TOKEN, onSetFacebookAccessToken)

		// Create images HTTP server and start
		imagesServer := imgserv.NewServer(session, model, cfg)
		go imagesServer.Run()

	}

	// Create shell and start listening
	go server.startShell()

	// start server loop
	server.run(bootStrap)
}
