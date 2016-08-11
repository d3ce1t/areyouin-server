package main

import (
	"fmt"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"
	"time"

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

	case model.SignalEventParticipantsInvited:
		// Enqueue signal in order to cancel previous notifications
		collapseKey := fmt.Sprintf("event#%v", signal.Data["EventID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)

	case model.SignalEventCancelled:
		// Enqueue signal in order to cancel previous notifications
		collapseKey := fmt.Sprintf("event#%v", signal.Data["EventID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)
		// Process signal inmediately
		m.processEventCancelledSignal(signal)

	case model.SignalEventInfoChanged:
		collapseKey := fmt.Sprintf("event-change#%v", signal.Data["EventID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)

	case model.SignalParticipantChanged:
		collapseKey := fmt.Sprintf("event#%v#%v", signal.Data["EventID"], signal.Data["UserID"])
		m.signalsQueue.AddWithKey(collapseKey, signal)

	case model.SignalFriendRequestAccepted:
		m.processFriendRequestAcceptedSignal(signal)

	default:
		m.signalsQueue.Add(signal)
	}
}

func (m *ModelObserver) processDelayedChanges() {

	item := m.signalsQueue.Remove()

	for item != nil {

		signal := item.(*model.Signal)

		switch signal.Type {

		case model.SignalEventParticipantsInvited:
			m.processNewEventSignal(signal)

		/*case model.SignalEventCancelled:
		m.processEventCancelledSignal(signal)*/

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

		if pID == event.AuthorId() {
			continue
		}

		go func(participantID int64) {

			// NOTE: May panic (call m.model.Events.ChangeDeliveryState)

			//session := m.server.getSession(participantID)
			//if session == nil {
			// Notification
			//if participantID != event.AuthorId() {
			sendNewEventNotification(event, participantID)
			//}
			//return
			//}

			/*coreEvent := convEvent2Net(event)
			coreEvent.Participants[session.UserId].Delivered = core.InvitationStatus_CLIENT_DELIVERED
			message := session.NewMessage().InvitationReceived(coreEvent)
			future := NewFuture(true)
			if ok := session.WriteAsync(future, message); ok {
				// Blocks until ACK (true) or timeout (false)
				if sent := <-future.C; sent {
					_, err := m.model.Events.ChangeDeliveryState(event, session.UserId, api.InvitationStatus_CLIENT_DELIVERED)
					if err != nil {
						log.Println("processNewEventSignal Err:", err)
					}
				} else if participantID != event.AuthorId() {
					// Notification
					sendNewEventNotification(event, participantID)
				}
			} else if participantID != event.AuthorId() {
				// Notification
				sendNewEventNotification(event, participantID)
			}*/
		}(pID)
	}

	oldParticipants := signal.Data["OldParticipants"].([]int64)

	if len(oldParticipants) > 0 {

		// Send participants change to old participants

		participantList := make(map[int64]*model.Participant)
		for _, id := range newParticipants {
			participantList[id] = event.GetParticipant(id)
		}
		netParticipants := convParticipantList2Net(participantList)

		for _, pID := range oldParticipants {

			session := m.server.getSession(pID)
			if session == nil {
				continue
			}

			go func(session *AyiSession) {
				message := session.NewMessage().AttendanceStatusWithNumGuests(event.Id(), netParticipants, int32(event.NumGuests()))
				if ok := session.Write(message); ok {
					log.Printf("< (%v) EVENT %v CHANGED (%v participants changed)\n", session.UserId, event.Id(), len(netParticipants))
				} else {
					log.Println("processNewEventSignal: Coudn't send notification to", session.UserId)
				}
			}(session)
		}
	}
}

func (m *ModelObserver) processParticipantChangeSignal(signal *model.Signal) {

	event := signal.Data["Event"].(*model.Event)
	participant := signal.Data["Participant"].(*model.Participant)
	oldResponse := signal.Data["OldResponse"].(api.AttendanceResponse)

	// Send participant change to event participants

	participantList := make(map[int64]*model.Participant)
	participantList[participant.Id()] = participant
	netParticipants := convParticipantList2Net(participantList)

	for _, pID := range event.ParticipantIds() {

		go func(participantID int64) {

			session := m.server.getSession(participantID)
			if session == nil {
				// Notification
				if oldResponse != participant.Response() && participant.Id() != participantID {
					sendEventResponseNotification(event, participant.Id(), participantID)
				}
				return
			}

			message := session.NewMessage().AttendanceStatus(event.Id(), netParticipants)
			if ok := session.Write(message); ok {
				log.Printf("< (%v) EVENT %v CHANGED (%v participants changed)\n", session.UserId, event.Id(), len(netParticipants))
			} else {
				log.Println("processParticipantChangeSignal: Coudn't send notification to", session.UserId)
			}

		}(pID)
	}
}

func (m *ModelObserver) processEventCancelledSignal(signal *model.Signal) {

	event := signal.Data["Event"].(*model.Event)
	cancelledBy := signal.Data["CancelledBy"].(int64)
	liteEvent := convEvent2Net(event.CloneEmptyParticipants())

	for _, pID := range event.ParticipantIds() {

		go func(participantID int64) {

			session := m.server.getSession(participantID)
			if session == nil {
				if participantID != cancelledBy {
					// Notification
					sendEventCancelledNotification(event, participantID)
				}
				return
			}

			packet := session.NewMessage().EventCancelled(cancelledBy, liteEvent)
			future := NewFuture(true)

			if ok := session.WriteAsync(future, packet); ok {
				// Blocks until ACK (true) or timeout (false)
				if sent := <-future.C; sent {
					log.Printf("< (%v) EVENT CANCELLED (event_id=%v)\n", session.UserId, event.Id())
				} else if session.UserId != cancelledBy {
					// Notification
					sendEventCancelledNotification(event, participantID)
				}
			} else if session.UserId != cancelledBy {
				// Notification
				sendEventCancelledNotification(event, participantID)
			}

		}(pID)
	}
}

func (m *ModelObserver) processEventChangedSignal(signal *model.Signal) {

	event := signal.Data["Event"].(*model.Event)
	liteEvent := convEvent2Net(event.CloneEmptyParticipants())

	for _, pID := range event.ParticipantIds() {

		session := m.server.getSession(pID)
		if session == nil {
			continue
		}

		go func(session *AyiSession) {
			msg := session.NewMessage().EventModified(liteEvent)
			if session.Write(msg) {
				log.Printf("< (%v) EVENT %v CHANGED\n", session.UserId, event.Id())
			} else {
				log.Println("processEventChangedSignal: Coudn't send notification to", session.UserId)
			}
		}(session)
	}
}

func (m *ModelObserver) processNewFriendRequestSignal(signal *model.Signal) {

	fromUser := signal.Data["FromUser"].(*model.UserAccount)
	toUser := signal.Data["ToUser"].(*model.UserAccount)
	friendRequest := signal.Data["FriendRequest"].(*model.FriendRequest)

	if session := m.server.getSession(toUser.Id()); session != nil {
		session.Write(session.NewMessage().FriendRequestReceived(convFriendRequest2Net(friendRequest)))
		log.Printf("< (%v) SEND FRIEND REQUEST: %v\n", session, friendRequest)
	} else {
		sendFriendRequestNotification(fromUser.Name(), toUser.Id())
	}
}

func (m *ModelObserver) processFriendRequestAcceptedSignal(signal *model.Signal) {

	fromUser := signal.Data["FromUser"].(*model.UserAccount)
	toUser := signal.Data["ToUser"].(*model.UserAccount)

	// Send Friend List to both users if connected
	sendFriends := func(userID int64, friendName string) {

		// TODO: May panic

		if session := m.server.getSession(userID); session != nil {

			friends, err := m.model.Friends.GetAllFriends(userID)
			if err != nil {
				log.Println("SendUserFriends Error:", err)
				return
			}

			if len(friends) > 0 {
				session.Write(session.NewMessage().FriendsList(convFriendList2Net(friends)))
				log.Printf("< (%v) SEND USER FRIENDS (num.friends: %v)\n", userID, len(friends))
			}

		} else {
			sendNewFriendNotification(friendName, userID)
		}
	}

	go sendFriends(fromUser.Id(), toUser.Name())
	go sendFriends(toUser.Id(), fromUser.Name())
}
