package main

import (
	"fmt"
	"math/rand"
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
	test         []string
}

func newClient(conn net.Conn) Client {
	return Client{
		incomingConn: conn,
		fromClient:   make(chan string, 10),
		toClient:     make(chan string, 10),
		join:         make(chan string, 50000),
		test:         make([]string, 0),
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
	// cha := make(chan string, 5)
	// go c.relaybrokerCommand(cha)
	for msg := range c.toClient {
		c.incomingConn.Write([]byte(msg + "\r\n"))
		//cha <- msg
	}
	//closeChannel(cha)
}

func closeChannel(c chan string) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(r)
		}
	}()
	close(c)
}

func (c *Client) close() {
	closeChannel(c.join)
	// keep bot running if he wants to reconnect
	if c.bot.ID != "" {
		// dont let the channel fill up and block
		for m := range c.toClient {
			if c.bot.clientConnected {
				bots[c.ID].toClient <- m
				return
			}
			Log.Debug("msg on dc bot")
		}
	}
	if c.bot.clientConnected {
		return
	}
	closeChannel(c.fromClient)
	closeChannel(c.toClient)
	c.bot.close()
	delete(bots, c.ID)
	Log.Debug("CLOSED CLIENT", c.bot.nick)

}

func (c *Client) handleMessage(line string) {
	c.test = append(c.test, line)
	defer func() {
		if r := recover(); r != nil {
			Log.Error(c.test)
			Log.Error(r)
			c.close()
		}
	}()
	spl := strings.SplitN(line, " ", 2)
	msg := spl[1]
	// irc command
	switch spl[0] {
	case "LOGIN": // log into relaybroker with bot id to enable reconnecting, example: LOGIN pajbot2
		if bot, ok := bots[msg]; ok {
			c.ID = msg
			c.bot = bot
			c.bot.client.toClient = c.toClient
			close(c.join)
			c.join = make(chan string, 50000)
			go c.joinChannels()
			c.bot.clientConnected = true
			Log.Debug("old bot reconnected", msg)
			return
		}
		c.bot = newBot(c)
		c.ID = msg
		c.bot.ID = msg
		c.bot.clientConnected = true
		c.bot.Init()
		bots[msg] = c.bot
	case "PASS":
		pass := msg
		if strings.Contains(msg, ";") {
			passwords := strings.Split(msg, ";")
			pass = passwords[1]
			if cfg.BrokerPass != "" {
				if passwords[0] != cfg.BrokerPass {
					c.toClient <- "invalid relaybroker password\r\n"
					c.close()
					Log.Error("invalid relaybroker password")
					return
				}
			}
		}
		if c.bot == nil {
			c.bot = newBot(c)
			c.bot.Init()
		}
		c.bot.pass = pass
	case "NICK":
		c.bot.nick = strings.ToLower(msg) // make sure the nick is lowercase
		// generate random ID
		if c.bot.ID == "" {
			r := rand.Int31n(123456)
			ID := fmt.Sprintf("%d%s%d", 1, c.bot.nick, r)
			bots[ID] = c.bot
			c.ID = ID
		}
	case "JOIN":
		if c.bot == nil {
			c.bot = newBot(c)
			c.bot.Init()
		}
		c.join <- msg
	case "USER":
	default:
		go c.bot.handleMessage(spl)

	}
}
