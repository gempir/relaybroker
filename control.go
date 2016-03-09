package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
)

// TCPServer simple tcp server for commands
func TCPServer() (ret int) {
	ln, err := net.Listen("tcp", ":"+TCPPort)
	if err != nil {
		log.Println("Error listening:", err.Error())
		return 1
	}
	log.Printf("Listening to port %s for connections...", TCPPort)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		bot := NewBot()
		bot.inconn = conn
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return 1
		}
		go handleRequest(conn, bot)
	}
}

// Start reading packets that are sent to the given connection
func handleRequest(conn net.Conn, bot *Bot) {
	// XXX(pajlada): Not sure if this is where we should close the connection or not
	// Perhaps in bot.ListenToConnection()?
	defer conn.Close()

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()

		if err != nil {
			fmt.Println("Read error:", err)
			fmt.Fprintf(conn, "Read error: %s", err)
			conn.Close()
			return
		}
		log.Println(line)
		handleMessage(line, bot)
	}
}

// Handle an IRC received from a bot
func handleMessage(message string, bot *Bot) {
	remoteAddr := bot.inconn.RemoteAddr().String()
	remoteAddrIP := strings.Split(remoteAddr, ":")

	if strings.Contains(message, "JOIN ") {
		joinComm := strings.Split(message, "JOIN ")
		bot.join(joinComm[1])
	} else if strings.Contains(message, "PASS ") {
		passComm := strings.Split(message, "PASS ")
		passwordParts := strings.Split(passComm[1], ";")
		if passwordParts[0] == TCPPass {
			bot.oauth = passwordParts[1]
			log.Printf("Authenticated! %s\n", remoteAddrIP)
		} else {
			log.Printf("Invalid broker pass! %s\n", remoteAddrIP)
			bot.inconn.Close()
			return
		}
	} else if strings.Contains(message, "NICK ") {
		nickComm := strings.Split(message, "NICK ")
		bot.nick = nickComm[1]

		if bot.oauth != "" {
			bot.CreateConnection()
		}
	} else if strings.Contains(message, "USER ") {
		if bot.nick != "" {
			nickComm := strings.Split(message, "USER ")
			bot.nick = nickComm[1]

			if bot.oauth != "" {
				bot.CreateConnection()
			}
		}
	} else if strings.Contains(message, "PRIVMSG ") {
		privmsgComm := strings.Split(message, "PRIVMSG ")
		remainingString := strings.Split(privmsgComm[1], " :")
		channel := remainingString[0]
		message := remainingString[1]
		bot.Message(channel, message)
	} else {
		log.Printf("Unhandled message: '%s'\n", message)
	}
}
