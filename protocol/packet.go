package protocol

import (
	"bytes"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"log"
	"runtime/debug"
)

func newPacketV2() *AyiPacket {
	packet := &AyiPacket{
		Header: &AyiHeaderV2{},
	}
	packet.Header.SetVersion(1) // Protocol version starts at 0.
	packet.Header.SetToken(0)
	packet.Header.SetType(M_ERROR)
	packet.Header.SetSize(0) // Payload size
	return packet
}

func newPacketV1() *AyiPacket {
	packet := &AyiPacket{
		Header: &AyiHeaderV1{},
	}
	packet.Header.SetVersion(0)
	packet.Header.SetToken(0)
	packet.Header.SetType(M_ERROR)
	packet.Header.SetSize(0) // Payload size (internally stores header + payload size)
	return packet
}

// An AyiPacket is a network container for a message
type AyiPacket struct {
	Header AyiHeader
	Data   []uint8 // Holds a message encoded as binary data
}

func (packet *AyiPacket) String() string {
	str := fmt.Sprintf("Header {%s} Data {%x}\n", packet.Header.String(), packet.Data)
	return str
}

func (packet *AyiPacket) Id() uint16 {
	return packet.Header.GetToken()
}

func (packet *AyiPacket) IsResponse() bool {
	return (packet.Header.GetToken() >> 15) > 0
}

func (packet *AyiPacket) ResponseId() uint16 {
	return packet.Header.GetToken() & 0x7FFF // Remove sign from token (16 bits token)
}

func (packet *AyiPacket) Type() PacketType {
	return packet.Header.GetType()
}

func (packet *AyiPacket) Version() uint32 {
	return packet.Header.GetVersion()
}

// Decodes a packet in order to get a message. If the message
// is unknown a nil message is returned
func (packet *AyiPacket) DecodeMessage() (Message, error) {

	// Check message first. Because of protocol buffers default messages
	// a packet could not convey payload but have a message with default
	// values, i.e. empty strings, zero values, and so on.
	message := createEmptyMessage(packet.Type())

	if message == nil {
		if !packet.HasPayload() {
			return nil, ErrNoPayload
		} else {
			return nil, ErrUnknownMessage
		}
	}

	if packet.HasPayload() {
		if err := proto.Unmarshal(packet.Data, message); err != nil {
			return nil, err
		}
	}

	return message, nil
}

func (packet *AyiPacket) SetMessage(message Message) {

	data, err := proto.Marshal(message)

	if err != nil {
		debug.PrintStack()
		log.Fatal("Marshaling error: ", err)
	}

	size := len(data)

	if size > MAX_PAYLOAD_SIZE {
		log.Printf("WARNING: Message exceeds max.size of %v bytes\n", MAX_PAYLOAD_SIZE)
	}

	packet.Data = data
	packet.Header.SetSize(uint(size))
}

// FIXME: Change this function to use directly a write stream (avoid copy)
func (packet *AyiPacket) Marshal() []byte {

	buf := &bytes.Buffer{}

	// Write Header
	err := packet.Header.Marshal(buf)
	if err != nil {
		log.Println("Build message failed (1):", err)
		return nil
	}

	// Write Payload
	if len(packet.Data) > 0 {
		_, err = buf.Write(packet.Data)
		if err != nil {
			log.Fatal("Build message failed (2):", err)
			return nil
		}
	}

	return buf.Bytes()
}

func (packet *AyiPacket) HasPayload() bool {
	return packet.Header.GetSize() > 0
}
