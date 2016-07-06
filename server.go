package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"os"
)

// Server who handles incoming messages to relaybroker from a client
type Server struct {
	ln   net.Listener
	conn net.Conn
}

func (s *Server) startServer(TCPPort int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", TCPPort))
	if err != nil {
		Log.Error(err)
		panic("tcp server not starting")
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			Log.Error(err)
			os.Exit(1)
		}
		go s.handleClient(newClient(conn))
	}
}

func (s *Server) stopServer() {
	s.ln.Close()
}

func (s *Server) handleClient(c Client) {
	Log.Info("new client: " + c.incomingConn.RemoteAddr().String())
	r := bufio.NewReader(c.incomingConn)
	tp := textproto.NewReader(r)
	c.init()

	for {
		line, err := tp.ReadLine()
		if err != nil {
			Log.Error("closing client", c.incomingConn.RemoteAddr().String(), err)
			c.close()
			return
		}
		c.handleMessage(line)
	}
}
