package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
)

// Bot struct for main config
type Bot struct {
	server      string
	groupserver string
	port        string
	channel     string
	conn        net.Conn
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:      "irc.twitch.tv",
		groupserver: "group.tmi.twitch.tv",
		port:        "6667",
		channel:     "#gempir",
		conn:        nil,
	}
}

// Connect basic connection
func (bot *Bot) Connect() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Fatal("unable to connect to IRC server ", err)
	}
	bot.conn = conn
	fmt.Fprintf(bot.conn, "PASS %s\r\n", BotPass)
	fmt.Fprintf(bot.conn, "USER %s\r\n", BotNick)
	fmt.Fprintf(bot.conn, "NICK %s\r\n", BotNick)
	log.Printf("Connected to IRC server %s (%s)\n", bot.server, bot.conn.RemoteAddr())
	return bot.conn, nil
}

func main() {
	ircbot := NewBot()
	go TCPServer(ircbot)
	conn, _ := ircbot.Connect()
	defer conn.Close()

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			break // break loop on errors
		}
		if strings.Contains(line, "PING") {
			pongdata := strings.Split(line, "PING ")
			fmt.Fprintf(ircbot.conn, "PONG %s\r\n", pongdata[1])
		}
		ircbot.Handle(line)
	}

}

// JoinChannel joins  an irc channel
func (bot *Bot) JoinChannel(channel string) {
	fmt.Fprintf(bot.conn, "JOIN %s\r\n", channel)
}

// Message to send a message
func (bot *Bot) Message(message string) {
	if message == "" {
		return
	}
	fmt.Printf("Bot: " + message + "\n")
	fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :"+message+"\r\n")
}

// Handle handles messages from irc
func (bot *Bot) Handle(line string) {

	if strings.Contains(line, ".tmi.twitch.tv PRIVMSG "+bot.channel) {
		userdata := strings.Split(line, ".tmi.twitch.tv PRIVMSG "+bot.channel)
		username := strings.Split(userdata[0], "@")
		usermessage := strings.Replace(userdata[1], " :", "", 1)
		fmt.Printf(username[1] + ": " + usermessage + "\n")
		bot.ProcessMessage(username[1], usermessage)
	}
}

// ProcessMessage push message to local irc chat
func (bot *Bot) ProcessMessage(username string, message string) {

}
