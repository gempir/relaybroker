package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
)

func joinChannel(conn net.Conn, channel string) {
	fmt.Fprintf(conn, "JOIN %s\r\n", channel)
}
func main() {
	handleConn()
}

func sendMessage(conn net.Conn, channel string, message string) {
	fmt.Fprintf(conn, "PRIVMSG %s :%s\r\n", channel, message)
}

func handleConn() {
	var conn net.Conn
	server := "irc.twitch.tv:6667"
	nick := "gempir"
	pass := ""
	channel := "#nymn_hs"

	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Fatal("unable to connect to IRC server ", err)
	}

	fmt.Fprintf(conn, "PASS %s\r\n", pass)
	fmt.Fprintf(conn, "USER %s\r\n", nick)
	fmt.Fprintf(conn, "NICK %s\r\n", nick)

	joinChannel(conn, channel)

	sendMessage(conn, channel, "Kappa")

	defer conn.Close()
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	for {
		line, err := tp.ReadLine()
		if err != nil {
			break // break loop on errors
		}
		fmt.Printf("%s\n", line)
	}
}
