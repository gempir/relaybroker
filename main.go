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

var (
	toJoin []string
)

// Bot struct for main config
type Bot struct {
	server      string
	groupserver string
	port        string
	conn        net.Conn
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:      "irc.twitch.tv",
		groupserver: "group.tmi.twitch.tv",
		port:        "6667",
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
	fmt.Fprintf(bot.conn, "CAP REQ :twitch.tv/tags\r\n")     // enable ircv3 tags
	fmt.Fprintf(bot.conn, "CAP REQ :twitch.tv/commands\r\n") // enable roomstate and such
	log.Printf("Connected to IRC server %s (%s)\n", bot.server, bot.conn.RemoteAddr())
	return bot.conn, nil
}

func main() {
	ircbot := NewBot()
	go TCPServer(ircbot)
	go ircbot.HandleJoin()
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

// AddToJoin will just add channels to the channels still to Join
func (bot *Bot) AddToJoin(channels []string) {
	for _, element := range channels {
		toJoin = append(toJoin, element)
	}
}

// HandleJoin will slowly join all channels given
// 45 per 11 seconds to deal with twitch ratelimits
func (bot *Bot) HandleJoin() {
	for {
		if len(toJoin) == 0 {
			continue
		}

		if len(toJoin) > 45 {

		} else {
			for _, channel := range toJoin {
				fmt.Fprintf(bot.conn, "JOIN %s\r\n", channel)
				time.Sleep(010 * time.Millisecond)
			}
			toJoin = toJoin[:0]
		}
		time.Sleep(11000 * time.Millisecond)
	}
}

// Message to send a message
func (bot *Bot) Message(channel string, message string) {
	if message == "" {
		return
	}
	fmt.Printf("Bot: " + message + "\n")
	fmt.Fprintf(bot.conn, "PRIVMSG "+channel+" :"+message+"\r\n")
}

// Handle handles messages from irc
func (bot *Bot) Handle(line string) {
	if strings.Contains(line, ".tmi.twitch.tv PRIVMSG ") {
		messageTMISplit := strings.Split(line, ".tmi.twitch.tv PRIVMSG ")
		messageChannelRaw := strings.Split(messageTMISplit[1], " :")
		channel := messageChannelRaw[0]
		go bot.ProcessMessage(channel, line)
	} else if strings.Contains(line, ":tmi.twitch.tv ROOMSTATE") {
		messageTMISplit := strings.Split(line, ":tmi.twitch.tv ROOMSTATE ")
		channel := messageTMISplit[1]
		go bot.ProcessMessage(channel, line)
	}
}

// ProcessMessage push message to local irc chat
func (bot *Bot) ProcessMessage(channel string, message string) {
	fmt.Println(channel + " ::: " + message)
}
