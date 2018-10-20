package main

import (
	"bufio"
	"net"
	"net/textproto"
	"os"
)

// Server who handles incoming messages to relaybroker from a client
type Server struct {
	ln   net.Listener
	conn net.Conn
}

func (s *Server) startServer() {
	ln, err := net.Listen("tcp", brokerHost)
	if err != nil {
		Log.Fatal("tcp server not starting", err)
	}
	defer ln.Close()
	Log.Info("started listening on", brokerHost)
	for {
		conn, err := ln.Accept()
		if err != nil {
			Log.Error(err.Error())
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
			Log.Info("closing client", c.incomingConn.RemoteAddr().String(), err)
			c.bot.clientConnected = false
			c.close()
			return
		}
		c.handleMessage(line)
	}
}
