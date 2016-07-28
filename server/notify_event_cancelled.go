package main

import (
	"log"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"
)

type UserData struct {
	Future   chan bool
	IIDToken *model.IIDToken
}

type NotifyEventCancelled struct {
	CancelledBy int64
	Event       *model.Event
	server      *Server
	user_data   map[int64]*UserData
}

func (t *NotifyEventCancelled) Run(ex *TaskExecutor) {

	if t.Event.NumGuests() == 0 {
		log.Println("NotifyEventCancelled: There aren't targetted participants to send notification")
		return
	}

	t.server = ex.server
	t.user_data = make(map[int64]*UserData)
	lite_event := convEvent2Net(t.Event.CloneEmptyParticipants())

	for _, participant := range t.Event.Participants() {

		session := t.server.GetSession(participant.Id())

		if session == nil {

			// Not connected

			log.Printf("* (%v) User isn't connected. Fallback to GcmNotification\n", participant.Id())
			t.sendNotificationFallback(participant.Id())

		} else {

			// Connected

			packet := session.NewMessage().EventCancelled(t.CancelledBy, lite_event)
			future := NewFuture(true)

			if ok := session.WriteAsync(future, packet); ok {
				t.user_data[participant.Id()] = &UserData{future.C, session.IIDToken}
				log.Printf("< (%v) EVENT CANCELLED (event_id=%v)\n", session.UserId, t.Event.Id())
			} else {
				log.Printf("* (%v) Session write failed. Fallback to GcmNotification\n", participant.Id())
				t.sendNotificationFallback(participant.Id())
			}
		}

	}

	for participant_id, data := range t.user_data {
		ok := <-data.Future // TODO: Code is blocked 10 seconds as much for each participant. CHANGE IT!!
		if !ok {
			log.Printf("* (%v) ACK Timeout. Fallback to GcmNotification\n", participant_id)
			t.sendNotificationFallback(participant_id)
		}
	}
}

func (t *NotifyEventCancelled) sendNotificationFallback(participantID int64) {

	user, err := t.server.Model.Accounts.GetUserAccount(participantID)
	if err != nil {
		log.Printf("* Notify event cancelled error (userId %v) %v\n", participantID, err)
		return
	}

	iidToken := user.PushToken()

	if iidToken.Token() == "" {
		log.Printf("* (%v) Coudn't send GCM event cancelled notification (Invalid IID token)", participantID)
		return
	}

	ttl := uint32(t.Event.StartDate()-utils.GetCurrentTimeMillis()) / 1000
	sendGcmDataAvailableNotification(participantID, &iidToken, ttl)
}
