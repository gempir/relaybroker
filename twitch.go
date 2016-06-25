package main

import "net"

// Connection simple type for connection to twitch
type Connection struct {
	msgCount int // number of messages in 30 seconds
	conn     net.Conn
}

func newConnection(nick, pass string) {

}
