package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"net"
	"net/textproto"
	"os"

	log "github.com/sirupsen/logrus"
)

// Server who handles incoming messages to relaybroker from a client
type Server struct {
	ln   net.Listener
	conn net.Conn
}

func (s *Server) startServer() {
	// RAW TCP
	ln, err := net.Listen("tcp", ":3333")
	if err != nil {
		log.Fatal("tcp server not starting", err)
	}
	defer ln.Close()
	log.Info("RAW TCP Server online")

	go s.startTLSServer()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		go s.handleClient(newClient(conn))
	}
}

func (s *Server) startTLSServer() {
	cert, err := tls.LoadX509KeyPair("/Users/gempir/certs/server.pem", "/Users/gempir/certs/server.key")
	if err != nil {
		log.Errorf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	config.Rand = rand.Reader

	listener, err := tls.Listen("tcp", "0.0.0.0:3334", &config)
	if err != nil {
		log.Errorf("server: listen: %s", err)
	}
	log.Info("TLS Server online")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Infof("server: accept: %s", err)
			break
		}
		defer conn.Close()

		go s.handleClient(newClient(conn))
	}
}

func (s *Server) stopServer() {
	s.ln.Close()
}

func (s *Server) handleClient(c Client) {
	log.Info("new client: " + c.incomingConn.RemoteAddr().String())
	r := bufio.NewReader(c.incomingConn)
	tp := textproto.NewReader(r)
	c.init()

	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Error("closing client", c.incomingConn.RemoteAddr().String(), err)
			c.bot.clientConnected = false
			c.close()
			return
		}
		c.handleMessage(line)
	}
}
