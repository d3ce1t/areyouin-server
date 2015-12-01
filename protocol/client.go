package protocol

import (
	"net"
	"time"
)

/*
 TODO: This implementation does not make sense. It's only a wrapper to add
 IsAuthenticated and UserId methods to the socket. Maybe it is better to create
 an AyiSession
*/
type AyiClient struct {
	socket        net.Conn
	user_id       uint64
	authenticated bool
}

func (c *AyiClient) String() string {
	return c.socket.RemoteAddr().String()
}

func (c *AyiClient) IsAuthenticated() bool {
	return c.authenticated
}

func (c *AyiClient) UserId() uint64 {
	return c.user_id
}

func (c *AyiClient) SetAuthenticated(flag bool) {
	c.authenticated = flag
}

func (c *AyiClient) SetUserId(user_id uint64) {
	c.user_id = user_id
}

// Read reads data from the connection.
// Read can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *AyiClient) Read(b []byte) (n int, err error) {
	return c.socket.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *AyiClient) Write(b []byte) (n int, err error) {
	return c.socket.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *AyiClient) Close() error {
	return c.socket.Close()
}

// LocalAddr returns the local network address.
func (c *AyiClient) LocalAddr() net.Addr {
	return c.socket.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *AyiClient) RemoteAddr() net.Addr {
	return c.socket.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future I/O, not just
// the immediately following call to Read or Write.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *AyiClient) SetDeadline(t time.Time) error {
	return c.socket.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (c *AyiClient) SetReadDeadline(t time.Time) error {
	return c.socket.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *AyiClient) SetWriteDeadline(t time.Time) error {
	return c.socket.SetWriteDeadline(t)
}
