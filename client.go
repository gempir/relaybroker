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
	if strings.HasPrefix(line, "PASS ") {
        c.pass = strings.Split(line, "PASS ")[1]
	}
}
