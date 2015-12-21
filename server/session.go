package main

import (
	"log"
	"net"
	proto "peeple/areyouin/protocol"
)

func NewSession(conn net.Conn, server *Server) *AyiSession {

	session := &AyiSession{
		Conn:                conn,
		UserId:              0,
		IsAuth:              false,
		NotificationChannel: make(chan *Notification, 5),
		SocketChannel:       make(chan *proto.AyiPacket),
		SocketError:         make(chan error),
		Server:              server,
	}

	// Read socket in background and send result through channels SocketChannel
	// and SocketError
	go func() {
		for !session.isClosed {
			session.doRead()
		}
	}()

	return session
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
	SocketChannel       chan *proto.AyiPacket
	SocketError         chan error
	Server              *Server
	isClosed            bool
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

// Do a read and blocks until its done or an error happens
func (s *AyiSession) doRead() {
	packet, err := proto.ReadPacket(s.Conn)
	if err != nil {
		s.SocketError <- err
	} else {
		s.SocketChannel <- packet
	}
}

func (s *AyiSession) Close() {
	s.isClosed = true
	s.Conn.Close()
	s.IsAuth = false
	s.UserId = 0
}
