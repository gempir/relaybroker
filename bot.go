package main

import (
	"strings"
	"time"
)

type bot struct {
	ID              string
	pass            string
	nick            string
	read            chan string
	toClient        chan string
	join            chan string
	channels        map[string][]*connection
	readconns       []*connection
	sendconns       []*connection
	whisperconn     *connection
	ticker          *time.Ticker
	clientConnected bool
	client          *Client
}

func newBot(client *Client) *bot {
	return &bot{
		read:      make(chan string, 10),
		join:      make(chan string, 10000000),
		channels:  make(map[string][]*connection),
		readconns: make([]*connection, 0),
		sendconns: make([]*connection, 0),
		ticker:    time.NewTicker(1 * time.Minute),
		client:    client,
	}
}

func (bot *bot) Init() {
	go bot.joinChannels()
	go bot.checkConnections()
	bot.newConn(connReadConn)
	// twitch changed something about whispers or there is some black magic going on,
	// but its only reading whispers once even with more connections
	bot.newConn(connWhisperConn)
}

// close all connections and delete bot
func (bot *bot) close() {
	bot.ticker.Stop()
	close(bot.read)
	close(bot.join)
	for _, conn := range bot.readconns {
		conn.close()
	}
	for _, conn := range bot.sendconns {
		conn.close()
	}
	bot.whisperconn.close()
	for k := range bot.channels {
		delete(bot.channels, k)
	}
	Log.Debug("CLOSED BOT", bot.nick)
}

func (bot *bot) checkConnections() {
	for _ = range bot.ticker.C {
		for _, co := range bot.readconns {
			conn := co
			conn.send("PING")
			go func() {
				time.Sleep(10 * time.Second)
				if !conn.active {
					Log.Debug("send connection died, reconnecting...")
					conn.restore()
					conn.close()
				}
			}()
		}
		for _, co := range bot.sendconns {
			conn := co
			go func() {
				conn.send("PING")
				time.Sleep(10 * time.Second)
				if !conn.active {
					Log.Debug("send connection died, closing...")
					conn.restore()
					conn.close()
				}
			}()
		}

		bot.whisperconn.send("PING")
		time.Sleep(10 * time.Second)
		if !bot.whisperconn.active {
			bot.newConn(connWhisperConn)
		}
	}
}

func (bot *bot) joinChannels() {
	for channel := range bot.join {
		bot.joinChannel(channel)
	}
}

func (bot *bot) joinChannel(channel string) {
	if _, ok := bot.channels[channel]; ok {
		// TODO: check msg ids and join channels more than one time
		Log.Debug("already joined channel", channel)
		return
	}
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
	if _, ok := bot.channels[channel]; !ok {
		bot.channels[channel] = make([]*connection, 0)
	}

	bot.channels[channel] = append(bot.channels[channel], conn)
	Log.Debug("joined channel", channel)

}

func (bot *bot) newConn(t connType) {
	switch t {
	case connReadConn:
		conn := newConnection(t)
		go conn.connect(bot.client, bot.pass, bot.nick)
		bot.readconns = append(bot.readconns, conn)
	case connSendConn:
		conn := newConnection(t)
		go conn.connect(bot.client, bot.pass, bot.nick)
		bot.sendconns = append(bot.sendconns, conn)
	case connWhisperConn:
		if bot.whisperconn != nil {
			bot.whisperconn.close()
		}
		conn := newConnection(t)
		go conn.connect(bot.client, bot.pass, bot.nick)
		bot.whisperconn = conn
	}
}

func (bot *bot) readChat() {
	for msg := range bot.toClient {
		bot.read <- msg
	}
}

// rate limiting is NOT tested properly, but it seems to work Keepo
func (bot *bot) say(msg string) {
	var conn *connection
	var min = 15
	// find connection with the least sent messages
	for _, c := range bot.sendconns {
		if c.msgCount < min {
			conn = c
			min = conn.msgCount
		}
	}
	if conn == nil || min > 10 {
		bot.newConn(connSendConn)
		Log.Debugf("created new conn, total: %d\n", len(bot.sendconns))
		bot.say(msg)
		return
	}
	// add to msg counter before waiting to stop other go routines from sending on this connection
	conn.countMsg()
	for !conn.active {
		time.Sleep(100 * time.Millisecond)
	}
	conn.send("PRIVMSG " + msg)
	Log.Debugf("%p   %d\n", conn, conn.msgCount)
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
