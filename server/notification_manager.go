package main

import (
	"log"
	"peeple/areyouin/model"
)

type NotificationManager struct {
	model *model.AyiModel
}

func newNotificationManager(model *model.AyiModel) *NotificationManager {
	return &NotificationManager{
		model: model,
	}
}

func (m *NotificationManager) run() {

	stream := m.model.Events.Observe()
	val := stream.Value()
	log.Printf(" * EVENT OBSERVER RUNNING: %v", val)

	for {
		select {
		// Wait for changes
		case <-stream.Changes():
			// advance to next value
			stream.Next()
			// new value
			val = stream.Value()
			log.Printf("got new value: %v", val)
		}
	}
}
