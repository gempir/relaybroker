package main

import ("net"
        "log"
        "bufio"
        "fmt"
        "net/textproto"
      )
type Bot struct{
        server string
        port string
        nick string
        user string
        channel string
        pass string
        pread, pwrite chan string
        conn net.Conn
}

func NewBot() *Bot {
        return &Bot{server: "irc.twitch.tv",
                    port: "6667",
                    nick: "justinfan321314364545123142435",
                    channel: "#nymn_hs",
                    pass: "",
                    conn: nil,
                    user: "blaze"}
}
func (bot *Bot) Connect() (conn net.Conn, err error){
  conn, err = net.Dial("tcp",bot.server + ":" + bot.port)
  if err != nil{
    log.Fatal("unable to connect to IRC server ", err)
  }
  bot.conn = conn
  log.Printf("Connected to IRC server %s (%s)\n", bot.server, bot.conn.RemoteAddr())
  return bot.conn, nil
}



func main(){
  ircbot := NewBot()
  conn, _ := ircbot.Connect()
  fmt.Fprintf(conn, "USER %s 8 * :%s\r\n", ircbot.nick, ircbot.nick)
  fmt.Fprintf(conn, "NICK %s\r\n", ircbot.nick)
  fmt.Fprintf(conn, "JOIN %s\r\n", ircbot.channel)
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
