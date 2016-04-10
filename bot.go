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

func (bot *Bot) reduceJoins() {
	bot.joins--
}

// Bot struct for main config
type Bot struct {
	server     string
	port       string
	oauth      string
	nick       string
	inconn     net.Conn
	mainconn   net.Conn
	connlist   []Connection
	connactive bool
	joins      int
	toJoin     []string
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
		connlist:   make([]Connection, 0),
		connactive: false,
		joins:      0,
	}
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

// Part to leave channels
func (bot *Bot) Part(channel string) {
	for !bot.connactive {
		log.Debugf("chat connection not active yet")
		time.Sleep(time.Second)
	}
	fmt.Fprintf(bot.mainconn, "PART %s\r\n", channel)
	log.Debugf("[chat] parted %s", channel)
}

// ListenToConnection listen
func (bot *Bot) ListenToConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Errorf("Error reading from chat connection: %v", err)
			bot.CreateConnection()
			break // break loop on errors
		}
		if strings.Contains(line, "tmi.twitch.tv 001") {
			bot.connactive = true
		}
		if strings.Contains(line, "PING ") {
			fmt.Fprintf(conn, "PONG tmi.twitch.tv\r\n")
		}
		fmt.Fprintf(bot.inconn, line+"\r\n")
	}
}

// KeepConnectionAlive listen
func (bot *Bot) KeepConnectionAlive(conn net.Conn) {
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Errorf("Error reading from chat connection: %v", err)
			bot.CreateConnection()
			break // break loop on errors
		}
		if strings.Contains(line, "PING ") {
			fmt.Fprintf(conn, "PONG tmi.twitch.tv\r\n")
		}
	}
}

// CreateConnection Add a new connection
func (bot *Bot) CreateConnection() {
	conn, err := net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Noticef("unable to connect to chat IRC server ", err)
		bot.CreateConnection()
		return
	}

	fmt.Fprintf(conn, "PASS %s\r\n", bot.oauth)
	fmt.Fprintf(conn, "USER %s\r\n", bot.nick)
	fmt.Fprintf(conn, "NICK %s\r\n", bot.nick)
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")     // enable ircv3 tags
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n") // enable roomstate and such
	log.Debugf("new connection to chat IRC server %s (%s)\n", bot.server, conn.RemoteAddr())

	connnection := NewConnection(conn)
	bot.connlist = append(bot.connlist, connnection)

	if len(bot.connlist) == 1 {
		bot.mainconn = conn
		go bot.ListenToConnection(conn)
	} else {
		go bot.KeepConnectionAlive(conn)
	}

}

// shuffle simple array shuffle functino
func shuffle(a []Connection) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

// Message to send a message
func (bot *Bot) Message(message string) {
	message = strings.TrimSpace(message)
	for !bot.connactive {
		// wait for connection to become active
	}
	shuffle(bot.connlist)
	for i := 0; i < len(bot.connlist); i++ {
		if bot.connlist[i].messages < 90 {
			bot.connlist[i].Message(message)
			return
		}
	}
	// open new connection when others too full
	log.Debugf("opened new connection, total: %d", len(bot.connlist))
	bot.CreateConnection()
	bot.Message(message)
}

// Whisper to send whispers
func (bot *Bot) Whisper(message string) {
	bot.Message("PRIVMSG #jtv :" + message)
}

// Close clean up bot things
func (bot *Bot) Close() {
	// Close the in connection
	bot.inconn.Close()

	// Close the read connectin
	if bot.mainconn != nil {
		bot.mainconn.Close()
	}

	// Close all listens connections
	for i := 0; i < len(bot.connlist); i++ {
		bot.connlist[i].conn.Close()
	}
}
