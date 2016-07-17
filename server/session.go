package main

import (
	"crypto/tls"
	"log"
	"net"
	proto "peeple/areyouin/protocol"
	core "peeple/areyouin/common"
	"time"
	"fmt"
)

const (
	// Times
	MAX_IDLE_TIME          = 30 * time.Minute
	MAX_LOGIN_TIME         = 30 * time.Second
	PING_INTERVAL_MS       = 29 * time.Minute
	PING_RETRY_INTERVAL_MS = 1 * time.Minute
	TICKER_INTERVAL        = 30 * time.Second

	// Channels Sizes
	//WRITE_CHANNEL_SIZE = 5
)

type WriteMsg struct {
	Packet *proto.AyiPacket
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
		readReadyChan:   make(chan *proto.AyiPacket, 1),
		errorChan:       make(chan error, 1),
		exitChan:        make(chan bool, 1),
		writeChan:       make(chan *WriteMsg, 1),
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
	IIDToken         *core.IIDToken
	closed           bool
	lastRecvMsg      time.Time
	pendingResp      map[uint16]chan bool
	nextToken        uint16 // most significant bit reserved (1 -> Response, 0 -> Normal). Reset each 32768 messages (0 - 32767)
	OnRead           func(s *AyiSession, packet *proto.AyiPacket)
	OnError          func(s *AyiSession, err error)
	OnClosed         func(s *AyiSession, peer bool)
	pingTime         time.Time
}

func (s *AyiSession) String() string {
	if s.IsAuth {
		return fmt.Sprintf("%v", s.UserId)
	} else {
		return s.Conn.RemoteAddr().String()
	}
}

func (s *AyiSession) IsClosed() bool {
	return s.closed
}

func (s *AyiSession) NewMessage() proto.MessageBuilder {
	return proto.NewPacket(s.ProtocolVersion)
}

// Alias for WriteAsync without indicating a future
func (s *AyiSession) Write(packet *proto.AyiPacket) (ok bool) {
	return s.WriteAsync(nil, packet)
}

func (s *AyiSession) WriteSync(packet *proto.AyiPacket) bool {
	var ok bool
	future := NewFuture(false)
	if ok = s.WriteAsync(future, packet); ok {
		ok = <-future.C
	}
	return ok
}

func (s *AyiSession) WriteResponse(token uint16, packet *proto.AyiPacket) (ok bool) {

	defer func() {
		if r := recover(); r != nil {
			ok = false
			log.Printf("Session %v Write Error: %v\n", s, r)
		}
	}()

	// Set most significant byte to 1 in order to mark
	// packet to be sent as a response to packet with given token
	var tokenResponse uint16 = token | (1 << 15)
	packet.Header.SetToken(tokenResponse)

	// may panic if writeChan is closed
	s.writeChan <- &WriteMsg{
		Packet:   packet,
		Future: nil,
	}

	return true
}

func (s *AyiSession) WriteAsync(future *Future, packet *proto.AyiPacket) (ok bool) {

	defer func() {
		if r := recover(); r != nil {
			ok = false
			log.Printf("Session %v Write Error: %v\n", s, r)
		}
	}()

	packet.Header.SetToken(s.nextToken)

	// may panic if writeChan is closed
	s.writeChan <- &WriteMsg{
		Packet:   packet,
		Future: future,
	}

	// Update next token
	s.nextToken = (s.nextToken + 1) % (1 << 15)

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
	s.ticker = time.NewTicker(TICKER_INTERVAL)

	defer func() {
		s.ticker.Stop()
		close(s.writeChan)
		close(s.readReadyChan)
		close(s.errorChan)
		close(s.exitChan)
		s.IsAuth = false
		s.UserId = 0
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
			s.closeSocket()
			s.OnClosed(s, peerClosed)
			return true
		}
	}
}

func (s *AyiSession) Exit() {
	if !s.closed {
		s.closed = true
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
		go func () {
			for !s.closed {
				for writeMsg := range s.writeChan {
					s.manageWrite(writeMsg)
				}
			}
		}()
	} else {
		panic(ErrSessionNotConnected)
	}
}

// Do a read and blocks until its done or an error happens
func (s *AyiSession) doRead() {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Session %v doRead panic: %v\n", s, r)
		}
	}()

	packet, err := proto.ReadPacket(s.Conn) // Blocked here

	if err == nil {

		// Read

		if !s.manageSessionMsg(packet) {
			s.readReadyChan <- packet
		}

	} else {

		// Manage Error

		if err == proto.ErrConnectionClosed || s.closed {

			// If socket was closed by remote peer, then write to exitChan. Otherwise,
			// ignore it because Exit() already writes to that channel.

			if !s.closed {
				s.closed = true
				s.exitChan <- true
			}
		} else {
			s.errorChan <- err
		}
	}
}

func (s *AyiSession) doWrite(msg *WriteMsg) {

	_, err := proto.WriteBytes(msg.Packet.Marshal(), s.Conn)
	if err != nil {
		s.errorChan <- err
	}
	if msg.Future != nil && msg.Future.C != nil {
		msg.Future.C <- (err == nil)
	}
}

func (s *AyiSession) doWriteWithAck(msg *WriteMsg) {

	waitResponse := make(chan bool, 1)
	s.pendingResp[msg.Packet.Id()] = waitResponse
	log.Printf("* (%v) Register write with ACK for packet %v\n", s, msg.Packet.Id())

	_, err := proto.WriteBytes(msg.Packet.Marshal(), s.Conn)
	if err != nil {
		delete(s.pendingResp, msg.Packet.Id())
		close(waitResponse)
		s.errorChan <- err
		msg.Future.C <- false
		return
	}

	go func() {
		var err error
		// TODO: Change ticker by timer
		timeout := time.NewTicker(10 * time.Second)

		defer func() {
			delete(s.pendingResp, msg.Packet.Id())
			close(waitResponse)
			timeout.Stop()

			msg.Future.C <- (err == nil)
			if err != nil {
				s.errorChan <- err
			}
		}()

		select {
		case <-waitResponse:
		case <-timeout.C:
			err = proto.ErrTimeout
		}
	}()
}

func (s *AyiSession) manageWrite(writeMsg *WriteMsg) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Session %v manageWrite panic: %v\n", s, r)
		}
	}()

	if writeMsg.Future != nil && writeMsg.Future.RequireAck {
			s.doWriteWithAck(writeMsg)
	} else {
			s.doWrite(writeMsg)
	}
}

// Manage session messages. Returns true if message has been
// managed or false otherwise
func (s *AyiSession) manageSessionMsg(packet *proto.AyiPacket) bool {

	s.lastRecvMsg = time.Now()
	s.pingTime = s.lastRecvMsg.Add(PING_INTERVAL_MS)

	if packet.Version() > 0 {

		// Before HELLO message was introduced, the protocol version was set from the
		// received packet each time. Now protocol version is set by HELLO message.
		// For compatibility purposes keep doing the same thing but get it from the first
		// received packet so that version from a posterior HELLO message doesn't get overrided.

		if s.ProtocolVersion == 0 {
			s.ProtocolVersion = uint8(packet.Version())
		}

	} else {
		// Do not allow connections from protocol version 0 (i.e AreYouIn Android/1.0.7 and lower)
		log.Printf("* (%v) Session is gonna be closed because of a deprecated app version being used\n", s)
		s.Exit()
		return true
	}

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
	if packet.Header.GetType() == proto.M_OK && !packet.HasPayload() {
		if c, ok := s.pendingResp[packet.Id()]; ok {
			log.Printf("> (%v) ACK for packet with id %v\n", s.UserId, packet.Id())
			c <- true
		}
		return true
	}

	// PONG
	if packet.Header.GetType() == proto.M_PONG {
		log.Printf("> (%v) PONG\n", s)
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
