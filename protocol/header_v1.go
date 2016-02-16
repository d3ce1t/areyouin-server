package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

type AyiHeaderV1 struct { // 6 bytes
	Version uint8
	Token   uint16 // Message ID
	Type    PacketType
	Size    uint16 // Packet + Payload size
}

func (h *AyiHeaderV1) SetVersion(version uint32) {
	h.Version = uint8(version)
}

func (h *AyiHeaderV1) SetToken(token uint16) {
	h.Token = token
}

func (h *AyiHeaderV1) SetType(packet_type PacketType) {
	h.Type = packet_type
}

func (h *AyiHeaderV1) SetSize(size uint) {
	h.Size = uint16(6 + size)
}

func (h *AyiHeaderV1) GetVersion() uint32 {
	return uint32(h.Version)
}

func (h *AyiHeaderV1) GetToken() uint16 {
	return h.Token
}

func (h *AyiHeaderV1) GetType() PacketType {
	return h.Type
}

func (h *AyiHeaderV1) GetSize() uint {
	return uint(h.Size - 6)
}

func (h *AyiHeaderV1) String() string {
	return fmt.Sprintf("Version: %v Token: %v Type: %v Size: %v\n", h.Version, h.Token, h.Type, h.Size)
}

func (h *AyiHeaderV1) Marshal(writer io.Writer) error {
	err := binary.Write(writer, binary.BigEndian, h) // X86 is LittleEndian, whereas ARM is BigEndian / Bi-Endian
	if err != nil {
		return err
	}
	return nil
}
