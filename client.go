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
	spl := strings.Split(line, " ")
	command := spl[0]
	switch command {
	case "PASS":
		c.pass = spl[1]
	}
}
