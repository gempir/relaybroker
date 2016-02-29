package main

import (
	"fmt"
	"io"
	"log"
	"net"
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
		}
		go handleRequest(conn, bot)
	}
}

// Start reading packets that are sent to the given connection
func handleRequest(conn net.Conn, bot *Bot) {
	// XXX(pajlada): Not sure if this is where we should close the connection or not
	// Perhaps in bot.ListenToConnection()?
	defer conn.Close()

	for {
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 256)
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Read error:", err)
			}
			break
		}
		buf = append(buf, tmp[:n]...)
		x := string(buf)
		commands := strings.Split(x, "\r\n")
		for _, command := range commands {
			if command != "" {
				handleMessage(command, bot)
				log.Println(command)
			}
		}
	}
}

// Handle an IRC received from a bot
func handleMessage(message string, bot *Bot) {
	if strings.Contains(message, "JOIN ") {
		joinComm := strings.Split(message, "JOIN ")
		channels := strings.Split(joinComm[1], " ")
		bot.HandleJoin(channels)
	} else if strings.Contains(message, "PASS ") {
		passComm := strings.Split(message, "PASS ")
		passwordParts := strings.Split(passComm[1], ";")
		if passwordParts[0] == TCPPass {
			bot.oauth = passwordParts[1]
			remoteAddr := bot.inconn.RemoteAddr().String()
			remoteAddrIP := strings.Split(remoteAddr, ":")
			log.Printf("Authenticated! %s\n", remoteAddrIP)
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
		go bot.Message(channel, message)
	} else {
		log.Printf("Unhandled message: '%s'\n", message)
		//bot.WriteToAllConns(message)
	}
}
