package main

import (
	"bufio"
	"fmt"
	"net"
)

var (
	auth = make([]string, 8)
)

// TCPServer simple tcp server for commands
func TCPServer(ircbot *Bot) {
	ln, err := net.Listen("tcp", ":"+TCPPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		go handleRequest(conn, ircbot)
	}

}

func handleRequest(conn net.Conn, ircbot *Bot) {

	message, _ := bufio.NewReader(conn).ReadString('\n')
	remoteAddr := conn.RemoteAddr().String()

	if stringInSlice(remoteAddr, auth) {
		handleMessage(message, ircbot)
		conn.Write([]byte("Message received."))
	} else if message == "auth "+TCPPass {
		auth = append(auth, remoteAddr)
		handleMessage(message, ircbot)
		conn.Write([]byte("Message received."))
	}
	fmt.Println(auth)
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
	fmt.Println(message)
}
