package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
	"time"
)

// Connection stores messages sent in the last 30 seconds and the connection itself
type Connection struct {
	conn     net.Conn
	messages int
}

// NewConnection initialize a Connection struct
func NewConnection(conn net.Conn) Connection {
	return Connection{
		conn:     conn,
		messages: 0,
	}

}

func (connection *Connection) reduceConnectionMessages() {
	connection.messages--
}

// Message called everytime you send a message
func (connection *Connection) Message(channel string, message string) {
	fmt.Fprintf(connection.conn, "PRIVMSG %s :%s\r\n", channel, message)
	connection.messages++
	time.AfterFunc(30*time.Second, connection.reduceConnectionMessages)
}

// Bot struct for main config
type Bot struct {
	server      string
	groupserver string
	port        string
	oauth       string
	nick        string
	inconn      net.Conn
	mainconn    net.Conn
	connlist    []Connection
	groupconn   net.Conn
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:      "irc.twitch.tv",
		groupserver: "group.tmi.twitch.tv",
		port:        "6667",
		oauth:       "",
		nick:        "",
		inconn:      nil,
		mainconn:    nil,
		connlist:    make([]Connection, 0),
		groupconn:   nil,
	}
}

// ListenToConnection listen
func (bot *Bot) ListenToConnection(conn net.Conn) {
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
		bot.Handle(line)
	}
}

// CreateConnection Add a new connection
func (bot *Bot) CreateConnection() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Fatal("unable to connect to IRC server ", err)
		return nil, err
	}
	fmt.Fprintf(conn, "PASS %s\r\n", bot.oauth)
	fmt.Fprintf(conn, "USER %s\r\n", bot.nick)
	fmt.Fprintf(conn, "NICK %s\r\n", bot.nick)
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")     // enable ircv3 tags
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n") // enable roomstate and such
	log.Printf("new Connected to IRC server %s (%s)\n", bot.server, conn.RemoteAddr())

	connnection := NewConnection(conn)
	bot.connlist = append(bot.connlist, connnection)

	if len(bot.connlist) == 1 {
		bot.mainconn = conn
	}

	go bot.ListenToConnection(conn)

	return conn, nil
}

func main() {
	ret := TCPServer()
	log.Printf("got ret code %d\n", ret)
	os.Exit(ret)
}

// HandleJoin will slowly join all channels given
// 45 per 11 seconds to deal with twitch ratelimits
func (bot *Bot) HandleJoin(channels []string) {
	if bot.mainconn == nil {
		log.Printf("No main conn set, can't join channels yet.\n")
		return
	}
	for _, channel := range channels {
		log.Printf("Joining %s\n", channel)
		fmt.Fprintf(bot.mainconn, "JOIN %s\r\n", channel)
	}
}

// Message to send a message
func (bot *Bot) Message(channel string, message string) {
	if message == "" {
		return
	}

	for i := 0; i < len(bot.connlist); i++ {
		if bot.connlist[i].messages < 90 {
			bot.connlist[i].Message(channel, message)
			return
		}
	}
	newConn, _ := bot.CreateConnection()
	fmt.Fprintf(newConn, "PRIVMSG %s :%s\r\n", channel, message)
	log.Println(newConn)
}

// Handle handles messages from irc
func (bot *Bot) Handle(line string) {
	if strings.Contains(line, ".tmi.twitch.tv PRIVMSG ") {
		bot.inconn.Write([]byte(line + "\r\n"))
		messageTMISplit := strings.Split(line, ".tmi.twitch.tv PRIVMSG ")
		messageChannelRaw := strings.Split(messageTMISplit[1], " :")
		channel := messageChannelRaw[0]
		bot.ProcessMessage(channel, line)
	} else if strings.Contains(line, ":tmi.twitch.tv ROOMSTATE") {
		messageTMISplit := strings.Split(line, ":tmi.twitch.tv ROOMSTATE ")
		channel := messageTMISplit[1]
		bot.ProcessMessage(channel, line)
	}
}

// ProcessMessage push message to local irc chat
func (bot *Bot) ProcessMessage(channel string, message string) {
	fmt.Println(channel + " ::: " + message)
}

// WriteToAllConns writes message to all connections for now
func (bot *Bot) WriteToAllConns(message string) {
	for _, connection := range bot.connlist {
		fmt.Fprintf(connection.conn, message+"\r\n")
	}
}
