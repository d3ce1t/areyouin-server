package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"log"
	"runtime/debug"
)

type AyiHeader struct { // 6 bytes
	Version uint8
	Token   uint16
	Type    PacketType
	Size    uint16
}

func (h *AyiHeader) String() string {
	return fmt.Sprint("Version:", h.Version, "Token:", h.Token, "Type:", h.Type, "Size:", h.Size)
}

// An AyiPacket is a network container for a message
type AyiPacket struct {
	Header AyiHeader
	Data   []uint8 // Holds a message encoded as binary data
}

func (packet *AyiPacket) String() string {
	str := fmt.Sprintf("Header {%s}\n", packet.Header.String())
	str += fmt.Sprintln("Data:", hex.EncodeToString(packet.Data))
	return str
}

func (packet *AyiPacket) Type() PacketType {
	return packet.Header.Type
}

// Decodes a packet in order to get a message. If the message
// is unknown a nil message is returned
func (packet *AyiPacket) DecodeMessage() Message {

	message := createEmptyMessage(packet.Type())

	if message != nil {
		err := proto.Unmarshal(packet.Data, message)

		if err != nil {
			log.Println("Unmarshaling error: ", err)
			log.Println(packet)
			return nil
		}
	}

	return message
}

func (packet *AyiPacket) SetMessage(message Message) {

	data, err := proto.Marshal(message)

	if err != nil {
		debug.PrintStack()
		log.Fatal("Marshaling error: ", err)
	}

	size := len(data)

	if size > 65530 {
		debug.PrintStack()
		log.Fatal("Message exceeds max.size of 65530 bytes")
	}

	packet.Data = data
	packet.Header.Size = 6 + uint16(size)
}

// FIXME: Change this function to use directly a write stream (avoid copy)
func (packet *AyiPacket) Marshal() []byte {

	buf := &bytes.Buffer{}

	// Write Header
	err := binary.Write(buf, binary.BigEndian, packet.Header) // X86 is LittleEndian, whereas ARM is BigEndian / Bi-Endian
	if err != nil {
		log.Println("Build message failed (1):", err)

	}

	// Write Payload
	if len(packet.Data) > 0 {
		_, err = buf.Write(packet.Data)
		if err != nil {
			log.Fatal("Build message failed (2):", err)
		}
	}

	return buf.Bytes()
}
