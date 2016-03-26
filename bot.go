package main

import (
	"bufio"
	"fmt"
	"log"
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
	server          string
	groupserver     string
	port            string
	groupport       string
	oauth           string
	nick            string
	inconn          net.Conn
	mainconn        net.Conn
	connlist        []Connection
	connactive      bool
	groupconn       net.Conn
	groupconnactive bool
	joins           int
	toJoin          []string
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:          "irc.chat.twitch.tv",
		groupserver:     "group.tmi.twitch.tv",
		port:            "80",
		groupport:       "6667",
		oauth:           "",
		nick:            "",
		inconn:          nil,
		mainconn:        nil,
		connlist:        make([]Connection, 0),
		connactive:      false,
		groupconn:       nil,
		groupconnactive: false,
		joins:           0,
	}
}

func (bot *Bot) join(channel string) {
	for !bot.connactive {
		log.Printf("chat connection not active yet")
		time.Sleep(time.Second)
	}

	if bot.joins < 45 {
		fmt.Fprintf(bot.mainconn, "JOIN %s\r\n", channel)
		log.Printf("[chat] joined %s", channel)
		bot.joins++
		time.AfterFunc(10*time.Second, bot.reduceJoins)
	} else {
		log.Printf("[chat] in queue to join %s", channel)
		time.Sleep(time.Second)
		bot.join(channel)
	}
}

// ListenToConnection listen
func (bot *Bot) ListenToConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Printf("Error reading from chat connection: %s", err)
			bot.CreateConnection()
			break // break loop on errors
		}
		if strings.Contains(line, "tmi.twitch.tv 001") {
			bot.connactive = true
		}
		if strings.Contains(line, "PING ") {
			fmt.Fprintf(conn, "PONG tmi.twitch.tv\r\n")
			log.Printf("PONG tmi.twitch.tv\r\n")
		}
		bot.inconn.Write([]byte(line + "\r\n"))
	}
}

// ListenToGroupConnection validate connection is running and listen to it
func (bot *Bot) ListenToGroupConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Printf("Error reading from group connection: %s", err)
			bot.CreateGroupConnection()
			break
		}
		if strings.Contains(line, "tmi.twitch.tv 001") {
			bot.groupconnactive = true
		}
		if strings.Contains(line, "PING ") {
			fmt.Fprintf(conn, "PONG tmi.twitch.tv\r\n")
			log.Printf("PONG tmi.twitch.tv\r\n")
		}
		bot.inconn.Write([]byte(line + "\r\n"))
	}
}

// CreateConnection Add a new connection
func (bot *Bot) CreateConnection() {
	conn, err := net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Println("unable to connect to chat IRC server ", err)
		bot.CreateConnection()
		return
	}
	fmt.Fprintf(conn, "PASS %s\r\n", bot.oauth)
	fmt.Fprintf(conn, "USER %s\r\n", bot.nick)
	fmt.Fprintf(conn, "NICK %s\r\n", bot.nick)
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")     // enable ircv3 tags
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n") // enable roomstate and such
	log.Printf("new connection to chat IRC server %s (%s)\n", bot.server, conn.RemoteAddr())

	connnection := NewConnection(conn)
	bot.connlist = append(bot.connlist, connnection)

	if len(bot.connlist) == 1 {
		bot.mainconn = conn
	}

	go bot.ListenToConnection(conn)

}

// CreateGroupConnection creates connection to recevie and send whispers
func (bot *Bot) CreateGroupConnection() {
	conn, err := net.Dial("tcp", bot.groupserver+":"+bot.groupport)
	if err != nil {
		log.Println("unable to connect to group IRC server ", err)
		bot.CreateGroupConnection()
		return
	}
	fmt.Fprintf(conn, "PASS %s\r\n", bot.oauth)
	fmt.Fprintf(conn, "USER %s\r\n", bot.nick)
	fmt.Fprintf(conn, "NICK %s\r\n", bot.nick)
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")     // enable ircv3 tags
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n") // enable roomstate and such
	log.Printf("new connection to group IRC server %s (%s)\n", bot.groupserver, conn.RemoteAddr())

	bot.groupconn = conn

	go bot.ListenToGroupConnection(conn)
}

// Message to send a message
func (bot *Bot) Message(message string) {
	for !bot.connactive {
		// wait for connection to become active
	}

	for i := 0; i < len(bot.connlist); i++ {
		if bot.connlist[i].messages < 90 {
			bot.connlist[i].Message(message)
			return
		}
	}
	// open new connection when others too full
	log.Printf("opened new connection, total: %d", len(bot.connlist))
	bot.CreateConnection()
	bot.Message(message)
}

// Whisper to send whispers
func (bot *Bot) Whisper(message string) {
	for !bot.groupconnactive {
		// wait for connection to become active
	}
	fmt.Fprintf(bot.groupconn, "PRIVMSG #jtv :"+message+"\r\n")
	log.Printf(message)
}
