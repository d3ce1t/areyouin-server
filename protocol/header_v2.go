package protocol

import (
	proto "github.com/golang/protobuf/proto"
	"io"
)

func (h *AyiHeaderV2) SetVersion(version uint32) {
	h.Version = version
}

func (h *AyiHeaderV2) SetToken(token uint16) {
	h.Token = uint32(token)
}

func (h *AyiHeaderV2) SetType(packet_type PacketType) {
	h.Type = uint32(packet_type)
}

func (h *AyiHeaderV2) SetSize(size uint) {
	h.PayloadSize = uint32(size)
}

func (h *AyiHeaderV2) GetVersion() uint32 {
	return h.Version
}

func (h *AyiHeaderV2) GetToken() uint16 {
	return uint16(h.Token)
}

func (h *AyiHeaderV2) GetType() PacketType {
	return PacketType(h.Type)
}

func (h *AyiHeaderV2) GetSize() uint {
	return uint(h.PayloadSize)
}

func (h *AyiHeaderV2) Marshal(writer io.Writer) error {

	header_data, err := proto.Marshal(h)

	if err != nil {
		return err
	}

	n, err := writer.Write([]byte{uint8(len(header_data))})
	if err != nil {
		return err
	} else if n != 1 {
		return ErrIncompleteWrite
	}

	n, err = writer.Write(header_data)
	if err != nil {
		return err
	} else if n != len(header_data) {
		return ErrIncompleteWrite
	}

	return nil
}

func (h *AyiHeaderV2) ParseHeader(reader io.Reader) error {

	data := []byte{0}

	_, err := reader.Read(data)
	if err != nil {
		protoerror := getError(err)
		return protoerror
	}

	header_size := uint(data[0])
	header_bytes := make([]byte, header_size)

	err = readLimit(reader, header_size, header_bytes)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(header_bytes, h)
	if err != nil {
		return err
	}

	return nil
}
