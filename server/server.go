package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"net"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	fb "peeple/areyouin/facebook"
	imgserv "peeple/areyouin/images_server"
	proto "peeple/areyouin/protocol"
	wh "peeple/areyouin/webhook"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gocql/gocql"
)

const (
	ALL_CONTACTS_GROUP         = 0 // Id for the main friend group of a user
	GCM_API_KEY                = "AIzaSyAf-h1zJCRWNDt-dI3liL1yx4NEYjOq5GQ"
	GCM_MAX_TTL                = 2419200
	PROFILE_PICTURE_MAX_WIDTH  = 512
	PROFILE_PICTURE_MAX_HEIGHT = 512
	EVENT_PICTURE_MAX_WIDTH    = 1280
	EVENT_PICTURE_MAX_HEIGHT   = 720
	EVENT_THUMBNAIL            = 100              // 100 px
	THUMBNAIL_MDPI_SIZE        = 50               // 50 px
	IMAGE_MDPI                 = 160              // 160dpi
	IMAGE_HDPI                 = 1.5 * IMAGE_MDPI // 240dpi
	IMAGE_XHDPI                = 2 * IMAGE_MDPI   // 320dpi
	IMAGE_XXHDPI               = 3 * IMAGE_MDPI   // 480dpi
	IMAGE_XXXHDPI              = 4 * IMAGE_MDPI   // 640dpi
	//MAX_READ_TIMEOUT   = 1 * time.Second
)

// GCM MESSAGES
const (
	GCM_NEW_EVENT_MESSAGE       = 1
	GCM_NEW_FRIEND_MESSAGE      = 2
	GCM_EVENT_CANCELLED_MESSAGE = 3
)

func NewServer() *Server {
	server := &Server{
		DbAddress: "192.168.1.2",
		Keyspace:  "areyouin",
	}
	server.init()
	return server
}

func NewTestServer() *Server {

	fmt.Println("----------------------------------------")
	fmt.Println("! WARNING WARNING WARNING              !")
	fmt.Println("! You have started a testing server    !")
	fmt.Println("! WARNING WARNING WARNING              !")
	fmt.Println("----------------------------------------")

	server := &Server{
		DbAddress: "192.168.1.10",
		Keyspace:  "areyouin",
		Testing:   true,
	}
	server.init()
	return server
}

type Callback func(proto.PacketType, proto.Message, *AyiSession)

type Server struct {
	TLSConfig     *tls.Config
	sessions      *SessionsMap
	task_executor *TaskExecutor
	id_gen_ch     chan uint64
	callbacks     map[proto.PacketType]Callback
	id_generators map[uint16]*core.IDGen
	cluster       *gocql.ClusterConfig
	dbsession     *gocql.Session
	Keyspace      string
	webhook       *wh.WebHookServer
	supportedDpi  []int32
	DbAddress     string
	serialChannel chan func()
	Testing       bool
}

func (s *Server) DbSession() *gocql.Session {
	return s.dbsession
}

// Setup server components
func (s *Server) init() {

	// Init TLS config
	cert, err := tls.LoadX509KeyPair("cert/fullchain.pem", "cert/privkey.pem")
	if err != nil {
		panic(err)
	}

	s.TLSConfig = &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{cert},
		ServerName:   "service.peeple.es",
	}

	// Init sessions holder
	s.sessions = NewSessionsMap()
	s.callbacks = make(map[proto.PacketType]Callback)

	// Serial execution
	s.serialChannel = make(chan func(), 8)
	go func() {
		for f := range s.serialChannel {
			f()
		}
	}()

	// ID generator
	s.id_generators = make(map[uint16]*core.IDGen)
	s.id_gen_ch = make(chan uint64)
	go s.generatorTask(1)

	// Task Executor
	s.task_executor = NewTaskExecutor(s)
	s.task_executor.Start()

	// Supported Screen densities
	s.supportedDpi = []int32{IMAGE_MDPI, IMAGE_HDPI, IMAGE_XHDPI,
		IMAGE_XXHDPI, IMAGE_XXXHDPI}

	// Connect to Cassandra
	s.cluster = gocql.NewCluster(s.DbAddress /*"192.168.1.3"*/)
	s.cluster.Keyspace = s.Keyspace
	s.cluster.Consistency = gocql.LocalQuorum

	for s.connectToDB() != nil {
		time.Sleep(5 * time.Second)
	}

	log.Println("Connected to Cassandra successfully")
}

func (s *Server) connectToDB() error {
	if session, err := s.cluster.CreateSession(); err == nil {
		s.dbsession = session
		return nil
	} else {
		log.Println("Error connecting to cassandra:", err)
		return err
	}
}

func (s *Server) executeSerial(f func()) {
	s.serialChannel <- f
}

func (s *Server) Run() {

	// Start webhook
	s.webhook = wh.New(fb.FB_APP_SECRET)
	s.webhook.RegisterCallback(s.onFacebookUpdate)
	s.webhook.Run()

	// Start up server listener
	listener, err := net.Listen("tcp", ":1822")

	if err != nil {
		panic("Couldn't start listening: " + err.Error())
	}

	defer listener.Close()

	// Main Loop
	for {
		client, err := listener.Accept()

		if err != nil {
			log.Println("Couldn't accept:", err.Error())
			continue
		}

		session := NewSession(client, s)
		go s.handleSession(session)
	}
}

func (s *Server) GetNewID() uint64 {
	return <-s.id_gen_ch
}

func (s *Server) RegisterCallback(command proto.PacketType, f Callback) {
	if s.callbacks == nil {
		s.callbacks = make(map[proto.PacketType]Callback)
	}
	s.callbacks[command] = f
}

func (s *Server) RegisterSession(session *AyiSession) {
	s.executeSerial(func() {
		if oldSession, ok := s.sessions.Get(session.UserId); ok {
			oldSession.WriteSync(oldSession.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD))
			log.Printf("< (%v) SEND INVALID USER OR PASSWORD\n", oldSession)
			oldSession.Exit()
			log.Printf("Closing old session %v for user %v\n", oldSession, oldSession.UserId)
		}
		s.sessions.Put(session.UserId, session)
		log.Printf("Register session %v for user %v\n", session, session.UserId)
	})
}

func (s *Server) UnregisterSession(session *AyiSession) {
	s.executeSerial(func() {
		user_id := session.UserId
		if !session.IsClosed() {
			session.Exit()
		}
		oldSession, ok := s.sessions.Get(user_id)
		if ok && oldSession == session {
			s.sessions.Remove(user_id)
			log.Printf("Unregister session %v for user %v\n", session, user_id)
		}
	})
}

func (s *Server) NewUserDAO() core.UserDAO {
	if s.dbsession == nil || s.dbsession.Closed() {
		s.connectToDB()
	}
	return dao.NewUserDAO(s.dbsession)
}

func (s *Server) NewFriendDAO() core.FriendDAO {
	if s.dbsession == nil || s.dbsession.Closed() {
		s.connectToDB()
	}
	return dao.NewFriendDAO(s.dbsession)
}

func (s *Server) NewEventDAO() core.EventDAO {
	if s.dbsession == nil || s.dbsession.Closed() {
		s.connectToDB()
	}
	return dao.NewEventDAO(s.dbsession)
}

func (s *Server) NewThumbnailDAO() core.ThumbnailDAO {
	if s.dbsession == nil || s.dbsession.Closed() {
		s.connectToDB()
	}
	return dao.NewThumbnailDAO(s.dbsession)
}

func (s *Server) NewAccessTokenDAO() core.AccessTokenDAO {
	if s.dbsession == nil || s.dbsession.Closed() {
		s.connectToDB()
	}
	return dao.NewAccessTokenDAO(s.dbsession)
}

// Private methods
func (server *Server) handleSession(session *AyiSession) {

	defer func() { // session.RunLoop() may throw panic
		if r := recover(); r != nil {
			log.Printf("Session %v Panic: %v\n", session, r)
		}
		session.Exit()
	}()

	log.Println("New connection from", session)

	session.OnRead = func(s *AyiSession, packet *proto.AyiPacket) {
		if err := s.Server.serveMessage(packet, s); err != nil { // may block until writes are performed
			error_code := getNetErrorCode(err, proto.E_OPERATION_FAILED)
			log.Printf("< (%v) ERROR %v: %v\n", session.UserId, error_code, err)
			//log.Printf("Involved Packet: %v\n", packet)
			s.Write(s.NewMessage().Error(packet.Type(), error_code))
		}
	}

	session.OnError = func() func(s *AyiSession, err error) {
		num_errors := 0
		last_error_time := time.Now()
		return func(s *AyiSession, err error) {

			if err == proto.ErrTimeout {
				s.Exit()

				// HACK: Compare error string because there is no ErrTlsXXXXX or alike
			} else if strings.Contains(err.Error(), "tls: first record does not look like a TLS handshake") {
				s.Exit()
			}

			log.Println("Session Error:", err)

			current_time := time.Now()
			if last_error_time.Add(1 * time.Second).Before(current_time) {
				last_error_time = current_time
				num_errors = 1
			} else {
				num_errors++
				if num_errors == 10 {
					s.Exit()
				}
			}
		}
	}()

	session.OnClosed = func(s *AyiSession, peer bool) {
		if s.IsAuth {
			//NOTE: If a user is deleted from user_account while it is still connected,
			// a row in invalid state will be created when updating last connection
			s.Server.NewUserDAO().SetLastConnection(s.UserId, core.GetCurrentTimeMillis())
		}

		s.Server.UnregisterSession(s)

		if peer {
			log.Printf("Session closed by client: %v %v\n", s.UserId, s)
		} else {
			log.Printf("Session closed %v %v\n", s.UserId, s)
		}
	}

	session.RunLoop() // Block here
}

func (s *Server) getIDGenerator(id uint16) *core.IDGen {

	generator, ok := s.id_generators[id]

	if !ok {
		generator = core.NewIDGen(id)
		s.id_generators[id] = generator
	}

	return generator
}

func (s *Server) generatorTask(gid uint16) {

	gen := s.getIDGenerator(gid)

	for {
		newId := gen.GenerateID()
		s.id_gen_ch <- newId
	}
}

func (s *Server) serveMessage(packet *proto.AyiPacket, session *AyiSession) (err error) {

	// Decodes payload. If message does not have payload, ignore that
	message, decode_err := packet.DecodeMessage()
	if err != nil && err != proto.ErrNoPayload {
		return decode_err
	}

	// Defer recovery
	defer func() {
		if r := recover(); r != nil {
			if err_tmp, ok := r.(error); ok {
				err = err_tmp
			} else {
				err = errors.New(fmt.Sprintf("%v", r))
			}
		}
	}()

	// Call function to manage this message
	if f, ok := s.callbacks[packet.Type()]; ok {
		f(packet.Type(), message, session)
	} else {
		err = ErrUnregisteredMessage
	}

	return
}

/*map[
	object:user
	entry:[
				map[uid:100212827024403
						id:100212827024403
						time:1.451828949e+09
						changed_fields:[friends]]
				map[uid:101253640253369
						id:101253640253369
						time:1.451828949e+09
						changed_fields:[friends]]
	 ]
]*/
func (s *Server) onFacebookUpdate(updateInfo *wh.FacebookUpdate) {

	if updateInfo.Object != "user" {
		log.Println("onFacebookUpdate: Invalid update type")
		return
	}

	userDao := s.NewUserDAO()

	for _, entry := range updateInfo.Entries {

		user_id, err := userDao.GetIDByFacebookID(entry.Id)
		if err != nil {
			log.Printf("onFacebookUpdate Error 1 (%v): %v\n", user_id, err)
			continue
		}

		user, err := userDao.Load(user_id)
		if err != nil {
			log.Printf("onFacebookUpdate Error 2 (%v): %v\n", user_id, err)
			continue
		}

		for _, changedField := range entry.ChangedFields {
			if changedField == "friends" {
				s.task_executor.Submit(&ImportFacebookFriends{
					TargetUser: user,
					Fbtoken:    user.Fbtoken,
				})
			}
		} // End inner loop
	} // End outter loop
}

func (s *Server) GetSession(user_id uint64) *AyiSession {
	if session, ok := s.sessions.Get(user_id); ok {
		return session
	} else {
		return nil
	}
}

func (s *Server) GetNewParticipants(participants_id []uint64, event *core.Event) []uint64 {
	result := make([]uint64, 0, len(participants_id))
	for _, id := range participants_id {
		if _, ok := event.Participants[id]; !ok {
			result = append(result, id)
		}
	}
	return result
}

// Insert an event into database, add participants to it and send it to users' inbox.
func (s *Server) PublishEvent(event *core.Event) error {

	eventDAO := s.NewEventDAO()

	if len(event.Participants) <= 1 { // I need more than the only author
		return ErrParticipantsRequired
	}

	if err := eventDAO.InsertEventAndParticipants(event); err != nil {
		return err
	}

	// Author is the last participant. Add it first in order to get the author receive
	// the event to add other participants if something fails
	event.Participants[event.AuthorId].Delivered = core.MessageStatus_CLIENT_DELIVERED
	tmp_error := s.inviteParticipantToEvent(event, event.Participants[event.AuthorId])
	if tmp_error != nil {
		log.Println("PublishEvent Error:", tmp_error)
		return ErrAuthorDeliveryError
	}

	for k, participant := range event.Participants {
		if k != event.AuthorId {
			participant.Delivered = core.MessageStatus_SERVER_DELIVERED
			s.inviteParticipantToEvent(event, participant) // TODO: Implement retry but do not return on error
		}
	}

	return nil
}

func (s *Server) inviteParticipantToEvent(event *core.Event, participant *core.EventParticipant) error {

	eventDAO := s.NewEventDAO()

	if err := eventDAO.InsertEventToUserInbox(participant, event); err != nil {
		log.Println("Coudn't deliver event", event.EventId, err)
		return err
	}

	log.Println("Event", event.EventId, "delivered to user", participant.UserId)
	return nil
}

// This function is used to create a participant list that will be added to an event.
// This event will be published on behalf of the author. By this reason, participants
// can only be current friends of the author. In this code, it is assumed that
// participants are already friends of the author (author has-friend way). However,
// it must be checked if participants have also the author as a friend (friend has-author way)
func (s *Server) loadUserParticipants(author_id uint64, participants_id []uint64) (map[uint64]*core.UserAccount, error, error) {

	var last_warning error
	result := make(map[uint64]*core.UserAccount)
	userDAO := s.NewUserDAO()
	friend_dao := s.NewFriendDAO()

	for _, p_id := range participants_id {

		if ok, err := friend_dao.IsFriend(p_id, author_id); ok {

			if uac, load_err := userDAO.Load(p_id); load_err == nil { // FIXME: Load several participants in one operation
				result[uac.Id] = uac
			} else if err == dao.ErrNotFound {
				last_warning = ErrUnregisteredFriendsIgnored
				log.Printf("loadUserParticipants() Warning at userid %v: %v\n", p_id, load_err)
			} else {
				return nil, nil, load_err
			}

		} else if err != nil {
			return nil, nil, err
		} else {
			log.Println("loadUserParticipants() Not friends", author_id, "and", p_id)
			last_warning = ErrNonFriendsIgnored
		}
	}

	return result, last_warning, nil
}

func (s *Server) createParticipantList(users map[uint64]*core.UserAccount) map[uint64]*core.EventParticipant {

	result := make(map[uint64]*core.EventParticipant)

	for _, uac := range users {
		result[uac.Id] = uac.AsParticipant()
	}

	return result
}

func (s *Server) createParticipantListFromMap(participants map[uint64]*core.EventParticipant) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants))

	for _, p := range participants {
		result = append(result, p)
	}

	return result
}

func (s *Server) createParticipantsFromFriends(author_id uint64) map[uint64]*core.EventParticipant {

	dao := s.NewFriendDAO()
	friends, _ := dao.LoadFriends(author_id, ALL_CONTACTS_GROUP)

	if friends != nil {
		return core.CreateParticipantsFromFriends(author_id, friends)
	} else {
		log.Println("createParticipantsFromFriends() no friends or error")
		return nil
	}
}

// Called from multiple threads
// TODO: Update to send events + participants in one single message
func sendPrivateEvents(session *AyiSession) {

	server := session.Server
	dao := server.NewEventDAO()

	current_time := core.GetCurrentTimeMillis()
	events, err := dao.LoadUserEventsAndParticipants(session.UserId, current_time)

	if err != nil {
		log.Printf("sendPrivateEvents() to %v Error: %v\n", session.UserId, err)
		return
	}

	// For compatibility, split events into event info and participants
	half_events := make([]*core.Event, 0, len(events))
	for _, event := range events {
		half_events = append(half_events, event.GetEventWithoutParticipants())
	}

	// Send events list to user
	log.Printf("< (%v) SEND PRIVATE EVENTS (num.events: %v)", session.UserId, len(half_events))
	session.Write(session.NewMessage().EventsList(half_events)) // TODO: Change after remove compatibility

	// Send participants info for each event,  update participant status as delivered and notify it
	for _, event := range events {

		if len(event.Participants) == 0 {
			log.Printf("WARNING: Event %v has zero participants\n", event.EventId)
			continue
		}

		participants_filtered := session.Server.filterParticipantsMap(session.UserId, event.Participants)

		// Send attendance info
		session.Write(session.NewMessage().AttendanceStatus(event.EventId, participants_filtered))

		// Update participant status of the session user
		ownParticipant, ok := event.Participants[session.UserId]

		if ok && ownParticipant.Delivered != core.MessageStatus_CLIENT_DELIVERED {
			ownParticipant.Delivered = core.MessageStatus_CLIENT_DELIVERED
			dao.SetParticipantStatus(session.UserId, event.EventId, ownParticipant.Delivered)

			// Notify change in participant status to the other participants
			task := &NotifyParticipantChange{
				Event:               event,
				ParticipantsChanged: []uint64{session.UserId},
				Target:              core.GetParticipantsIdSlice(event.Participants),
			}

			// I'm also sending notification to the author. Could avoid this because author already knows
			// that the event has been send to him
			server.task_executor.Submit(task)
		}
	}
}

func sendAuthError(session *AyiSession) {
	session.Write(session.NewMessage().Error(proto.M_USER_AUTH, proto.E_INVALID_USER_OR_PASSWORD))
	log.Printf("< (%v) SEND INVALID USER OR PASSWORD\n", session)
}

func checkAuthenticated(session *AyiSession) {
	if !session.IsAuth {
		panic(ErrAuthRequired)
	}
}

func checkUnauthenticated(session *AyiSession) {
	if session.IsAuth {
		panic(ErrNoAuthRequired)
	}
}

func checkNoErrorOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func checkAtLeastOneEventOrPanic(events []*core.Event) {
	if len(events) == 0 {
		panic(ErrEventNotFound)
	}
}

func checkEventWritableOrPanic(event *core.Event) {
	current_time := core.GetCurrentTimeMillis()
	if event.StartDate < current_time || event.State == core.EventState_CANCELLED {
		panic(ErrNotWritableEvent)
	}
}

func checkEventAuthorOrPanic(author_id uint64, event *core.Event) {
	if event.AuthorId != author_id {
		panic(ErrAuthorMismatch)
	}
}

// Returns a participant list where users that will not assist the event or aren't
// friends of the given user are removed */
func (s *Server) filterParticipantsMap(participant uint64, participants map[uint64]*core.EventParticipant) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants))

	for _, p := range participants {
		if s.canSee(participant, p) {
			result = append(result, p)
		} else {
			result = append(result, p.AsAnonym())
		}
	}

	return result
}

func (s *Server) filterEventParticipants(targetParticipant uint64, participants map[uint64]*core.EventParticipant) map[uint64]*core.EventParticipant {

	result := make(map[uint64]*core.EventParticipant)

	for key, p := range participants {
		if s.canSee(targetParticipant, p) {
			result[key] = p
		} else {
			result[key] = p.AsAnonym()
		}
	}

	return result
}

func (s *Server) filterParticipantsSlice(participant uint64, participants []*core.EventParticipant) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants))

	for _, p := range participants {
		if s.canSee(participant, p) {
			result = append(result, p)
		} else {
			result = append(result, p.AsAnonym())
		}
	}

	return result
}

// Tells if participant p1 can see changes of participant p2
func (s *Server) canSee(p1 uint64, p2 *core.EventParticipant) bool {
	dao := s.NewFriendDAO()
	if p2.Response == core.AttendanceResponse_ASSIST ||
		/*p2.Response == core.AttendanceResponse_CANNOT_ASSIST ||*/
		p1 == p2.UserId {
		return true
	} else if ok, _ := dao.IsFriend(p2.UserId, p1); ok {
		return true
	}
	return false
}

func (s *Server) createUserAccount(user *core.UserAccount) error {

	// Check if its a valid user, so the input was correct
	if _, err := user.IsValid(); err != nil {
		return err
	}

	userDAO := s.NewUserDAO()

	// If it's a Facebook account (fbid and fbtoken are not empty) check token
	if user.HasFacebookCredentials() {
		fbsession := fb.NewSession(user.Fbtoken)
		if _, err := fb.CheckAccess(user.Fbid, fbsession); err != nil {
			return err
		}
	}

	// Insert into users database. Insert will fail if an existing user with the same
	// e-mail address already exists, or if the Facebook address is already being used
	// by another user. It also controls orphaned user_facebook_credentials rows due
	// to the way insertion is performed in Cassandra. When orphaned row is found and
	// grace period has not elapsed, an ErrGracePeriod error is triggered. A different
	// error message could be sent to the client whenever this happens. This way client
	// could be notified to wait grace period seconds and retry. However, an OPERATION
	// FAILED message is sent so far. Read UserDAO.insert for more info.
	if err := userDAO.Insert(user); err != nil {
		return err
	}

	// Insert OK
	if user.HasFacebookCredentials() {
		// Load profie picture from Facebook
		s.task_executor.Submit(&LoadFacebookProfilePicture{
			User:    user,
			Fbtoken: user.Fbtoken,
		})
		// Import Facebook friends that uses AreYouIN if needed
		s.task_executor.Submit(&ImportFacebookFriends{
			TargetUser: user,
			Fbtoken:    user.Fbtoken,
		})
	}

	return nil
}

// Assume picture is 512x512
func (s *Server) saveProfilePicture(user_id uint64, picture *core.Picture) error {

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > PROFILE_PICTURE_MAX_WIDTH || srcImage.Bounds().Dy() > PROFILE_PICTURE_MAX_HEIGHT {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := s.createThumbnails(srcImage, THUMBNAIL_MDPI_SIZE)
	if err != nil {
		return err
	}

	// Save profile picture (max 512x512)
	user_dao := s.NewUserDAO()

	err = user_dao.SaveProfilePicture(user_id, picture)
	if err != nil {
		return err
	}

	// Save thumbnails (50x50 to 200x200)
	thumbDAO := s.NewThumbnailDAO()
	err = thumbDAO.Insert(user_id, picture.Digest, thumbnails)
	if err != nil {
		return err
	}

	// Store digest in user's friends so that friends can know that
	// user profile picture has been changed next time they retrieve
	// user list
	friend_dao := s.NewFriendDAO()
	friends, err := friend_dao.LoadFriends(user_id, ALL_CONTACTS_GROUP)
	if err != nil {
		return err
	}

	for _, friend := range friends {
		err := friend_dao.SetPictureDigest(friend.GetUserId(), user_id, picture.Digest)
		if err != nil {
			log.Printf("Error setting picture digest for user %v and friend %v\n", user_id, friend.GetUserId())
		}
	}

	return nil
}

func (s *Server) removeProfilePicture(user_id uint64, picture *core.Picture) error {

	// Save empty profile picture
	user_dao := s.NewUserDAO()

	err := user_dao.SaveProfilePicture(user_id, picture)
	if err != nil {
		return err
	}

	// Remove thumbnails
	thumbDAO := s.NewThumbnailDAO()
	err = thumbDAO.Remove(user_id)
	if err != nil {
		return err
	}

	// Update digest in user's friends
	friend_dao := s.NewFriendDAO()
	friends, err := friend_dao.LoadFriends(user_id, ALL_CONTACTS_GROUP)
	if err != nil {
		return err
	}

	for _, friend := range friends {
		err := friend_dao.SetPictureDigest(friend.GetUserId(), user_id, picture.Digest)
		if err != nil {
			log.Printf("Error setting picture digest for user %v and friend %v\n", user_id, friend.GetUserId())
		}
	}

	return nil
}

func (s *Server) saveEventPicture(event_id uint64, picture *core.Picture) error {

	// The whole picture is the header. But It is also needed thumbnails for
	// event icon.

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > EVENT_PICTURE_MAX_WIDTH || srcImage.Bounds().Dy() > EVENT_PICTURE_MAX_HEIGHT {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := s.createThumbnails(srcImage, EVENT_THUMBNAIL)
	if err != nil {
		return err
	}

	// Save thumbnails
	thumbDAO := s.NewThumbnailDAO()
	err = thumbDAO.Insert(event_id, picture.Digest, thumbnails)
	if err != nil {
		return err
	}

	// Save event picture (always does it after thumbnails)
	eventDAO := s.NewEventDAO()
	err = eventDAO.SetEventPicture(event_id, picture)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) removeEventPicture(event_id uint64, picture *core.Picture) error {

	// Remove event picture
	eventDAO := s.NewEventDAO()
	err := eventDAO.SetEventPicture(event_id, picture)
	if err != nil {
		return err
	}

	// Remove thumbnails
	thumbDAO := s.NewThumbnailDAO()
	err = thumbDAO.Remove(event_id)
	if err != nil {
		return err
	}

	return nil
}

// Sync server-side friends groups with client-side friends groups. If groups
// provided by the client contains all of the groups, then a full sync is required,
// i.e. server-side groups that does not exist client-side are removed. Otherwise,
// if provided groups are only a subset, a merge of client and server data is
// performed. Conversely to full sync, merging process does not remove existing
// groups from the server but add new groups and modify existing ones. Regarding
// full sync, it is assumed that clientGroups contains all of the groups in client.
// Hence, if a group doesn't exist in client, it will be removed from server. Like
// a regular sync, new groups in client will be added to server. In whatever case,
// if a group already exists server-side, it will be updated with members from client
// group, removing those members that does not exist in client group (client is master).
// In other words, groups at server will be equal to groups at client at the end of the
// synchornisation process.
func (s *Server) syncFriendGroups(owner uint64, serverGroups []*core.Group,
	clientGroups []*core.Group, syncBehaviour core.SyncBehaviour) {

	friendsDAO := s.NewFriendDAO()

	// Copy map because it's gonna be modified
	clientGroupsCopy := make(map[int32]*core.Group)
	for _, group := range clientGroups {
		clientGroupsCopy[group.Id] = group
	}

	// Loop through server groups in order to know what
	// to do: update/replace or remove group from server
	for _, group := range serverGroups {

		if clientGroup, ok := clientGroupsCopy[group.Id]; ok {

			// Group exists.

			if clientGroup.Size == -1 && len(clientGroup.Members) == 0 {

				// Special case

				if clientGroup.Name == "" {

					// Group is marked for removal. So remove it from server

					err := friendsDAO.DeleteGroup(owner, group.Id)
					checkNoErrorOrPanic(err)

				} else if group.Name != clientGroup.Name {

					// Only Rename group
					err := friendsDAO.SetGroupName(owner, group.Id, clientGroup.Name)
					checkNoErrorOrPanic(err)

				}

			} else {

				// Update case

				if group.Name != clientGroup.Name {
					err := friendsDAO.SetGroupName(owner, group.Id, clientGroup.Name)
					checkNoErrorOrPanic(err)
				}

				s.syncGroupMembers(owner, group.Id, group.Members, clientGroup.Members)
			}

			// Delete also from copy because it has been processed
			delete(clientGroupsCopy, group.Id)

		} else if syncBehaviour == core.SyncBehaviour_TRUNCATE {

			// Remove

			err := friendsDAO.DeleteGroup(owner, group.Id)
			checkNoErrorOrPanic(err)
		}
	}

	// clientIndex contains only new groups. So add groups to server

	for _, group := range clientGroupsCopy {
		err := friendsDAO.AddGroup(owner, group)
		checkNoErrorOrPanic(err)
	}
}

func (s *Server) syncGroupMembers(user_id uint64, group_id int32, serverMembers []uint64, clientMembers []uint64) {

	// Index client members
	index := make(map[uint64]bool)
	for _, id := range clientMembers {
		index[id] = true
	}

	// Loop through all of the members of group owned by the server.
	// If member also exists client side, then keep it. Otherwise,
	// remove it.
	remove_ids := make([]uint64, 0, len(serverMembers)/2)

	for _, serverMemberId := range serverMembers {
		if _, ok := index[serverMemberId]; !ok {
			remove_ids = append(remove_ids, serverMemberId)
		} else {
			delete(index, serverMemberId)
		}
	}

	// After removing already existing members. Index contains only new members,
	// so add them.
	add_ids := make([]uint64, 0, len(clientMembers))
	for id, _ := range index {
		add_ids = append(add_ids, id)
	}

	// Proceed database I/O
	friendDAO := s.NewFriendDAO()

	if len(remove_ids) > 0 {
		err := friendDAO.DeleteMembers(user_id, group_id, remove_ids...)
		checkNoErrorOrPanic(err)
	}

	if len(add_ids) > 0 {
		err := friendDAO.AddMembers(user_id, group_id, add_ids...)
		checkNoErrorOrPanic(err)
	}
}

// Creates picture thumbnails for every supported dpi. Thumbnails size are
// (size_xy, size_xy)*scale_factor. Thumbnails are returned as byte slide
// JPEG encoded.
func (s *Server) createThumbnails(srcImage image.Image, size_xy int) (map[int32][]byte, error) {

	// Create thumbnails for distinct sizes
	thumbnails := make(map[int32][]byte)

	for _, dpi := range s.supportedDpi {
		// Compute size
		size := float32(size_xy) * (float32(dpi) / float32(IMAGE_MDPI))
		// Resize and crop image to fill size x size area
		dstImage := imaging.Thumbnail(srcImage, int(size), int(size), imaging.Lanczos)
		// Encode
		bytes := &bytes.Buffer{}
		err := jpeg.Encode(bytes, dstImage, nil)
		if err != nil {
			return nil, err
		}
		thumbnails[dpi] = bytes.Bytes()
	}

	return thumbnails, nil
}

func (s *Server) resizeImage(picture image.Image, width int) ([]byte, error) {
	resize_image := imaging.Resize(picture, width, 0, imaging.Lanczos)
	bytes := &bytes.Buffer{}
	err := jpeg.Encode(bytes, resize_image, nil)
	if err != nil {
		return nil, err
	}
	return bytes.Bytes(), nil
}

func (s *Server) getClosestDpi(reqDpi int32) int32 {

	if reqDpi <= IMAGE_MDPI {
		return IMAGE_MDPI
	} else if reqDpi >= IMAGE_XXXHDPI {
		return IMAGE_XXXHDPI
	}

	min_dist := math.MaxFloat32
	dpi_index := 0

	for i, dpi := range s.supportedDpi {
		dist := math.Abs(float64(reqDpi - dpi))
		if dist < min_dist {
			min_dist = dist
			dpi_index = i
		}
	}

	if s.supportedDpi[dpi_index] < reqDpi {
		dpi_index++
	}

	return s.supportedDpi[dpi_index]
}

func main() {

	// Create and init server
	server := NewTestServer()
	//server := NewServer() // Server is global

	// Register callbacks
	server.RegisterCallback(proto.M_PING, onPing)
	server.RegisterCallback(proto.M_PONG, onPong)
	server.RegisterCallback(proto.M_USER_CREATE_ACCOUNT, onCreateAccount)
	server.RegisterCallback(proto.M_USER_NEW_AUTH_TOKEN, onUserNewAuthToken)
	server.RegisterCallback(proto.M_USER_AUTH, onUserAuthentication)
	server.RegisterCallback(proto.M_GET_ACCESS_TOKEN, onNewAccessToken)
	server.RegisterCallback(proto.M_CREATE_EVENT, onCreateEvent)
	//server.RegisterCallback(proto.M_MODIFY_EVENT, onModifyEvent)
	server.RegisterCallback(proto.M_CANCEL_EVENT, onCancelEvent)
	server.RegisterCallback(proto.M_INVITE_USERS, onInviteUsers)
	server.RegisterCallback(proto.M_CONFIRM_ATTENDANCE, onConfirmAttendance)
	server.RegisterCallback(proto.M_USER_FRIENDS, onUserFriends)
	server.RegisterCallback(proto.M_GET_USER_ACCOUNT, onGetUserAccount)
	server.RegisterCallback(proto.M_CHANGE_PROFILE_PICTURE, onChangeProfilePicture)
	server.RegisterCallback(proto.M_CLOCK_REQUEST, onClockRequest)
	server.RegisterCallback(proto.M_IID_TOKEN, onIIDTokenReceived)
	server.RegisterCallback(proto.M_CHANGE_EVENT_PICTURE, onChangeEventPicture)
	server.RegisterCallback(proto.M_SYNC_GROUPS, onSyncGroups)
	server.RegisterCallback(proto.M_GET_GROUPS, onGetGroups)
	server.RegisterCallback(proto.M_LIST_PRIVATE_EVENTS, onRequestPrivateEvents)

	// Create shell and start listening in 2022 tcp port
	shell := NewShell(server)
	go shell.StartTermSSH()

	// Create images HTTP server and start
	images_server := imgserv.NewServer("192.168.1.10", "areyouin")
	images_server.Run()

	// start server loop
	server.Run()
}
