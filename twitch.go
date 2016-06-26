package main

import (
	"flag"
	"net"
	"net/url"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "irc-ws.chat.twitch.tv:80", "https websockets address")

// Connection simple type for connection to twitch
type Connection struct {
	msgCount      int // number of messages in 30 seconds
	conn          net.Conn
	channels      []string
	authenticated bool
	alive         bool
}

func newConnection(nick string, pass string) Connection {
	return Connection{
		msgCount:      0,
		conn:          nil,
		channels:      nil,
		authenticated: false,
		alive:         false,
	}
}

func connect(ct Connection) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	Log.Info("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		Log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	defer c.Close()
	defer close(done)

	c.WriteMessage(websocket.TextMessage, []byte("PING"))
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			Log.Debug("read:", err)
			return
		}
		Log.Debug(string(message))
	}
}
