package main

import "net"

// Connection simple type for connection to twitch
type Connection struct {
	msgCount      int // number of messages in 30 seconds
	conn          net.Conn
	channels      []string
	authenticated bool
	alive         bool
}

func newConnection(nick string, pass string) Connection {
	return Connection{
		msgCount:      0,
		conn:          nil,
		channels:      nil,
		authenticated: false,
		alive:         false,
	}
}

func connect(ct Connection) {

}
