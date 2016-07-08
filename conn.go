package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
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
	if conn.conn != nil {
		conn.conn.Close()
	}
	conn.alive = false
}

func (conn *connection) restore() {
	/*
	   i was sometimes getting slice bounds out of range errors when restoring
	   a lot of connections at a time due to network outages, i hope this fixes it
	*/

	if conn.conntype == connReadConn {

		var i int
		var channels []string
		conn.client.bot.Lock()
		for index, co := range conn.client.bot.readconns {
			if conn == co {
				i = index
				channels = co.joins
				break
			}
		}
		Log.Error("readconn died, lost joins:", channels)
		conn.client.bot.readconns = append(conn.client.bot.readconns[:i], conn.client.bot.readconns[i+1:]...)
		for _, channel := range channels {
			conns := conn.client.bot.channels[channel]
			for i, co := range conns {
				if conn == co {
					conn.client.bot.channels[channel] = append(conns[:i], conns[i+1:]...)
				}
			}
		}

		conn.client.bot.Unlock()
		for _, ch := range channels {
			conn.client.bot.join <- ch
		}
	} else if conn.conntype == connSendConn {
		Log.Error("sendconn died")
		var i int
		conn.client.bot.Lock()
		for index, co := range conn.client.bot.sendconns {
			if conn == co {
				i = index
				break
			}
		}
		conn.client.bot.sendconns = append(conn.client.bot.sendconns[:i], conn.client.bot.sendconns[i+1:]...)
		conn.client.bot.Unlock()
	} else {
		Log.Error("conn died, ", conn.conntype)
	}
}

func (conn *connection) connect(client *Client, pass string, nick string) {
	c, err := tls.Dial("tcp", *addr, nil)
	if err != nil {
		Log.Error("unable to connect to irc server", err)
		time.Sleep(2 * time.Second)
		conn.restore()
	}

	conn.conn = c
	conn.client = client

	conn.login(pass, nick)
	conn.send("CAP REQ :twitch.tv/tags")
	conn.send("CAP REQ :twitch.tv/commands")

	defer func() {
		if r := recover(); r != nil {
			Log.Error(r)
		}
		conn.close()
	}()
	reader := bufio.NewReader(conn.conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			Log.Error("read:", err)
			conn.restore()
			return
		}
		if conn.conntype == connDelete {
			conn.restore()
		}
		if strings.HasPrefix(line, "PING") {
			conn.send(strings.Replace(line, "PING", "PONG", 1))
		} else if strings.HasPrefix(line, "PONG") {
			Log.Debug("PONG")
		} else {
			client.toClient <- line
		}
		conn.active = true
		stats.totalMsgsReceived++
	}
}

func (conn *connection) send(msg string) {
	conn.Lock()
	fmt.Fprint(conn.conn, msg+"\r\n")
	stats.totalMsgsSent++
	conn.Unlock()
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
