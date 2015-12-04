package main

import (
	"log"
	"net"
)

type Notification struct {
	Message  []byte
	Callback func()
}

type AyiSession struct {
	Conn                net.Conn
	UserId              uint64
	IsAuth              bool
	NotificationChannel chan *Notification // Channel used to send notifications to clients
}

func (s *AyiSession) String() string {
	return s.Conn.RemoteAddr().String()
}

func (s *AyiSession) Notify(notification *Notification) {
	s.NotificationChannel <- notification
}

func (s *AyiSession) ProcessNotification(notification *Notification) {
	log.Println("Send notification to", s.UserId)
	writeReply(notification.Message, s)
	if notification.Callback != nil {
		notification.Callback()
	}
}

func (s *AyiSession) Close() {
	s.Conn.Close()
	s.IsAuth = false
	s.UserId = 0
}
