package main

import (
	"net"
)

func NewSession(conn net.Conn) *AyiSession {
	return &AyiSession{
		Conn:   conn,
		UserId: 0,
		IsAuth: false,
	}
}

type AyiSession struct {
	Conn   net.Conn
	UserId uint64
	IsAuth bool
}

func (c *AyiSession) String() string {
	return c.Conn.RemoteAddr().String()
}
