package main

import (
	"net"
)

type AyiSession struct {
	Conn                net.Conn
	UserId              uint64
	IsAuth              bool
	NotificationChannel chan []byte // Channel used to send notifications to clients
}

func (s *AyiSession) String() string {
	return s.Conn.RemoteAddr().String()
}

func (s *AyiSession) Notify(msg []byte) {
	s.NotificationChannel <- msg
}

func (s *AyiSession) Close() {
	s.Conn.Close()
	s.IsAuth = false
	s.UserId = 0
}
