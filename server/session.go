package main

import (
	"log"
	"net"
	proto "peeple/areyouin/protocol"
	"time"
)

const (
	// Event types
	READ_MESSAGE_EVENT      = 0
	IDLE_EVENT              = 1
	ERROR_EVENT             = 2
	CONNECTION_CLOSED_EVENT = 3

	// Times
	MAX_IDLE_TIME          = 30 * time.Minute
	MAX_LOGIN_TIME         = 30 * time.Second
	PING_INTERVAL_MS       = 29 * time.Minute
	PING_RETRY_INTERVAL_MS = 20 * time.Second

	// Channels Sizes
	EVENT_CHANNEL_SIZE = 10
	WRITE_CHANNEL_SIZE = 3
)

type SessionEvent struct {
	Id     uint32
	Object interface{}
}

type WriteMsg struct {
	Data []byte
	C    chan bool
}

// Creates a new sessions with an already connected client
func NewSession(conn net.Conn, server *Server) *AyiSession {

	session := &AyiSession{
		Conn:        conn,
		UserId:      0,
		IsAuth:      false,
		eventChan:   make(chan *SessionEvent, EVENT_CHANNEL_SIZE),
		writeChan:   make(chan *WriteMsg, WRITE_CHANNEL_SIZE),
		Server:      server,
		lastRecvMsg: time.Now().UTC(),
	}

	return session
}

type AyiSession struct {
	Conn        net.Conn
	UserId      uint64
	IsAuth      bool
	eventChan   chan *SessionEvent
	writeChan   chan *WriteMsg
	ticker      *time.Ticker
	Server      *Server
	closed      bool
	lastRecvMsg time.Time
	OnRead      func(s *AyiSession, packet *proto.AyiPacket)
	OnError     func(s *AyiSession, err error)
	OnClosed    func(s *AyiSession, peer bool)
	pingTime    time.Time
}

func (s *AyiSession) IsClosed() bool {
	return s.closed
}

func (s *AyiSession) String() string {
	return s.Conn.RemoteAddr().String()
}

func (s *AyiSession) Write(data []byte) (ok bool) {
	return s.WriteAsync(data, nil)
}

func (s *AyiSession) WriteSync(data []byte) bool {
	var ok bool
	c := make(chan bool, 1)
	if ok = s.WriteAsync(data, c); ok {
		ok = <-c // Block here
	}
	return ok
}

func (s *AyiSession) WriteAsync(data []byte, c chan bool) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
			log.Printf("Session %v Write Error: %v\n", s, r)
		}
	}()
	s.writeChan <- &WriteMsg{Data: data, C: c} // may panic if writeChan is closed
	return true
}

func (s *AyiSession) RunLoop() {

	defer func() { // startRecv and startWrite throw panic
		if r := recover(); r != nil {
			log.Printf("Session %v Panic: %v\n", s, r)
		}
	}()

	s.startRecv()
	stop_chan := s.startTicker()
	s.startWrite()

	defer func() {
		stop_chan <- true
		close(stop_chan)
		s.IsAuth = false
		//s.UserId = 0 // Need to preserve UserID in order to unregister session
	}()

	s.pingTime = time.Now().Add(PING_INTERVAL_MS) // Move to session
	exit := false

	for !exit {
		exit = s.eventLoop() // if panic returns false
	}

	log.Println("Event loop stopped")
}

func (s *AyiSession) eventLoop() (exit bool) {

	defer func() {
		if r := recover(); r != nil {
			exit = s.closed
			log.Printf("Session %v EventLoop Panic: %v\n", s, r)
		}
	}()

	for event := range s.eventChan {

		switch event.Id {
		case READ_MESSAGE_EVENT:
			s.OnRead(s, event.Object.(*proto.AyiPacket))
		/*case POST_MESSAGE_EVENT:
		s.processPost(event.Object.(*Notification))*/
		case CONNECTION_CLOSED_EVENT:
			s.OnClosed(s, event.Object.(bool))
			close(s.eventChan)
		case ERROR_EVENT:
			s.OnError(s, event.Object.(error))
		case IDLE_EVENT:
			s.keepAlive()
		}
	}

	return s.closed
}

func (s *AyiSession) Exit() {
	s.ticker.Stop()
	s.closeSocket()
}

func (s *AyiSession) closeSocket() {
	if !s.closed {
		s.closed = true
		close(s.writeChan)
		s.Conn.Close()
	}
}

// Read socket in background and send result through EventLoopChan
func (s *AyiSession) startRecv() {
	if s.Conn != nil {
		go func() {
			for !s.closed {
				s.doRead()
			}
		}()
	} else {
		panic(ErrSessionNotConnected)
	}
}

func (s *AyiSession) startWrite() {
	if s.Conn != nil {
		go func() {
			for writeMsg := range s.writeChan {

				_, err := proto.WriteBytes(writeMsg.Data, s.Conn)

				if err != nil {
					s.eventChan <- &SessionEvent{Id: ERROR_EVENT, Object: err}
				}

				if writeMsg.C != nil {
					writeMsg.C <- err == nil
				}
			}
		}()
	} else {
		panic(ErrSessionNotConnected)
	}
}

func (s *AyiSession) startTicker() (stop chan bool) {

	stop = make(chan bool, 1)

	go func() {

		s.ticker = time.NewTicker(5 * time.Second)
		exit := false

		for !exit {
			select {
			case <-s.ticker.C:
				s.eventChan <- &SessionEvent{Id: IDLE_EVENT}
			case <-stop:
				exit = true
			}
		}

		log.Println("Ticker stopped")

	}()

	return stop
}

// Do a read and blocks until its done or an error happens
func (s *AyiSession) doRead() {
	packet, err := proto.ReadPacket(s.Conn) // Blocked here
	if err == nil {
		s.lastRecvMsg = time.Now()
		s.pingTime = s.lastRecvMsg.Add(PING_INTERVAL_MS)
		s.eventChan <- &SessionEvent{Id: READ_MESSAGE_EVENT, Object: packet}
	} else {
		if err == proto.ErrConnectionClosed || s.closed {
			peer := !s.closed
			if peer {
				s.Exit()
			}
			s.eventChan <- &SessionEvent{Id: CONNECTION_CLOSED_EVENT, Object: peer}
		} else {
			s.eventChan <- &SessionEvent{Id: ERROR_EVENT, Object: err}

		}
	}
}

func (s *AyiSession) keepAlive() {

	current_time := time.Now()

	if !s.IsAuth {
		if current_time.After(s.lastRecvMsg.Add(MAX_LOGIN_TIME)) {
			log.Println("Connection IDLE", s)
			s.Exit()
		}
	} else {
		if current_time.After(s.lastRecvMsg.Add(MAX_IDLE_TIME)) {
			log.Println("Connection IDLE", s)
			s.Exit()
		} else if current_time.After(s.pingTime) {
			s.ping()
			log.Println("< PING to", s)
		}
	}
}

func (s *AyiSession) ping() {
	ping_msg := proto.NewMessage().Ping().Marshal()
	s.Write(ping_msg)
	s.pingTime = time.Now().Add(PING_RETRY_INTERVAL_MS)
}
