package protocol

import (
	"log"
	"net"
)

type AyiListener struct {
	socket    net.Listener
	callbacks map[PacketType]Callback
}

// Accept waits for and returns the next connection to the listener.
func (l *AyiListener) Accept() (c *AyiClient, err error) {
	client := &AyiClient{}
	socket, e := l.socket.Accept()
	client.socket = socket
	client.authenticated = false
	client.user_id = ""
	return client, e
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

func (l *AyiListener) ServeMessage(packet *AyiPacket, client *AyiClient) error {

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
