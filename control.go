package main

import (
	//"bufio"
	"fmt"
	"io"
	//"io/ioutil"
	"net"
	"os"
	"strings"
)

var (
	auth = make([]string, 8)
)

// TCPServer simple tcp server for commands
func TCPServer(ircbot *Bot) {
	ln, err := net.Listen("tcp", ":"+TCPPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		ircbot.inconn = conn
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		go handleRequest(conn, ircbot)
	}
	fmt.Println("CLOSING TCPServerConnection")
}

func handleRequest(conn net.Conn, ircbot *Bot) {
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
				fmt.Printf("Command: '%s'\n", command)
				handleMessage(command, ircbot)
			}
		}
		/*
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(message)
		*/

		/*
			remoteAddr := conn.RemoteAddr().String()
			remoteAddrIP := strings.Split(remoteAddr, ":")
				if stringInSlice(remoteAddrIP[0], auth) {
					fmt.Printf("Handling message '%s'\n", message)
					handleMessage(message, ircbot)
					// conn.Write([]byte("Message received"))
				} else if message == "AUTH "+TCPPass {

					auth = append(auth, remoteAddrIP[0])
					fmt.Println(auth)
					conn.Write([]byte("Authenticated\r\n"))
				} else if strings.Contains(message, "PASS ") {
					fmt.Println(message)
					passComm := strings.Split(message, "PASS ")
					passwordParts := strings.Split(passComm[1], ":")
					if passwordParts[0] == TCPPass {
						auth = append(auth, remoteAddrIP[0])
						fmt.Printf("Authenticated! %s\n", auth)
						// conn.Write([]byte("Authenticated\r\n"))
					}
				} else {
					fmt.Printf("asd\n")
					conn.Write([]byte("not authenticated use \"AUTH password\" to authenticate"))
				}
		*/
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func handleMessage(message string, ircbot *Bot) {
	fmt.Printf("handleMessage(%s)\n", message)
	if strings.Contains(message, "JOIN ") {
		joinComm := strings.Split(message, "JOIN ")
		channels := strings.Split(joinComm[1], " ")
		go ircbot.HandleJoin(channels)
	} else if strings.Contains(message, "PRIVMSG ") {
		privmsgComm := strings.Split(message, "PRIVMSG ")
		remainingString := strings.Split(privmsgComm[1], " :")
		channel := remainingString[0]
		message := remainingString[1]
		go ircbot.Message(channel, message)
	}
}
