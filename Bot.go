package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type connType uint32

const (
	connWhisperconn connType = iota + 1
	connSendconn
	connReadconn
)

type msgType uint32

const (
	msgPrivmsg msgType = iota + 1
	msgWhisper
	msgOther
)

// Bot struct for main config
type Bot struct {
	sync.Mutex
	server      string
	port        string
	oauth       string
	nick        string
	inconn      net.Conn
	whisperconn *Connection
	readconn    []*Connection
	connlist    []*Connection
	connactive  bool
	login       bool
	anon        bool
	join        chan string
}

// NewBot main config
func NewBot() *Bot {
	return &Bot{
		server:     "irc.chat.twitch.tv",
		port:       "80",
		oauth:      "",
		nick:       "",
		inconn:     nil,
		readconn:   make([]*Connection, 0),
		connlist:   make([]*Connection, 0),
		connactive: false,
		login:      false,
		anon:       true,
		join:       make(chan string, 100000),
	}
}

func getmsgType(line string) msgType {
	if !strings.Contains(line, ".tmi.twitch.tv ") {
		return msgOther
	}
	spl := strings.SplitN(line, ".tmi.twitch.tv ", 2)
	t := strings.Split(spl[1], " ")[0]
	if t == "WHISPER" {
		return msgWhisper
	} else if t == "PRIVMSG" {
		return msgPrivmsg
	} else {
		return msgOther
	}
}

func (bot *Bot) getReadconn() *Connection {
	var conn *Connection
	for _, c := range bot.readconn {
		if len(c.joins) < 50 {
			conn = c
			break
		}
	}
	if conn == nil {
		bot.CreateConnection(connReadconn)
		return bot.getReadconn()
	}
	return conn
}

// Join joins a channel
func (bot *Bot) Join() {
	var isOpen = true
	for isOpen {
		channel, isOpen := <-bot.join
		if !isOpen {
			bot.Close()
			return
		}
		alreadyJoined := false
		func() {
			for _, co := range bot.readconn {
				for _, ch := range co.joins {
					if channel == ch {
						alreadyJoined = true
						return
					}
				}
			}
		}()

		if alreadyJoined {
			log.Debug("already joined channel ", channel)
		} else {
			for !bot.connactive {
				log.Debugf("chat connection not active yet [%s]\n", bot.nick)
				time.Sleep(time.Second)
			}
			conn := bot.getReadconn()
			fmt.Fprintf(conn.conn, "JOIN %s\r\n", channel)
			log.Debugf("[chat] joined %s", channel)
			conn.joins = append(conn.joins, channel)
			time.Sleep(300 * time.Millisecond)
		}
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
func (bot *Bot) CreateConnection(conntype connType) {
	conn, err := net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Errorf("unable to connect to chat IRC server %v", err)
		bot.CreateConnection(conntype)
		return
	}
	connection := NewConnection(conn)

	if bot.oauth != "" {
		fmt.Fprintf(connection.conn, "PASS %s\r\n", bot.oauth)
		connection.anon = false
	}
	fmt.Fprintf(connection.conn, "USER %s\r\n", bot.nick)
	fmt.Fprintf(connection.conn, "NICK %s\r\n", bot.nick)
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n")
	log.Debugf("new connection to chat IRC server %s (%s)\n", bot.server, conn.RemoteAddr())

	if conntype == connReadconn {
		bot.readconn = append(bot.readconn, &connection)
		go bot.ListenToConnection(&connection)

	} else if conntype == connWhisperconn {
		bot.whisperconn = &connection
		go bot.ListenForWhispers(&connection)

	} else {
		go bot.KeepConnectionAlive(&connection)
		bot.connlist = append(bot.connlist, &connection)
	}

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
		if strings.HasPrefix(line, "PING ") {
			fmt.Fprintf(connection.conn, "PONG tmi.twitch.tv\r\n")
		}
		if strings.HasPrefix(line, "PONG ") {
			connection.alive = true
		}
		if getmsgType(line) != msgWhisper && !strings.HasPrefix(line, "PONG ") {
			fmt.Fprint(bot.inconn, line+"\r\n")
		}
	}
}

//ListenForWhispers only reads whispers
func (bot *Bot) ListenForWhispers(connection *Connection) {
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
		if strings.HasPrefix(line, "PING ") {
			fmt.Fprintf(connection.conn, "PONG tmi.twitch.tv\r\n")
		}
		if strings.HasPrefix(line, "PONG ") {
			connection.alive = true
		}
		if getmsgType(line) == msgWhisper {
			fmt.Fprint(bot.inconn, line+"\r\n")
		}
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
			bot.CreateConnection(connSendconn)
			break // break loop on errors
		}
		if strings.Contains(line, "tmi.twitch.tv 001") {
			connection.active = true
		}
		if strings.HasPrefix(line, "PING ") {
			fmt.Fprintf(connection.conn, "PONG tmi.twitch.tv\r\n")
		}
		if strings.HasPrefix(line, "PONG ") {
			connection.alive = true
		}
	}
}

func (bot *Bot) rejoinChannels(channels []string) {
	for _, channel := range channels {
		bot.join <- channel
	}
}

func getIndex(conn *Connection, s []*Connection) int {
	for i, c := range s {
		if c == conn {
			return i
		}
	}
	return 0
}

func (bot *Bot) checkConnections() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		<-ticker.C
		for _, conn := range bot.readconn {
			go bot.checkConnection(conn)
		}
		for _, conn := range bot.connlist {
			go bot.checkConnection(conn)
		}
		go bot.checkConnection(bot.whisperconn)
	}
}

func (bot *Bot) checkConnection(conn *Connection) {
	died := conn.checkIfAlive()
	if died {
		bot.Lock()
		defer bot.Unlock()
		if len(conn.joins) != 0 {
			i := getIndex(conn, bot.readconn)
			bot.readconn = append(bot.readconn[:i], bot.readconn[i+1:]...)
			bot.rejoinChannels(conn.joins)
		} else if conn == bot.whisperconn {
			bot.CreateConnection(connWhisperconn)
		} else {
			i := getIndex(conn, bot.readconn)
			bot.connlist = append(bot.connlist[:i], bot.connlist[i+1:]...)
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
		if bot.connlist[i].messages < 15 {
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
	bot.CreateConnection(connSendconn)
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
