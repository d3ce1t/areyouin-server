package protocol

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"syscall"
)

func Listen(net_proto, laddr string) (*AyiListener, error) {
	listener := &AyiListener{}
	var err error
	listener.socket, err = net.Listen(net_proto, laddr)
	return listener, err
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
	oe, ok := err.(*net.OpError)

	// Manage Error
	if err == io.EOF || ok && oe.Err == syscall.ECONNRESET {
		log.Println("Connection closed by client")
		return nil
	} else if err != nil {
		packet = nil
		log.Fatal("Parsing message error: ", err)
	}

	// Read Payload
	packet.Data = make([]uint8, packet.Header.Size-6)
	_, err = reader.Read(packet.Data)
	oe, ok = err.(*net.OpError)

	// Manage Error
	if err == io.EOF || ok && oe.Err == syscall.ECONNRESET {
		log.Println("Connection closed by client")
		return nil
	} else if err != nil {
		packet = nil
		log.Fatal("Parsing message error: ", err)
	}

	return packet
}

// Needed by Listener
type Callback func(PacketType, Message, *AyiClient)

type Message interface {
	Reset()
	String() string
	ProtoMessage()
}

// Used by listener.go
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
		message = &CancelUsersInvitation{}
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
		///case M_USER_FRIENDS: UserFriends has no payload
	}

	return message
}
