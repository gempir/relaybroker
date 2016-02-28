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
	inconn      net.Conn
	connlist    []net.Conn
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:      "irc.twitch.tv",
		groupserver: "group.tmi.twitch.tv",
		port:        "6667",
		inconn:      nil,
		connlist:    make([]net.Conn, 0),
	}
}

// Add a new connection
func (bot *Bot) CreateConnection() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Fatal("unable to connect to IRC server ", err)
		return nil, err
	}
	fmt.Fprintf(conn, "PASS %s\r\n", BotPass)
	fmt.Fprintf(conn, "USER %s\r\n", BotNick)
	fmt.Fprintf(conn, "NICK %s\r\n", BotNick)
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")     // enable ircv3 tags
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n") // enable roomstate and such
	log.Printf("Connected to IRC server %s (%s)\n", bot.server, conn.RemoteAddr())
	return conn, nil
}

// Connect basic connection
/*
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
*/

func main() {
	ircbot := NewBot()
	go TCPServer(ircbot)
	/*
		conn, _ := ircbot.Connect()
		defer conn.Close()
	*/
	conn, err := ircbot.CreateConnection()
	fmt.Printf("conn:%s\n", conn)
	fmt.Printf("err:%s\n", err)
	ircbot.connlist = append(ircbot.connlist, conn)
	fmt.Printf("connlist:%s\n", ircbot.connlist)
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
			fmt.Fprintf(conn, "PONG %s\r\n", pongdata[1])
		}
		ircbot.Handle(line)
	}

}

// HandleJoin will slowly join all channels given
// 45 per 11 seconds to deal with twitch ratelimits
func (bot *Bot) HandleJoin(channels []string) {
	for _, channel := range channels {
		for _, conn := range bot.connlist {
			fmt.Println("Joining " + channel)
			fmt.Println(conn)
			fmt.Fprintf(conn, "JOIN %s\r\n", channel)
		}
	}
}

// Message to send a message
func (bot *Bot) Message(channel string, message string) {
	if message == "" {
		return
	}
	fmt.Printf("Bot: " + message + "\n")

	/* Find a suitable connection to use */
	for _, conn := range bot.connlist {
		fmt.Fprintf(conn, "PRIVMSG %s :%s\r\n", channel, message)
	}
}

// Handle handles messages from irc
func (bot *Bot) Handle(line string) {
	if strings.Contains(line, ".tmi.twitch.tv PRIVMSG ") {
		fmt.Println("Sending data to inconn!")
		bot.inconn.Write([]byte(line + "\r\n"))
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
