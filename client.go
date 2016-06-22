package main

import (
	"net"
	"strings"
)

type Client struct {
	incomingConn net.Conn
	pass         string
	username     string
}

func newClient(conn net.Conn) Client {
	return Client{
		incomingConn: conn,
	}
}

func (c *Client) handleMessage(line string) {
	log.Debug(line)
	// spl[0] is the command and spl[1] the rest of the message
	spl := strings.SplitN(line, " ", 2)
	switch spl[0] {
	case "PASS":
		c.pass = spl[1]
	}
}
