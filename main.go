package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"sync/atomic"
)

func main() {
	log.SetOutput(os.Stdout)
	TCPServer()
}

// Connection stores messages sent in the last 30 seconds and the connection itself
type Connection struct {
	conn     net.Conn
	messages int32
}

// NewConnection initialize a Connection struct
func NewConnection(conn net.Conn) Connection {
	return Connection{
		conn:     conn,
		messages: 0,
	}

}

func (connection *Connection) reduceConnectionMessages() {
	atomic.AddInt32(&connection.messages, -1)
}

// Message called everytime you send a message
func (connection *Connection) Message(message string) {
	log.Println(connection.conn, message)
	fmt.Fprintf(connection.conn, "%s\r\n", message)
	atomic.AddInt32(&connection.messages, 1)
	time.AfterFunc(30*time.Second, connection.reduceConnectionMessages)
}
