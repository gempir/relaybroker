package main

import (
	"strings"
	"time"
)

type bot struct {
	ID        string
	pass      string
	nick      string
	read      chan string
	toClient  chan string
	join      chan string
	channels  map[string][]*connection
	readconns []*connection
	sendconns []*connection
}

func newBot(toClient chan string) *bot {
	return &bot{
		read:      make(chan string, 10),
		toClient:  toClient,
		join:      make(chan string, 10000000),
		readconns: make([]*connection, 0),
		sendconns: make([]*connection, 0),
	}
}

func (bot *bot) Init() {
	go bot.joinChannels()
	bot.newConn(connReadConn)

}

func (bot *bot) joinChannels() {
	for channel := range bot.join {
		bot.joinChannel(channel)
	}
}

func (bot *bot) joinChannel(channel string) {
	var conn *connection
	for _, c := range bot.readconns {
		if len(c.joins) < 50 {
			conn = c
			break
		}
	}
	if conn == nil {
		bot.newConn(connReadConn)
		bot.joinChannel(channel)
		return
	}
	for !conn.active {
		time.Sleep(100 * time.Millisecond)
	}
	conn.send("JOIN " + channel)
	Log.Debug("joined channel", channel)

}

func (bot *bot) newConn(t connType) {
	switch t {
	case connReadConn:
		conn := newConnection(t)
		go conn.connect(bot.toClient, bot.pass, bot.nick)
		//conn.login(bot.pass, bot.nick)
		bot.readconns = append(bot.readconns, conn)
	case connSendConn:
		conn := newConnection(t)
		go conn.connect(bot.toClient, bot.pass, bot.nick)
		//conn.login(bot.pass, bot.nick)
		bot.sendconns = append(bot.sendconns, conn)
	}
}

func (bot *bot) readChat() {
	for msg := range bot.toClient {
		bot.read <- msg
	}
}

func (bot *bot) say(msg string) {
	var conn *connection
	for _, c := range bot.sendconns {
		if c.msgCount < 15 {
			conn = c
			break
		}
	}
	if conn == nil {
		bot.newConn(connSendConn)
		bot.say(msg)
		return
	}
	for !conn.active {
		time.Sleep(100 * time.Millisecond)
		Log.Debug("conn not active yet")
	}
	conn.send("PRIVMSG " + msg)
	conn.countMsg()
	Log.Debug("sent:", msg)
}

func (bot *bot) handleMessage(spl []string) {
	msg := spl[1]
	switch spl[0] {
	case "JOIN":
		bot.join <- strings.ToLower(msg)
	case "PART":
		//
	case "PRIVMSG":
		bot.say(msg)
	default:
		Log.Error("unhandled message", spl[0], msg)
	}
}
