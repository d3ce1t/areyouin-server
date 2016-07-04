package main

import (
  core "peeple/areyouin/common"
  proto "peeple/areyouin/protocol"
  "log"
  "encoding/base64"
)

type UserData struct {
  Future   chan bool
  IIDToken *core.IIDToken
}

type NotifyEventCancelled struct {
	CancelledBy int64
	Event       *core.Event
  server      *Server
  user_data map[int64]*UserData
}

func (t *NotifyEventCancelled) Run(ex *TaskExecutor) {

  if len(t.Event.GetParticipants()) == 0 {
    log.Println("NotifyEventCancelled: There aren't targetted participants to send notification")
    return
  }

	t.server = ex.server
	t.user_data = make(map[int64]*UserData)
	lite_event := t.Event.GetEventWithoutParticipants()
	gcm_data := proto.NewPacket(1).EventCancelled(t.CancelledBy, lite_event)
	base64_data := base64.StdEncoding.EncodeToString(gcm_data.Marshal())

	for _, participant := range t.Event.GetParticipants() {

    session := t.server.GetSession(participant.UserId)

    if session == nil {

      // Not connected

      log.Printf("* (%v) User isn't connected. Fallback to GcmNotification\n", participant.UserId)
      t.sendNotificationFallback(participant.UserId, base64_data)

    } else {

      // Connected

      packet := session.NewMessage().EventCancelled(t.CancelledBy, lite_event)
      future := NewFuture(true)

      if ok := session.WriteAsync(future, packet); ok {
        t.user_data[participant.UserId] = &UserData{future.C, session.IIDToken}
        log.Printf("< (%v) EVENT CANCELLED (event_id=%v)\n", session.UserId, t.Event.EventId)
      } else {
        log.Printf("* (%v) Session write failed. Fallback to GcmNotification\n", participant.UserId)
        t.sendNotificationFallback(participant.UserId, base64_data)
      }
    }

	} // End loop

	for participant_id, data := range t.user_data {
		ok := <-data.Future // TODO: Code is blocked 10 seconds as much for each participant. CHANGE IT!!
		if !ok {
			log.Printf("* (%v) ACK Timeout. Fallback to GcmNotification\n", participant_id)
      t.sendNotificationFallback(participant_id, base64_data)
		}
	}
}

func (t *NotifyEventCancelled) sendNotificationFallback(participant_id int64, gcm_data string) {

  userDAO := t.server.NewUserDAO()
  user, err := userDAO.Load(participant_id)
  if err != nil {
    log.Printf("* Notify event cancelled error (userId %v) %v\n", user.GetUserId(), err)
    return
  }

  if user.IIDtoken == "" {
    log.Printf("* (%v) Coudn't send GCM event cancelled notification (Invalid IID token)", participant_id)
    return
  }

  if user.NetworkVersion == 0 || user.NetworkVersion == 1 {
    sendGcmEventNotification(participant_id, user.IIDtoken, t.Event.StartDate, gcm_data)
  } else {
    ttl := uint32(t.Event.StartDate - core.GetCurrentTimeMillis()) / 1000
    sendGcmDataAvailableNotification(participant_id, user.IIDtoken,  ttl)
  }
}
