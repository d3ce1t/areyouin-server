package protocol

type AyiHeader interface {
	SetVersion(version uint32)
	SetToken(token uint16)
	SetType(packet_type PacketType)
	SetSize(size uint)
	GetVersion() uint32
	GetToken() uint16
	GetType() PacketType
	GetSize() uint
	String() string
	Marshall() []byte
}
