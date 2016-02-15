package main

import (
	"crypto/tls"
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
	Id     uint16
	Data   []byte
	Future *Future
}

type Future struct {
	C          chan bool // Future called when the message is sent or, if required, when the message is acknowledged
	RequireAck bool
}

func NewFuture(ack bool) *Future {
	return &Future{
		C:          make(chan bool, 1),
		RequireAck: ack,
	}
}

// Creates a new sessions with an already connected client
func NewSession(conn net.Conn, server *Server) *AyiSession {

	session := &AyiSession{
		Conn:        conn,
		Version:     0, // Use v1 by default
		UserId:      0,
		IsAuth:      false,
		eventChan:   make(chan *SessionEvent, EVENT_CHANNEL_SIZE),
		writeChan:   make(chan *WriteMsg, WRITE_CHANNEL_SIZE),
		Server:      server,
		lastRecvMsg: time.Now().UTC(),
		pendingResp: make(map[uint16]chan bool),
	}

	return session
}

type AyiSession struct {
	Conn             net.Conn
	UserId           uint64
	Version          uint8
	IsAuth           bool
	eventChan        chan *SessionEvent
	writeChan        chan *WriteMsg
	ticker           *time.Ticker
	Server           *Server
	IIDToken         string
	closed           bool
	lastRecvMsg      time.Time
	pendingResp      map[uint16]chan bool
	outputMsgCounter uint16 // Reseat each 65535 messages
	OnRead           func(s *AyiSession, packet *proto.AyiPacket)
	OnError          func(s *AyiSession, err error)
	OnClosed         func(s *AyiSession, peer bool)
	pingTime         time.Time
}

func (s *AyiSession) IsClosed() bool {
	return s.closed
}

func (s *AyiSession) String() string {
	return s.Conn.RemoteAddr().String()
}

func (s *AyiSession) NewMessage() proto.MessageBuilder {
	return proto.NewPacket(s.Version)
}

// Alias for WriteAsync without indicating a future
func (s *AyiSession) Write(packet ...*proto.AyiPacket) (ok bool) {
	return s.WriteAsync(nil, packet...)
}

func (s *AyiSession) WriteSync(packet *proto.AyiPacket) bool {
	var ok bool
	future := NewFuture(false)
	if ok = s.WriteAsync(future, packet); ok {
		ok = <-future.C
	}
	return ok
}

func (s *AyiSession) WriteAsync(future *Future, packets ...*proto.AyiPacket) (ok bool) {

	defer func() {
		if r := recover(); r != nil {
			ok = false
			log.Printf("Session %v Write Error: %v\n", s, r)
		}
	}()

	data := make([]byte, 0, packets[0].Header.GetSize())

	for _, packet := range packets {
		packet.Header.SetToken(s.outputMsgCounter)
		data = append(data, packet.Marshal()...)
	}

	// may panic if writeChan is closed
	s.writeChan <- &WriteMsg{
		Id:     s.outputMsgCounter,
		Data:   data,
		Future: future,
	}

	s.outputMsgCounter++

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

	//log.Println("Event loop stopped")
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
				if writeMsg.Future != nil && writeMsg.Future.RequireAck {
					s.doWriteWithAck(writeMsg)
				} else {
					s.doWrite(writeMsg)
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

		//log.Println("Ticker stopped")

	}()

	return stop
}

// Do a read and blocks until its done or an error happens
func (s *AyiSession) doRead() {

	packet, err := proto.ReadPacket(s.Conn) // Blocked here

	if err == nil {
		s.lastRecvMsg = time.Now()
		s.pingTime = s.lastRecvMsg.Add(PING_INTERVAL_MS)
		s.Version = uint8(packet.Version())

		if !s.manageSessionMsg(packet) {
			s.eventChan <- &SessionEvent{Id: READ_MESSAGE_EVENT, Object: packet}
		}

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

func (s *AyiSession) doWrite(msg *WriteMsg) {
	_, err := proto.WriteBytes(msg.Data, s.Conn)
	if err != nil {
		s.eventChan <- &SessionEvent{Id: ERROR_EVENT, Object: err}
	}
	if msg.Future != nil && msg.Future.C != nil {
		msg.Future.C <- err == nil
	}
}

func (s *AyiSession) doWriteWithAck(msg *WriteMsg) {

	waitResponse := make(chan bool, 1)
	s.pendingResp[msg.Id] = waitResponse

	defer func() {
		delete(s.pendingResp, msg.Id)
	}()

	_, err := proto.WriteBytes(msg.Data, s.Conn)

	if err != nil {
		s.eventChan <- &SessionEvent{Id: ERROR_EVENT, Object: err}
		msg.Future.C <- false
		return
	}

	timeout := time.NewTicker(10 * time.Second)

	select {
	case <-waitResponse:
	case <-timeout.C:
		err = proto.ErrTimeout
	}

	timeout.Stop()
	msg.Future.C <- err == nil

	if err != nil {
		s.eventChan <- &SessionEvent{Id: ERROR_EVENT, Object: err}
	}
}

// Manage session messages. Returns true if message has been
// managed or false otherwise
func (s *AyiSession) manageSessionMsg(packet *proto.AyiPacket) bool {

	if packet.Header.GetType() == proto.M_USE_TLS {
		log.Printf("> (%v) USE TLS\n", s)
		if tlsconn, ok := s.Conn.(*tls.Conn); !ok {
			log.Printf("Changing to use TLS for session %v\n", s)
			if tlsconn = tls.Server(s.Conn, s.Server.TLSConfig); tlsconn != nil {
				//oldConn := s.Conn
				s.Conn = tlsconn
				//oldConn.Close()
				//defer conn.Close()
			}
		}
		return true
	}

	if packet.Header.GetType() == proto.M_OK && packet.Header.GetToken() != 0 {
		if c, ok := s.pendingResp[packet.Header.GetToken()]; ok {
			log.Printf("> (%v) ACK %v\n", s.UserId, packet.Header.GetToken())
			c <- true
		}
		return true
	}

	return false
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
			log.Printf("< (%v) PING", s.UserId)
		}
	}
}

func (s *AyiSession) ping() {
	s.Write(s.NewMessage().Ping())
	s.pingTime = time.Now().Add(PING_RETRY_INTERVAL_MS)
}
