package main

import (
	"bufio"
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
		log.Errorf("[control] error listening: %v", err)
		return 1
	}
	log.Infof("[control] listening to port %s for connections...", TCPPort)
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

// CloseBot closes all bots
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
		handleMessage(line, bot)
	}
}

// Handle an IRC received from a bot
func handleMessage(message string, bot *Bot) {
	remoteAddr := bot.inconn.RemoteAddr().String()
	remoteAddrIP := strings.Split(remoteAddr, ":")

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
			log.Infof("[control] authenticated! %s\n", remoteAddrIP)
		} else {
			log.Infof("[control] invalid broker pass! %s\n", remoteAddrIP)
			bot.inconn.Close()
			return
		}
	} else if strings.HasPrefix(message, "NICK ") || strings.HasPrefix(message, "USER ") {
		if strings.HasPrefix(message, "NICK ") {
			nickComm := strings.Split(message, "NICK ")
			bot.nick = nickComm[1]
		} else if strings.HasPrefix(message, "USER ") {
			nickComm := strings.Split(message, "USER ")
			bot.nick = nickComm[1]
		}

		if bot.oauth != "" {
			bot.CreateConnection()
			go bot.CreateConnection()
			go bot.CreateConnection()
			go bot.CreateConnection()
			go bot.CreateConnection()
		}
	} else if strings.HasPrefix(message, "PRIVMSG #jtv :/w ") {
		privmsgComm := strings.Split(message, "PRIVMSG #jtv :")
		go bot.Whisper(privmsgComm[1])
	} else {
		bot.Message(message)
	}
}
