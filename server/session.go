package main

import (
	"log"
	"net"
	proto "peeple/areyouin/protocol"
	"time"
)

// Creates a new sessions with an already connected client
func NewSession(conn net.Conn, server *Server) *AyiSession {

	session := &AyiSession{
		Conn:                conn,
		UserId:              0,
		IsAuth:              false,
		NotificationChannel: make(chan *Notification, 5),
		SocketChannel:       make(chan *proto.AyiPacket),
		SocketError:         make(chan error),
		Server:              server,
		lastRecvMsg:         time.Now().UTC(),
	}

	// Read socket in background and send result through channels SocketChannel
	// and SocketError
	if conn != nil {
		go func() {
			for !session.IsClosed {
				session.doRead()
			}
		}()
	} else {
		log.Println("WARNING: NewSession connection is nil")
	}

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
	IsClosed            bool
	lastRecvMsg         time.Time
}

func (s *AyiSession) String() string {
	return s.Conn.RemoteAddr().String()
}

func (s *AyiSession) Notify(notification *Notification) {
	s.NotificationChannel <- notification
}

func (s *AyiSession) ProcessNotification(notification *Notification) {

	if err := s.WriteReply(notification.Message); err != nil {
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
	packet, err := proto.ReadPacket(s.Conn) // Blocked here
	if err != nil {
		s.SocketError <- err
	} else {
		s.SocketChannel <- packet
	}
}

func (s *AyiSession) SendPing() {
	ping_msg := proto.NewMessage().Ping().Marshal()
	s.WriteReply(ping_msg)
}

func (s *AyiSession) WriteReply(reply []byte) error {
	client := s.Conn
	client.SetWriteDeadline(time.Now().Add(MAX_WRITE_TIMEOUT))
	_, err := client.Write(reply)
	if err != nil {
		log.Println("Coudn't send reply: ", err)
	}
	return err
}

func (s *AyiSession) Close() {
	s.IsClosed = true
	s.IsAuth = false
	s.UserId = 0
	s.Conn.Close()
}
