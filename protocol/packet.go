package protocol

import (
	"bytes"
	"encoding/binary"
	proto "github.com/golang/protobuf/proto"
	"log"
)

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
