package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type connType uint32

const (
	connWhisperConn = iota
	connReadConn
	connSendConn
	connDelete
)

type connection struct {
	sync.Mutex
	conn     net.Conn
	active   bool
	anon     bool
	joins    []string
	msgCount int
	lastUse  time.Time
	alive    bool
	conntype connType
	bot      *bot
	name     string
}

func newConnection(t connType) *connection {
	return &connection{
		joins:    make([]string, 0),
		conntype: t,
		lastUse:  time.Now(),
		name:     randomHash(),
	}
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
	if conn.conn != nil {
		conn.conn.Close()
	}
	for _, channel := range conn.joins {
		conn.part(channel)
	}
	conn.alive = false
}

func (conn *connection) part(channel string) {
	channel = strings.ToLower(channel)
	for i, ch := range conn.joins {
		if ch == channel {
			conn.joins = append(conn.joins[:i], conn.joins[i+1:]...)
		}
	}
}

func (conn *connection) restore() {
	defer func() {
		if r := recover(); r != nil {
			Log.Error("cannot restore connection")
		}
	}()
	if conn.conntype == connReadConn {
		var i int
		var channels []string
		for index, co := range conn.bot.readconns {
			if conn == co {
				i = index
				channels = co.joins
				break
			}
		}
		Log.Error("readconn died, lost joins:", channels)
		conn.bot.Lock()
		conn.bot.readconns = append(conn.bot.readconns[:i], conn.bot.readconns[i+1:]...)
		conn.bot.Unlock()
		for _, channel := range channels {
			conns := conn.bot.channels[channel]
			for i, co := range conns {
				if conn == co {
					conn.bot.Lock()
					conn.bot.channels[channel] = append(conns[:i], conns[i+1:]...)
					conn.bot.Unlock()
					conn.part(channel)
				}
			}
			conn.bot.join <- channel

		}

	} else if conn.conntype == connSendConn {
		Log.Error("sendconn died")
		var i int
		for index, co := range conn.bot.sendconns {
			if conn == co {
				i = index
				break
			}
		}
		conn.bot.Lock()
		conn.bot.sendconns = append(conn.bot.sendconns[:i], conn.bot.sendconns[i+1:]...)
		conn.bot.Unlock()
	} else if conn.conntype == connWhisperConn {
		Log.Error("whisperconn died, reconnecting")
		conn.close()
		conn.bot.newConn(connWhisperConn)
	}
	conn.conntype = connDelete
}

func (conn *connection) connect(client *Client, pass string, nick string) {
	dialer := &net.Dialer{
		KeepAlive: time.Second * 10,
	}

	conn.bot = client.bot
	c, err := tls.DialWithDialer(dialer, "tcp", *addr, &tls.Config{})
	if err != nil {
		Log.Error("unable to connect to irc server", err)
		time.Sleep(2 * time.Second)
		conn.restore()
		return
	}
	conn.conn = c

	conn.send("CAP REQ :twitch.tv/tags twitch.tv/commands")
	conn.login(pass, nick)

	defer func() {
		if r := recover(); r != nil {
			Log.Error("error connecting")
		}
		conn.restore()
	}()
	tp := textproto.NewReader(bufio.NewReader(conn.conn))

	for {
		line, err := tp.ReadLine()
		if err != nil {
			Log.Errorf("[READERROR:%s] %s", conn.name, err.Error())
			conn.restore()
			return
		}
		Log.Debugf("[TWITCH:%s] %s", conn.name, line)
		if conn.conntype == connDelete {
			conn.restore()
		}
		if strings.HasPrefix(line, "PING") {
			conn.send(strings.Replace(line, "PING", "PONG", 1))
		} else if strings.HasPrefix(line, "PONG") {
			Log.Debug("PONG")
		} else {
			if isWhisper(line) && conn.conntype != connWhisperConn {
				// throw away message
			} else {
				client.toClient <- line
			}
		}
		conn.active = true
	}
}

func isWhisper(line string) bool {
	if !strings.Contains(line, ".tmi.twitch.tv WHISPER ") {
		return false
	}
	spl := strings.SplitN(line, " :", 3)
	if strings.Contains(spl[1], ".tmi.twitch.tv WHISPER ") {
		return true
	}
	return false
}

func (conn *connection) send(msg string) error {
	if conn.conn == nil {
		Log.Error("conn is nil", conn, conn.conn)
		return errors.New("connection is nil")
	}
	_, err := fmt.Fprint(conn.conn, msg+"\r\n")
	if err != nil {
		Log.Error("error sending message")
		return err
	}
	Log.Debugf("[OUTGOING:%s] %s", conn.name, msg)
	return nil
}

func (conn *connection) reduceMsgCount() {
	conn.msgCount--
}

func (conn *connection) countMsg() {
	conn.msgCount++
	time.AfterFunc(30*time.Second, conn.reduceMsgCount)
}

func randomHash() string {
	n := 5
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", b)
}
