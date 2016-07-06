package main

import (
	"net"
	"strings"
)

// Client for connection to relaybroker
type Client struct {
	ID           string
	bot          *bot
	incomingConn net.Conn
	fromClient   chan string
	toClient     chan string
	join         chan string // TODO: this should be some kind of priority queue
	connected    bool
}

func newClient(conn net.Conn) Client {
	return Client{
		incomingConn: conn,
		fromClient:   make(chan string, 10),
		toClient:     make(chan string, 10),
		join:         make(chan string, 10000000),
	}
}

func (c *Client) init() {
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
	// keep bot running if he wants to reconnect
	if c.ID != "" {
		// dont let the channel fill up and block
		for {
			if c.connected {
				return
			}
			<-c.toClient
		}
	}
	close(c.toClient)
	c.bot.close()
	Log.Debug("CLOSED CLIENT", c.bot.nick)

}

func (c *Client) handleMessage(line string) {
	spl := strings.SplitN(line, " ", 2)
	msg := spl[1]
	// irc command
	switch spl[0] {
	case "LOGIN": // log into relaybroker with bot id to enable reconnecting, example: LOGIN pajbot2
		if bot, ok := bots[msg]; ok {
			c.bot = bot
			c.bot.toClient = c.toClient
			c.ID = msg
			Log.Debug("old bot reconnected")
			return
		}
		c.bot = newBot(c.toClient)
		c.bot.Init()
		bots[msg] = c.bot
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
