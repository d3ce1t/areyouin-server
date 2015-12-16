package main

import (
	"log"
	"net"
)

func NewSession(conn net.Conn, server *Server) *AyiSession {
	return &AyiSession{
		Conn:                conn,
		UserId:              0,
		IsAuth:              false,
		NotificationChannel: make(chan *Notification, 5),
		Server:              server,
	}
}

type Notification struct {
	Message  []byte
	Callback func()
}

type AyiSession struct {
	Conn                net.Conn
	UserId              uint64
	IsAuth              bool
	NotificationChannel chan *Notification // Channel used to send notifications to clients
	Server              *Server
}

func (s *AyiSession) String() string {
	return s.Conn.RemoteAddr().String()
}

func (s *AyiSession) Notify(notification *Notification) {
	s.NotificationChannel <- notification
}

func (s *AyiSession) ProcessNotification(notification *Notification) {

	if err := writeReply(notification.Message, s); err != nil {
		log.Println("ProcessNotification:", err)
		return
	}

	log.Println("Send notification to", s.UserId)
	if notification.Callback != nil {
		notification.Callback()
	}
}

func (s *AyiSession) Close() {
	s.Conn.Close()
	s.IsAuth = false
	s.UserId = 0
}
