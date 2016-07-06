package main

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type connType uint32

const (
	connWhisperConn = iota
	connReadConn
	connSendConn
)

type connection struct {
	sync.Mutex
	conn     *websocket.Conn
	active   bool
	anon     bool
	joins    []string
	msgCount int32
	alive    bool
	conntype connType
	client   *Client
}

func newConnection(t connType) *connection {
	c := &connection{
		joins:    make([]string, 0),
		conntype: t,
	}

	return c
}

func (conn *connection) login(pass string, nick string) {
	conn.anon = pass == ""
	if !conn.anon {
		conn.send("PASS " + pass)
		conn.send("NICK " + nick)
		return
	}
	conn.send("NICK justinfan123")
}

func (conn *connection) close() {
	conn.conn.Close()
	conn.alive = false
}

func (conn *connection) restore() {
	if conn.conntype == connReadConn {
		var i int
		var channels []string
		for index, co := range conn.client.bot.sendconns {
			if conn == co {
				i = index
				channels = co.joins
				break
			}
		}
		conn.client.bot.readconns = append(conn.client.bot.readconns[:i], conn.client.bot.readconns[i+1:]...)
		for _, ch := range channels {
			conn.client.bot.join <- ch
		}
	} else {
		var i int
		for index, co := range conn.client.bot.sendconns {
			if conn == co {
				i = index
				break
			}
		}
		conn.client.bot.sendconns = append(conn.client.bot.sendconns[:i], conn.client.bot.sendconns[i+1:]...)
	}
}

func (conn *connection) connect(client *Client, pass string, nick string) {
	u := url.URL{Scheme: "wss", Host: *addr, Path: "/"}
	Log.Info("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		Log.Error("dial:", err)
		return
	}
	conn.conn = c
	conn.client = client

	conn.login(pass, nick)
	conn.send("CAP REQ :twitch.tv/tags")
	conn.send("CAP REQ :twitch.tv/commands")

	//done := make(chan struct{})

	defer conn.close()
	// close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			Log.Debug("read:", err)
			conn.restore()
			return
		}
		// idk if i have to do this , seems a little wierd to me but it didnt work before
		m := string(message)
		lines := strings.Split(m, "\r\n")
		for _, l := range lines {
			if l != "" {
				if strings.HasPrefix(l, "PING") {
					conn.send(strings.Replace(l, "PING", "PONG", 1))
				} else if strings.HasPrefix(l, "PONG") {
					Log.Debug("PONG")
				} else {
					client.toClient <- l
				}
				conn.active = true
			}
		}
	}
}

func (conn *connection) send(msg string) {
	conn.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (conn *connection) reduceMsgCount() {
	conn.Lock()
	conn.msgCount--
	conn.Unlock()
}

func (conn *connection) countMsg() {
	conn.Lock()
	conn.msgCount++
	conn.Unlock()
	time.AfterFunc(30*time.Second, conn.reduceMsgCount)
}
