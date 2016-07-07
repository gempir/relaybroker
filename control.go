package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"
	"strconv"
)

var (
	botlist = make(map[string]*Bot)
	pendingBots []*Bot
	xd int
	tcpPass string
)

// TCPServer simple tcp server for commands
func TCPServer(brokerPort int, brokerPass string) (ret int) {
	tcpPass = brokerPass
	tcpPort := strconv.Itoa(brokerPort)

	ln, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		log.Errorf("[control] error listening %v", err)
		return 1
	}
	log.Debugf("[control] listening to port %s for connections...", tcpPort)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		bot := NewBot()
		go bot.Join()
		bot.inconn = conn
		pendingBots = append(pendingBots, bot)
		if err != nil {
			log.Errorf("[control] error accepting: %v", err)
			return 1
		}
		go handleRequest(conn, bot)
	}
}

// CloseBot close all connectons for a specific bot
func CloseBot(bot *Bot) {
	log.Debugf("closing bot %s", bot.nick)

	delete(botlist, bot.nick)
	// Let the bot clean itself up
	bot.Close()
}

func deletePendingBot(bot *Bot) {
	log.Debug("deleting bot")
	for i := range pendingBots {
		if bot == pendingBots[i] {
			// Remove the closed bot from the list
			pendingBots = append(pendingBots[:i], pendingBots[i+1:]...)
			return
		}
	}
}

func handleRequest(conn net.Conn, bot *Bot) {
	xd++
	log.Debug(xd)
	x := xd
	bot.handler[x] = true
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			<-ticker.C
			if !bot.handler[x] || !bot.open {
				return
			}
			bot.checkConnections()
		}
	}()
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	for bot.open && bot.handler[x] {
		line, err := tp.ReadLine()
		if err != nil {
			fmt.Println("[control] read error:", err)
			fmt.Fprintf(conn, "[control] read error: %s", err)
			conn.Close()
			return
		}
		if !strings.HasPrefix(line, "PASS ") {
			log.Debug("[control] " + line)
		}
		err = handleMessage(line, bot)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

// Handle an IRC received from a bot
func handleMessage(message string, bot *Bot) error {
	if !strings.HasPrefix(message, "PASS ") && !bot.login && !bot.anon {
		return errors.New("not authenticated")
	}

	if strings.HasPrefix(message, "JOIN ") {
		joinComm := strings.Split(message, "JOIN ")
		bot.join <- joinComm[1]
	} else if strings.HasPrefix(message, "PART ") {
		partComm := strings.Split(message, "PART ")
		go bot.Part(partComm[1])
	} else if strings.HasPrefix(message, "PASS ") {
		passComm := strings.Split(message, "PASS ")
		passwordParts := strings.Split(passComm[1], ";")
		if passwordParts[0] == tcpPass {
			bot.oauth = passwordParts[1]
			bot.login = true
			bot.anon = false
			log.Info("[control] authenticated!")
		} else {
			bot.inconn.Close()
			return errors.New("invalid broker pass")
		}
	} else if strings.HasPrefix(message, "NICK ") || strings.HasPrefix(message, "USER ") {
		if bot.nick == "" {
			if strings.HasPrefix(message, "NICK ") {
				nickComm := strings.Split(message, "NICK ")
				bot.nick = nickComm[1]
				if oldBot, ok := botlist[bot.nick]; ok {
					// replace bots
					botlist[bot.nick] = bot
					oldBot.open = false
					bot.CreateConnection(connWhisperconn)
					bot.CreateConnection(connSendconn)
					bot.readconn = oldBot.readconn
					for _, conn := range bot.readconn {
						go bot.ListenToConnection(conn)
					}
					oldBot.readconn = make([]*Connection, 0)
					oldBot.inconn = nil
					oldBot.Close()
					deletePendingBot(bot)
					var x int
					for k, v := range bot.handler {
						if v {
							x = k
						}
					}
					bot.handler[x] = false
					go handleRequest(bot.inconn, bot)

					return fmt.Errorf("reconnected old bot %s", oldBot.nick)
				}
				botlist[bot.nick] = bot
				deletePendingBot(bot)

			} else if strings.HasPrefix(message, "USER ") {
				nickComm := strings.Split(message, "USER ")
				bot.nick = nickComm[1]
			}

			if bot.oauth != "" || strings.Contains(strings.ToLower(bot.nick), "justinfan") {
				bot.CreateConnection(connReadconn)
				bot.CreateConnection(connReadconn)
				go bot.CreateConnection(connWhisperconn)
				go bot.CreateConnection(connSendconn)
				go bot.CreateConnection(connSendconn)
				go bot.CreateConnection(connSendconn)
			}
		}
	} else if strings.HasPrefix(message, "PRIVMSG #jtv :/w ") {
		privmsgComm := strings.Split(message, "PRIVMSG #jtv :")
		go bot.Whisper(privmsgComm[1])
	} else if strings.HasPrefix(message, "CAP ") {
		// Throw this message away
	} else {
		bot.Message(message)
	}
	return nil
}
