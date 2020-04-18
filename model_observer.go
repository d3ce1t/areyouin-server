package main

import (
	"fmt"
	"log"
	"time"

	"github.com/d3ce1t/areyouin-server/model"
	"github.com/d3ce1t/areyouin-server/utils"

	"github.com/imkira/go-observer"
)

// Model observer constants
const (
	WindowTemporalSize = 20 * time.Second
)

type ModelObserver struct {
	model        *model.AyiModel
	server       *Server
	signalsQueue *utils.Queue
	eventStream  observer.Stream
	friendStream observer.Stream
}

func newModelObserver(server *Server) *ModelObserver {
	return &ModelObserver{
		model:        server.Model,
		signalsQueue: utils.NewQueue(),
		server:       server,
	}
}

func (m *ModelObserver) run() {

	m.eventStream = m.model.Events.Observe()
	m.friendStream = m.model.Friends.Observe()
	tickC := time.Tick(WindowTemporalSize)

	for {
		m.receiveSignals(tickC)
	}
}

func (m *ModelObserver) receiveSignals(tickC <-chan time.Time) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("* modelObserver receiveSignals err: %v", r)
		}
	}()

	for {
		select {
		case <-m.eventStream.Changes():
			m.eventStream.Next()
			eventSignal := m.eventStream.Value().(*model.Signal)
			m.processSignal(eventSignal)

		case <-m.friendStream.Changes():
			m.friendStream.Next()
			friendSignal := m.friendStream.Value().(*model.Signal)
			m.processSignal(friendSignal)

		case <-tickC:
			m.processDelayedChanges()
		}
	}
}

func (m *ModelObserver) processSignal(signal *model.Signal) {
	switch signal.Type {

	case model.SignalNewEvent:
		fallthrough
	case model.SignalEventParticipantsInvited:
		// Enqueue signal in order to cancel previous notifications
		collapseKey := fmt.Sprintf("event#%v", signal.Data["EventID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)

	case model.SignalEventCancelled:
		// Enqueue signal in order to cancel previous notifications
		collapseKey := fmt.Sprintf("event#%v", signal.Data["EventID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)

	case model.SignalEventInfoChanged:
		collapseKey := fmt.Sprintf("event-change#%v", signal.Data["EventID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)

	case model.SignalParticipantChanged:
		collapseKey := fmt.Sprintf("event#%v#%v", signal.Data["EventID"], signal.Data["UserID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)

	case model.SignalFriendRequestAccepted:
		m.processFriendRequestAcceptedSignal(signal)

	case model.SignalNewFriendsImported:
		m.processFriendsImported(signal)

	default:
		m.signalsQueue.Add(signal)
	}
}

func (m *ModelObserver) processDelayedChanges() {

	item := m.signalsQueue.Remove()

	for item != nil {

		signal := item.(*model.Signal)

		switch signal.Type {

		case model.SignalNewEvent:
			fallthrough
		case model.SignalEventParticipantsInvited:
			m.processNewEventSignal(signal)

		case model.SignalEventCancelled:
			m.processEventCancelledSignal(signal)

		case model.SignalEventInfoChanged:
			m.processEventChangedSignal(signal)

		case model.SignalParticipantChanged:
			m.processParticipantChangeSignal(signal)

		case model.SignalNewFriendRequest:
			m.processNewFriendRequestSignal(signal)

			/*case model.SignalFriendRequestAccepted:
			m.processFriendRequestAcceptedSignal(signal)*/
		}

		item = m.signalsQueue.Remove()
	}
}

func (m *ModelObserver) processNewEventSignal(signal *model.Signal) {

	event := signal.Data["Event"].(*model.Event)
	newParticipants := signal.Data["NewParticipants"].([]int64)

	// Send invitation to new participants
	for _, pID := range newParticipants {

		if pID == event.AuthorID() {
			continue
		}

		go func(participantID int64) {
			// Notification
			sendNewEventNotification(event, participantID)
		}(pID)
	}

	oldParticipants := signal.Data["OldParticipants"].([]int64)

	if len(oldParticipants) > 0 {

		// Send participants change to old participants

		participantList := make(map[int64]*model.Participant)
		for _, id := range newParticipants {
			participantList[id], _ = event.Participants.Get(id)
		}
		netParticipants := convParticipantList2Net(participantList)

		for _, pID := range oldParticipants {

			session := m.server.getSession(pID)
			if session == nil {
				continue
			}

			go func(session *AyiSession) {
				message := session.NewMessage().AttendanceStatusWithNumGuests(event.Id(), netParticipants, event.NumGuests())
				if ok := session.Write(message); ok {
					log.Printf("< (%v) EVENT %v ATTENDANCE STATUS CHANGED (%v participants changed)\n", session.UserId, event.Id(), len(netParticipants))
				} else {
					log.Println("* processNewEventSignal: Coudn't send message to", session.UserId)
				}
			}(session)
		}
	}
}

func (m *ModelObserver) processEventCancelledSignal(signal *model.Signal) {

	event := signal.Data["Event"].(*model.Event)
	cancelledBy := signal.Data["CancelledBy"].(int64)

	for _, pID := range event.Participants.Ids() {

		if pID == cancelledBy {
			continue
		}

		go func(participantID int64) {
			// Notification
			sendEventCancelledNotification(event, participantID)
		}(pID)
	}
}

func (m *ModelObserver) processParticipantChangeSignal(signal *model.Signal) {

	participant := signal.Data["Participant"].(*model.Participant)
	oldParticipant := signal.Data["OldParticipant"].(*model.Participant)
	eventID := participant.EventID()
	event, err := m.model.Events.LoadEvent(eventID)

	if err != nil {
		log.Printf("* processParticipantChangeSignal Error: %v", err)
		return
	}

	// Send participant change to event participants

	participantList := make(map[int64]*model.Participant)
	participantList[participant.Id()] = participant
	netParticipant := convParticipantList2Net(participantList)

	for _, pID := range event.Participants.Ids() {

		go func(userID int64) {

			// Notification
			if participant.Id() != userID && oldParticipant.Response() != participant.Response() {
				sendEventResponseNotification(event, participant.Id(), userID)
			}

			if session := m.server.getSession(userID); session != nil {
				message := session.NewMessage().AttendanceStatus(event.Id(), netParticipant)
				if ok := session.Write(message); ok {
					log.Printf("< (%v) EVENT %v ATTENDANCE STATUS (%v participants changed)\n", session.UserId, event.Id(), len(netParticipant))
				} else {
					log.Println("* processParticipantChangeSignal: Coudn't send message to", session.UserId)
				}
			}

		}(pID)
	}
}

func (m *ModelObserver) processEventChangedSignal(signal *model.Signal) {

	event := signal.Data["Event"].(*model.Event)
	liteEvent := convEvent2Net(event.CloneWithEmptyParticipants())

	for _, pID := range event.Participants.Ids() {

		session := m.server.getSession(pID)
		if session == nil {
			continue
		}

		go func(session *AyiSession) {
			msg := session.NewMessage().EventModified(liteEvent)
			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED\n", session.UserId, event.Id())
			} else {
				log.Println("processEventChangedSignal: Coudn't send message to", session.UserId)
			}
		}(session)
	}
}

func (m *ModelObserver) processNewFriendRequestSignal(signal *model.Signal) {

	fromUser := signal.Data["FromUser"].(*model.UserAccount)
	toUser := signal.Data["ToUser"].(*model.UserAccount)
	friendRequest := signal.Data["FriendRequest"].(*model.FriendRequest)

	// Notification
	sendFriendRequestNotification(fromUser.Name(), toUser.Id())

	if session := m.server.getSession(toUser.Id()); session != nil {
		session.Write(session.NewMessage().FriendRequestReceived(convFriendRequest2Net(friendRequest)))
		log.Printf("< (%v) SEND FRIEND REQUEST: %v\n", session, friendRequest)
	}
}

func (m *ModelObserver) processFriendRequestAcceptedSignal(signal *model.Signal) {

	fromUser := signal.Data["FromUser"].(*model.UserAccount)
	toUser := signal.Data["ToUser"].(*model.UserAccount)

	// Send Friend list to userID and notify with "you and friendName
	// are now friends"
	notifyFriend := func(userID int64, friendName string) {
		sendNewFriendNotification(friendName, userID)
		m.sendFriends(userID)
	}

	go notifyFriend(fromUser.Id(), toUser.Name())
	go notifyFriend(toUser.Id(), fromUser.Name())
}

func (m *ModelObserver) processFriendsImported(signal *model.Signal) {

	// User who performed the import
	user := signal.Data["User"].(*model.UserAccount)

	// New friends added
	addedFriends := signal.Data["NewFriends"].([]*model.UserAccount)

	// Initial import is true if account was just created. Otherwise, is false.
	initialImport := signal.Data["InitialImport"].(bool)

	// Send Friend list to userID and notify with "you and friendName
	// are now friends"
	notifyFriend := func(userID int64, friendName string, initialImport bool) {

		if initialImport {
			// Do another notification
		} else {
			sendNewFriendNotification(friendName, userID)
		}

		m.sendFriends(userID)
	}

	// Loop through added friends in order to notify them
	for _, newFriend := range addedFriends {
		// Send friend list and notify newFriend that he is now friend of user
		go notifyFriend(newFriend.Id(), user.Name(), initialImport)
	}

	if len(addedFriends) > 0 {
		// Send friend list to user user
		go m.sendFriends(user.Id())
	}
}

func (m *ModelObserver) sendFriends(userID int64) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("* SendFriends Error: %v\n", r)
		}
	}()

	if session := m.server.getSession(userID); session != nil {

		// May panic so defer was added above
		friends, err := m.model.Friends.GetAllFriends(userID)
		if err != nil {
			log.Println("SendUserFriends Error:", err)
			return
		}

		if len(friends) > 0 {
			session.Write(session.NewMessage().FriendsList(convFriendList2Net(friends)))
			log.Printf("< (%v) SEND USER FRIENDS (num.friends: %v)\n", userID, len(friends))
		}
	}
}
