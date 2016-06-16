package main

import (
	"crypto/tls"
	"log"
	"net"
	proto "peeple/areyouin/protocol"
	"time"
	"fmt"
)

const (

	// Times
	MAX_IDLE_TIME          = 30 * time.Minute
	MAX_LOGIN_TIME         = 30 * time.Second
	PING_INTERVAL_MS       = 29 * time.Minute
	PING_RETRY_INTERVAL_MS = 20 * time.Second

	// Channels Sizes
	WRITE_CHANNEL_SIZE = 3
)

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
		Conn:            conn,
		ProtocolVersion: 0, // Use v1 by default
		UserId:          0,
		IsAuth:          false,
		readReadyChan:   make(chan *proto.AyiPacket),
		errorChan:       make(chan error),
		exitChan:        make(chan bool),
		writeChan:       make(chan *WriteMsg, WRITE_CHANNEL_SIZE),
		Server:          server,
		lastRecvMsg:     time.Now().UTC(),
		pendingResp:     make(map[uint16]chan bool),
	}

	return session
}

type AyiSession struct {
	Conn   net.Conn
	UserId int64

	// Network protocol version
	ProtocolVersion uint8

	// Client version
	ClientVersion string

	// Platform
	Platform string

	// Platform Version
	PlatformVersion string

	IsAuth           bool
	readReadyChan    chan *proto.AyiPacket
	errorChan        chan error
	exitChan         chan bool
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
	if s.IsAuth {
		return fmt.Sprintf("%v", s.UserId)
	} else {
		return s.Conn.RemoteAddr().String()
	}
}

func (s *AyiSession) NewMessage() proto.MessageBuilder {
	return proto.NewPacket(s.ProtocolVersion)
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
	s.startWrite()
	s.ticker = time.NewTicker(5 * time.Second)

	defer func() {
		s.ticker.Stop()
		close(s.writeChan)
		close(s.readReadyChan)
		close(s.errorChan)
		close(s.exitChan)
		s.IsAuth = false
	}()

	s.pingTime = time.Now().Add(PING_INTERVAL_MS) // Move to session
	exit := false

	for !exit {
		exit = s.eventLoop() // if panic returns false
	}
}

func (s *AyiSession) eventLoop() (exit bool) {

	defer func() {
		if r := recover(); r != nil {
			exit = s.closed
			log.Printf("Session %v EventLoop Panic: %v\n", s, r)
		}
	}()

	for {

		select {
		case packet := <-s.readReadyChan:
			s.OnRead(s, packet)
		case err := <- s.errorChan:
			s.OnError(s, err)
		case <- s.ticker.C:
			s.keepAlive()
		case peerClosed := <- s.exitChan:
			if peerClosed {
				s.closeSocket()
			}
			s.OnClosed(s, peerClosed)
			return true
		}
	}

}

func (s *AyiSession) Exit() {
	if !s.closed {
		s.closed = true
		s.closeSocket()
		go func() {
			s.exitChan <- false
		}()
	}
}

func (s *AyiSession) closeSocket() {
	if err := s.Conn.Close(); err != nil {
		log.Printf("* (%v) Socket error while closing: %v", s, err)
	} else {
		log.Printf("* (%v) Socket closed", s)
	}
}

// Read socket in background and send result through EventLoopChan
func (s *AyiSession) startRecv() {
	if s.Conn != nil {
		go func() {
			// FIXME: Capture Panic here or in doRead()
			for !s.closed {
				s.doRead()
			}
			//log.Println("Read Loop Finished")
		}()
	} else {
		panic(ErrSessionNotConnected)
	}
}

func (s *AyiSession) startWrite() {
	if s.Conn != nil {
		go func() {
			// Loop that will run until s.writeChan gets closed
			for writeMsg := range s.writeChan {
				if writeMsg.Future != nil && writeMsg.Future.RequireAck {
					s.doWriteWithAck(writeMsg)
				} else {
					s.doWrite(writeMsg)
				}
			}
			//log.Println("Write Loop Finished")
		}()
	} else {
		panic(ErrSessionNotConnected)
	}
}

// Do a read and blocks until its done or an error happens
func (s *AyiSession) doRead() {

	packet, err := proto.ReadPacket(s.Conn) // Blocked here

	if err == nil {

		// Read

		s.lastRecvMsg = time.Now()
		s.pingTime = s.lastRecvMsg.Add(PING_INTERVAL_MS)

		// Before HELLO message was introduced, the protocol version was set from the
		// received packet each time. Now protocol version is set by HELLO message.
		// For compatibility purposes keep doing the same thing but get it from the first
		// received packet so that version from a posterior HELLO message doesn't get overrided.
		if s.ProtocolVersion == 0 {
			s.ProtocolVersion = uint8(packet.Version())
		}

		if !s.manageSessionMsg(packet) {
			s.readReadyChan <- packet
		}

	} else {

		// Manage Error

		if err == proto.ErrConnectionClosed || s.closed {

			// If socket was closed by remote peer, then write to exitChan. Otherwise,
			// ignore it because Exit() already writes to that channel.

			if !s.closed {
				s.exitChan <- true
			}
		} else {
			s.errorChan <- err
		}
	}
}

func (s *AyiSession) doWrite(msg *WriteMsg) {
	_, err := proto.WriteBytes(msg.Data, s.Conn)
	if err != nil {
		s.errorChan <- err
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
		s.errorChan <- err
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
		s.errorChan <- err
	}
}

// Manage session messages. Returns true if message has been
// managed or false otherwise
func (s *AyiSession) manageSessionMsg(packet *proto.AyiPacket) bool {

	// USE TLS
	if packet.Header.GetType() == proto.M_USE_TLS {
		log.Printf("> (%v) USE TLS\n", s)
		if tlsconn, ok := s.Conn.(*tls.Conn); !ok {
			log.Printf("* (%v) Changing to use TLS\n", s)
			if tlsconn = tls.Server(s.Conn, s.Server.TLSConfig); tlsconn != nil {
				s.Conn = tlsconn
			}
		} else {
			log.Printf("* (%v) Error changing to use TLS\n", s)
		}
		return true
	}

	// HELLO
	if packet.Header.GetType() == proto.M_HELLO {
		generic_message, _ := packet.DecodeMessage()
		if generic_message != nil {
			hello_info := generic_message.(*proto.Hello)
			s.ProtocolVersion = uint8(hello_info.ProtocolVersion)
			s.ClientVersion = hello_info.ClientVersion
			s.Platform = hello_info.Platform
			s.PlatformVersion = hello_info.PlatformVersion
			log.Printf("> (%v) HELLO %v\n", s, hello_info)
		}
		return true
	}

	// REQUIRE ACK
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
