package protocol

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"syscall"
	"time"
)

const (
	MAX_WRITE_TIMEOUT = 15 * time.Second
	MAX_PAYLOAD_SIZE  = 1024000 // 1 Mb
	VERSION_1         = 0
	VERSION_2         = 1
)

func NewPacket(version uint8) *PacketBuilder {
	mb := &PacketBuilder{}
	if version == 0 {
		mb.message = newPacketV1()
	} else {
		mb.message = newPacketV2()
	}
	return mb
}

func getError(err error) (protoerror error) {

	oe, ok := err.(*net.OpError)

	switch {
	case err == io.EOF:
		protoerror = ErrConnectionClosed
	case ok && oe.Err == syscall.ECONNRESET:
		protoerror = ErrConnectionClosed
	case ok && oe.Timeout():
		protoerror = ErrTimeout
	default:
		protoerror = err
	}

	return
}

func WriteBytes(data []byte, conn net.Conn) (int, error) {

	if conn == nil {
		return -1, ErrInvalidSocket
	}

	conn.SetWriteDeadline(time.Now().Add(MAX_WRITE_TIMEOUT))
	n, err := conn.Write(data)

	if err != nil {
		return n, getError(err)
	}

	return n, err
}

// Reads a packet from net.Conn. This function reads packets with header formated
// as v1 or v2.
func ReadPacket(conn net.Conn) (*AyiPacket, error) {

	if conn == nil {
		return nil, ErrInvalidSocket
	}

	reader := bufio.NewReaderSize(conn, 1500)

	// FIXME: // I'm creating a lot of memory each time. GC will have to work hard
	packet := &AyiPacket{}

	// Read header
	header, err := readHeader(reader)
	if err != nil {
		protoerror := getError(err)
		return nil, protoerror
	}

	packet.Header = header

	// Read Payload
	payload_size := packet.Header.GetSize()

	if payload_size > MAX_PAYLOAD_SIZE {
		return nil, ErrMaxPayloadExceeded
	}

	if payload_size > 0 {
		packet.Data = make([]uint8, payload_size)
		err := readLimit(reader, payload_size, packet.Data)
		if err != nil {
			protoerror := getError(err)
			return nil, protoerror
		}
	}

	return packet, nil
}

func readLimit(reader io.Reader, limit uint, data []byte) error {

	var total_read uint

	for total_read != limit {
		n, err := reader.Read(data[total_read:])
		if err != nil {
			protoerror := getError(err)
			return protoerror
		}
		total_read += uint(n)
	}

	return nil
}

// In header v1 first byte is version. However, this byte
// is header size in header v2. So max.size for header is 256 bytes.
func readHeader(reader *bufio.Reader) (AyiHeader, error) {

	packet_version, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if err := reader.UnreadByte(); err != nil {
		return nil, err
	}

	if packet_version == 0 {
		return readHeaderV1(reader)
	} else {
		return readHeaderV2(reader)
	}
}

func readHeaderV1(reader *bufio.Reader) (AyiHeader, error) {

	header := &AyiHeaderV1{}

	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func readHeaderV2(reader *bufio.Reader) (AyiHeader, error) {

	header := &AyiHeaderV2{}

	err := header.ParseHeader(reader)
	if err != nil {
		return nil, err
	}

	return header, nil
}

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
		message = &AccessToken{}
	case M_CHANGE_PROFILE_PICTURE:
		message = &UserAccount{}
	case M_CHANGE_EVENT_PICTURE:
		message = &ModifyEvent{}
	case M_SYNC_GROUPS:
		message = &SyncGroups{}
	case M_HELLO:
		message = &Hello{}
	case M_IID_TOKEN:
		message = &InstanceIDToken{}

	// Requests
	case M_PING:
		message = &TimeInfo{}
	case M_READ_EVENT:
		message = &ReadEvent{}
	/*case M_LIST_AUTHORED_EVENTS:
		message = &ListCursor{}*/
	/*case M_LIST_PRIVATE_EVENTS:
		message = &ListCursor{}*/
	/*case M_LIST_PUBLIC_EVENTS:
		message = &ListPublicEvents{}*/
	/*case M_HISTORY_AUTHORED_EVENTS:
		fallthrough*/
	case M_HISTORY_PRIVATE_EVENTS:
		message = &EventListRequest{}
	/*case M_HISTORY_PUBLIC_EVENTS:
		message = &ListCursor{}*/
		///case M_USER_FRIENDS: UserFriends has no payload
	// Replies
	case M_PONG:
		message = &TimeInfo{}
	}

	return message
}
