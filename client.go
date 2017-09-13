package main

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
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
			Log.Error("error closing channel ")
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
	Log.Debug(line)
	c.test = append(c.test, line)
	defer func() {
		if r := recover(); r != nil {
			Log.Error("message handling error")
			c.close()
		}
	}()
	spl := strings.SplitN(line, " ", 2)
	msg := spl[1]
	if c.bot == nil {
		if c.registerBot(spl[0], msg) {
			return
		}
	}
	// irc command
	switch spl[0] {
	case "PASS":
		pass := msg
		if strings.Contains(msg, ";") {
			passwords := strings.Split(msg, ";")
			pass = passwords[1]
			if passwords[0] != brokerPass {
				c.toClient <- "invalid relaybroker password\r\n"
				c.close()
				Log.Error("invalid relaybroker password")
				return
			}
		}
		c.bot.pass = pass
	case "NICK":
		c.bot.nick = strings.ToLower(msg) // make sure the nick is lowercase
		// start bot when we got all login info
		c.bot.Init()
	case "JOIN":
		c.join <- msg
	case "USER":
	default:
		go c.bot.handleMessage(spl)
	}
}

/*
if first line from client == LOGIN, reconnect to old bot
if its something else, create new bot
return true on LOGIN, false on any other command so it can be processed further
*/
func (c *Client) registerBot(cmd string, msg string) bool {
	if cmd == "LOGIN" {
		if bot, ok := bots[msg]; ok {
			c.ID = msg
			c.bot = bot
			c.bot.client.toClient = c.toClient
			close(c.join)
			c.join = make(chan string, 50000)
			go c.joinChannels()
			c.bot.clientConnected = true
			Log.Debug("old bot reconnected", msg)
			return true
		}
		c.bot = newBot(c)
		c.ID = msg
		c.bot.ID = msg
		c.bot.clientConnected = true
		bots[msg] = c.bot
		return true
	}
	c.bot = newBot(c)
	// generate random ID
	if c.bot.ID == "" {
		rand.Seed(int64(time.Now().Nanosecond()))
		r := rand.Int31n(123456)
		ID := fmt.Sprintf("%s%d", c.bot.nick, r)
		bots[ID] = c.bot
		c.ID = ID
	}
	return false
}
