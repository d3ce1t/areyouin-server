package protocol

import (
	"bytes"
	"encoding/binary"
	proto "github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
)

func Listen(net_proto, laddr string) (*AyiListener, error) {
	listener := &AyiListener{}
	var err error
	listener.socket, err = net.Listen(net_proto, laddr)
	return listener, err
}

type AyiListener struct {
	socket    net.Listener
	callbacks map[PacketType]Callback
}

// Accept waits for and returns the next connection to the listener.
func (l *AyiListener) Accept() (c net.Conn, err error) {
	return l.socket.Accept()
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *AyiListener) Close() error {
	return l.socket.Close()
}

// Addr returns the listener's network address.
func (l *AyiListener) Addr() net.Addr {
	return l.socket.Addr()
}

func (l *AyiListener) RegisterCallback(command PacketType, f Callback) {
	if l.callbacks == nil {
		l.callbacks = make(map[PacketType]Callback)
	}
	l.callbacks[command] = f
}

func (l *AyiListener) ServeMessage(packet *AyiPacket, client net.Conn) error {

	message := createEmptyMessage(packet.Type())

	if message == nil {
		log.Fatal("Unknown message", packet)
	}

	packet.DecodeMessage(message)

	if f, ok := l.callbacks[packet.Type()]; ok {
		f(packet.Type(), message, client)
	}

	return nil
}

type AyiClient struct {
}

func NewMessage() *MessageBuilder {
	mb := &MessageBuilder{}
	mb.message = &AyiPacket{}
	mb.message.Header.Version = 0
	mb.message.Header.Token = 0
	mb.message.Header.Type = M_ERROR
	mb.message.Header.Size = 6
	return mb
}

// Reads a message from an io.Reader
func ReadPacket(reader io.Reader) *AyiPacket {

	packet := &AyiPacket{}

	// Read header
	err := binary.Read(reader, binary.BigEndian, &packet.Header)

	if err == io.EOF {
		log.Println("Connection closed by client")
		return nil
	} else if err != nil {
		packet = nil
		log.Fatal("Parsing message error: ", err)
	}

	// Read Payload
	packet.Data = make([]uint8, packet.Header.Size-6)
	_, err = reader.Read(packet.Data)

	if err != nil {
		packet = nil
		log.Fatal("Parsing message error:", err)
	}

	return packet
}

type Callback func(PacketType, Message, net.Conn)

type Message interface {
	Reset()
	String() string
	ProtoMessage()
}

type AyiHeader struct { // 6 bytes
	Version uint8
	Token   uint16
	Type    PacketType
	Size    uint16
}

// An AyiPacket is a network container for a message
type AyiPacket struct {
	Header AyiHeader
	Data   []uint8 // Holds a message encoded as binary data
}

func (packet *AyiPacket) Type() PacketType {
	return packet.Header.Type
}

func (packet *AyiPacket) DecodeMessage(dst_msg Message) {

	err := proto.Unmarshal(packet.Data, dst_msg)

	if err != nil {
		log.Fatal("Unmarshaling error: ", err)
	}
}

func (packet *AyiPacket) SetMessage(message Message) {

	data, err := proto.Marshal(message)

	if err != nil {
		log.Fatal("Marshaling error: ", err)
	}

	size := len(data)

	if size > 65530 {
		log.Fatal("Message exceeds max.size of 65530 bytes")
	}

	packet.Data = data
	packet.Header.Size = 6 + uint16(size)
}

// Change this function to use directly a write stream (avoid copy)
func (packet *AyiPacket) Marshal() []byte {
	buf := new(bytes.Buffer)

	// Write Header
	err := binary.Write(buf, binary.BigEndian, packet.Header) // X86 is LittleEndian, whereas ARM is BigEndian / Bi-Endian
	if err != nil {
		log.Fatal("Build message failed:", err)
	}

	// Write Payload
	if len(packet.Data) > 0 {
		_, err = buf.Write(packet.Data)
		if err != nil {
			log.Fatal("Build message failed:", err)
		}
	}

	return buf.Bytes()
}

func createEmptyMessage(packet_type PacketType) Message {

	var message Message = nil

	switch packet_type {
	// Modifiers
	case M_CREATE_EVENT:
		message = &CreateEvent{}
	case M_CANCEL_EVENT:
		message = &CancelEvent{}
	case M_INVITE_USERS:
		message = &InviteUsers{}
	case M_CANCEL_USERS_INVITATION:
		message = &CancelUserInvitation{}
	case M_CONFIRM_ATTENDANCE:
		message = &ConfirmAttendance{}
	case M_MODIFY_EVENT_DATE:
		fallthrough
	case M_MODIFY_EVENT_MESSAGE:
		fallthrough
	case M_MODIFY_EVENT:
		message = &ModifyEvent{}
	case M_VOTE_CHANGE:
		message = &VoteChange{}
	case M_USER_POSITION:
		message = &UserPosition{}
	case M_USER_POSITION_RANGE:
		message = &UserPositionRange{}
	case M_USER_CREATE_ACCOUNT:
		message = &CreateUserAccount{}
	case M_USER_NEW_AUTH_TOKEN:
		message = &NewAuthToken{}
	case M_USER_AUTH:
		message = &UserAuthentication{}

	// Requests
	case M_PING:
		message = &Ping{}
	case M_READ_EVENT:
		message = &ReadEvent{}
	case M_LIST_AUTHORED_EVENTS:
		message = &ListCursor{}
	case M_LIST_PRIVATE_EVENTS:
		message = &ListCursor{}
	case M_LIST_PUBLIC_EVENTS:
		message = &ListPublicEvents{}
	case M_HISTORY_AUTHORED_EVENTS:
		fallthrough
	case M_HISTORY_PRIVATE_EVENTS:
		fallthrough
	case M_HISTORY_PUBLIC_EVENTS:
		message = &ListCursor{}
	}

	return message
}
