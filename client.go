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
	spl := strings.SplitN(line, " ", 2)
	// irc command
	switch spl[0] {
		case "PASS":
			c.pass = spl[1]
	}
}
