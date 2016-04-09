package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	log.SetOutput(os.Stdout)
	TCPServer()
}

// Connection stores messages sent in the last 30 seconds and the connection itself
type Connection struct {
	conn       net.Conn
	messages   int32
	channels   []string
	joins      int32
	connactive bool
}

// NewConnection initialize a Connection struct
func NewConnection(conn net.Conn) Connection {
	return Connection{
		conn:       conn,
		messages:   0,
		channels:   make([]string, 0),
		joins:      0,
		connactive: false,
	}
}

func (connection *Connection) reduceConnectionMessages() {
	atomic.AddInt32(&connection.messages, -1)
}

func (connection *Connection) reduceConnectionJoins() {
	atomic.AddInt32(&connection.joins, -1)
}

func (connection *Connection) activateConn() {
	connection.connactive = true
}

// Message called everytime you send a message
func (connection *Connection) Message(message string) {
	log.Println(connection.conn, message)
	fmt.Fprintf(connection.conn, "%s\r\n", message)
	atomic.AddInt32(&connection.messages, 1)
	time.AfterFunc(30*time.Second, connection.reduceConnectionMessages)
}

// Join controls joins
func (connection *Connection) Join(channel string) {
	log.Println(connection.conn, "JOIN "+channel)
	fmt.Fprintf(connection.conn, "JOIN %s\r\n", channel)
	atomic.AddInt32(&connection.joins, 1)
	time.AfterFunc(10*time.Second, connection.reduceConnectionJoins)
	connection.channels = append(connection.channels, channel)
}
