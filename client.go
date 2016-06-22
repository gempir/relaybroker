package main

import (
	"net"
	"strings"
)

type Client struct {
	incomingConn net.Conn
	pass         string
	nick         string
}

func newClient(conn net.Conn) Client {
	return Client{
		incomingConn: conn,
	}
}

func (c *Client) handleMessage(line string) {
	log.Debug(line)
	spl := strings.SplitN(line, " ", 2)
	// irc command
	switch spl[0] {
	case "PASS":
		c.pass = spl[1]
	case "NICK":
		c.nick = strings.ToLower(spl[1]) // make sure the nick is lowercase
	}
}
