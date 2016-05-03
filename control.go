package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"strings"
)

var botlist []*Bot

// TCPServer simple tcp server for commands
func TCPServer() (ret int) {
	ln, err := net.Listen("tcp", ":"+TCPPort)
	if err != nil {
		log.Errorf("[control] error listening %v", err)
		return 1
	}
	log.Debugf("[control] listening to port %s for connections...", TCPPort)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		bot := NewBot()
		bot.inconn = conn
		botlist = append(botlist, bot)
		if err != nil {
			log.Errorf("[control] error accepting: %v", err)
			return 1
		}
		go handleRequest(conn, bot)
	}
}

// CloseBot close all connectons for a specific bot
func CloseBot(bot *Bot) {
	// Iterate over the list of bots
	for i := range botlist {
		if bot == botlist[i] {
			// Remove the closed bot from the list
			botlist = append(botlist[:i], botlist[i+1:]...)
			return
		}
	}

	// Let the bot clean itself up
	bot.Close()
}

func handleRequest(conn net.Conn, bot *Bot) {
	defer CloseBot(bot)
	defer bot.Close()

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
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
		go bot.Join(joinComm[1])
	} else if strings.HasPrefix(message, "PART ") {
		partComm := strings.Split(message, "PART ")
		go bot.Part(partComm[1])
	} else if strings.HasPrefix(message, "PASS ") {
		passComm := strings.Split(message, "PASS ")
		passwordParts := strings.Split(passComm[1], ";")
		if passwordParts[0] == TCPPass {
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
			} else if strings.HasPrefix(message, "USER ") {
				nickComm := strings.Split(message, "USER ")
				bot.nick = nickComm[1]
			}

			if bot.oauth != "" || strings.Contains(strings.ToLower(bot.nick), "justinfan") {
				bot.CreateConnection()
				go bot.CreateConnection()
				go bot.CreateConnection()
				go bot.CreateConnection()
				go bot.CreateConnection()
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
