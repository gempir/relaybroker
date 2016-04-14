package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"net/textproto"
	"strings"
	"time"
)

// Bot struct for main config
type Bot struct {
	server     string
	port       string
	oauth      string
	nick       string
	inconn     net.Conn
	mainconn   net.Conn
	connlist   []*Connection
	connactive bool
	login      bool
	joins      int
	anon       bool
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:     "irc.chat.twitch.tv",
		port:       "80",
		oauth:      "",
		nick:       "",
		inconn:     nil,
		mainconn:   nil,
		connlist:   make([]*Connection, 0),
		connactive: false,
		login:      false,
		joins:      0,
		anon:       true,
	}
}
func (bot *Bot) reduceJoins() {
	bot.joins--
}

// Join joins a channel
func (bot *Bot) Join(channel string) {
	for !bot.connactive {
		log.Debugf("chat connection not active yet [%s]\n", bot.nick)
		time.Sleep(time.Second)
	}

	if bot.joins < 45 {
		fmt.Fprintf(bot.mainconn, "JOIN %s\r\n", channel)
		log.Debugf("[chat] joined %s", channel)
		bot.joins++
		time.AfterFunc(10*time.Second, bot.reduceJoins)
	} else {
		log.Debugf("[chat] in queue to join %s", channel)
		time.Sleep(time.Second)
		bot.Join(channel)
	}
}

// Whisper to send whispers
func (bot *Bot) Whisper(message string) {
	bot.Message("PRIVMSG #jtv :" + message)
}

// Part part channels
func (bot *Bot) Part(channel string) {
	// loop connections and find channel
}

// CreateConnection Add a new connection
func (bot *Bot) CreateConnection() {
	conn, err := net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Errorf("unable to connect to chat IRC server %v", err)
		bot.CreateConnection()
		return
	}
	connnection := NewConnection(conn)

	if bot.oauth != "" {
		fmt.Fprintf(connnection.conn, "PASS %s\r\n", bot.oauth)
		connnection.anon = false
	}
	fmt.Fprintf(connnection.conn , "USER %s\r\n", bot.nick)
	fmt.Fprintf(connnection.conn, "NICK %s\r\n", bot.nick)
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n")
	log.Debugf("new connection to chat IRC server %s (%s)\n", bot.server, conn.RemoteAddr())

	if len(bot.connlist) == 0 {
		bot.mainconn = connnection.conn
		go bot.ListenToConnection(&connnection)
	} else {
		go bot.KeepConnectionAlive(&connnection)
	}
	bot.connlist = append(bot.connlist, &connnection)
}


// ListenToConnection listen
func (bot *Bot) ListenToConnection(connection *Connection) {
	reader := bufio.NewReader(connection.conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Errorf("Error reading from chat connection: %s", err)
			break // break loop on errors
		}
		if strings.Contains(line, "tmi.twitch.tv 001") {
			connection.active = true
			bot.connactive = true
		}
		if strings.Contains(line, "PING ") {
			fmt.Fprintf(connection.conn, "PONG tmi.twitch.tv\r\n")
		}
		fmt.Fprint(bot.inconn, line+"\r\n")
	}
}

// KeepConnectionAlive listen
func (bot *Bot) KeepConnectionAlive(connection *Connection) {
	reader := bufio.NewReader(connection.conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Errorf("Error reading from chat connection: %v", err)
			bot.CreateConnection()
			break // break loop on errors
		}
		if strings.Contains(line, "tmi.twitch.tv 001") {
			connection.active = true
		}
		if strings.Contains(line, "PING ") {
			fmt.Fprintf(connection.conn, "PONG tmi.twitch.tv\r\n")
		}
	}
}

// shuffle simple array shuffle functino
func shuffleConnections(a []*Connection) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

// Message to send a message
func (bot *Bot) Message(message string) {
	message = strings.TrimSpace(message)
	shuffleConnections(bot.connlist)
	for i := 0; i < len(bot.connlist); i++ {
		if bot.connlist[i].messages < 90 {
			err := bot.connlist[i].Message(message)
			if err != nil {
				log.Error(err)
				if err.Error() == "connection is anonymous" {
					return
				}
				time.Sleep(time.Second)
				bot.Message(message)
			}
			return
		}
	}
	// open new connection when others too full
	log.Debugf("opened new connection, total: %d", len(bot.connlist))
	bot.CreateConnection()
	bot.Message(message)
}

// Close clean up bot things
func (bot *Bot) Close() {
	// Close the in connection
	bot.inconn.Close()

	// Close all listens connections
	for i := 0; i < len(bot.connlist); i++ {
		bot.connlist[i].conn.Close()
	}
}
