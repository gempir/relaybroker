package main

import (
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/op/go-logging"
)

var (
	log    = logging.MustGetLogger("example")
	format = logging.MustStringFormatter(
		`%{color}[%{time:2006-01-02 15:04:05}] [%{level:.4s}] %{color:reset}%{message}`,
	)
)

func main() {
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)

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
	log.Debug(connection.conn, message)
	fmt.Fprintf(connection.conn, "%s\r\n", message)
	atomic.AddInt32(&connection.messages, 1)
	time.AfterFunc(30*time.Second, connection.reduceConnectionMessages)
}
