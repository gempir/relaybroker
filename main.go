package main

import ("net"
        "log"
        "bufio"
        "fmt"
        "net/textproto"
      )

var channels = [...] string {"#pajlada","#nymn_hs"}


type Bot struct{
    server string
    port string
    nick string
    pread, pwrite chan string
    conn net.Conn
}

func NewBot() *Bot {
    return &Bot{
        server: "irc.twitch.tv",
        port: "6667",
        nick: "justinfan321314364545123142435",
        conn: nil,
    }
}

func (bot *Bot) Connect() (conn net.Conn, err error) {
    conn, err = net.Dial("tcp", bot.server + ":" + bot.port)
    if err != nil {
        log.Fatal("[ERROR] ", err)
    }
    bot.conn = conn
    return bot.conn, nil
}

func joinChannel(conn net.Conn, channel string) {
    fmt.Fprintf(conn, "JOIN %s\r\n", channel)
}

func main(){
    ircbot := NewBot()
    conn, _ := ircbot.Connect()
    fmt.Fprintf(conn, "USER %s 8 * :%s\r\n", ircbot.nick, ircbot.nick)
    fmt.Fprintf(conn, "NICK %s\r\n", ircbot.nick)


    for index,element := range channels {
        fmt.Print(index)
        joinChannel(conn, element)
    }

    defer conn.Close()

    reader := bufio.NewReader(conn)
    tp := textproto.NewReader( reader )

    for {
        line, err := tp.ReadLine()
        if err != nil {
            break // break loop on errors
        }
        fmt.Printf("%s\n", line)
    }

}
