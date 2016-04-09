package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/textproto"
	"strings"
	"time"
)

// Bot struct for main config
type Bot struct {
	server   string
	port     string
	oauth    string
	nick     string
	inconn   net.Conn
	connlist []Connection
	channels []string
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:   "irc.chat.twitch.tv",
		port:     "80",
		oauth:    "",
		nick:     "",
		inconn:   nil,
		connlist: make([]Connection, 0),
		channels: make([]string, 0),
	}
}

// Join join a random channel
func (bot *Bot) Join(channel string) {
	for i := range bot.channels {
		if channel == bot.channels[i] {
			log.Println("already joined " + channel)
			return
		}
	}

	shuffleConnections(bot.connlist)

	for i := 0; i < len(bot.connlist); i++ {
		if bot.connlist[i].joins < 45 && bot.connlist[i].connactive {
			bot.connlist[i].Join(channel)
			bot.channels = append(bot.channels, channel)
			log.Printf("[chat] joined %s", channel)
			return
		}
	}
	time.Sleep(time.Second)
	bot.Join(channel)
}

// Part part channels
func (bot *Bot) Part(channel string) {
	// loop connections and find channel
}

// ListenToConnection listen
func (bot *Bot) ListenToConnection(connection *Connection) {
	reader := bufio.NewReader(connection.conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Printf("Error reading from chat connection: %s", err)
			break // break loop on errors
		}
		if strings.Contains(line, "tmi.twitch.tv 001") {
			connection.activateConn()
		}
		if strings.Contains(line, "PING ") {
			fmt.Fprintf(connection.conn, "PONG tmi.twitch.tv\r\n")
		}
		fmt.Fprintf(bot.inconn, line+"\r\n")
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
	log.Printf("new connection to chat IRC server %s (%s)\n", bot.server, conn.RemoteAddr())

	bot.connlist = append(bot.connlist, NewConnection(conn))

	go bot.ListenToConnection(&bot.connlist[len(bot.connlist)-1])
}

// shuffle simple array shuffle functino
func shuffleConnections(a []Connection) {
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
		if bot.connlist[i].messages < 90 && bot.connlist[i].connactive {
			bot.connlist[i].Message(message)
			return
		}
	}
	// open new connection when others too full
	log.Printf("opened new connection, total: %d", len(bot.connlist))
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
