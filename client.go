package main

import (
	"net"
	"strings"
)

// Client for connection to relaybroker
type Client struct {
	bot          *bot
	incomingConn net.Conn
	fromClient   chan string
	toClient     chan string
	join         chan string
}

func newClient(conn net.Conn) Client {
	return Client{
		incomingConn: conn,
		fromClient:   make(chan string, 10),
		toClient:     make(chan string, 10),
		join:         make(chan string, 10000000),
	}
}

func (c *Client) Init() {
	go c.joinChannels()
	go c.read()
}

func (c *Client) joinChannels() {
	for channel := range c.join {
		c.bot.join <- channel
	}
}

func (c *Client) read() {
	for msg := range c.toClient {
		c.incomingConn.Write([]byte(msg + "\r\n"))
	}
}

func (c *Client) close() {
	close(c.join)
	close(c.fromClient)
}

func (c *Client) handleMessage(line string) {
	Log.Debug(line)
	spl := strings.SplitN(line, " ", 2)
	msg := spl[1]
	// irc command
	switch spl[0] {
	case "LOGIN": // log into relaybroker with bot id, example: LOGIN pajbot2
		if bot, ok := bots[msg]; ok {
			c.bot = bot
			c.bot.toClient = c.toClient
			Log.Debug("old bot reconnected")
		}
	case "PASS":
		pass := msg
		if strings.HasPrefix(msg, "test;") {
			pass = strings.Split(msg, ";")[1]
		}
		if c.bot == nil {
			c.bot = newBot(c.toClient)
			c.bot.Init()
		}
		c.bot.pass = pass
	case "NICK":
		c.bot.nick = strings.ToLower(msg) // make sure the nick is lowercase
	case "JOIN":
		c.join <- msg
	case "USER":
	default:
		go c.bot.handleMessage(spl)

	}
}
