package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pajlada/pajbot2/common"
)

type brokerStats struct {
	totalMsgsSent     int
	totalMsgsReceived int
	startTime         time.Time
	cd                bool
}

var stats = &brokerStats{
	startTime: time.Now(),
}

func statsAPI() {
	http.HandleFunc("/"+cfg.APIPath, apiData)
}

func apiData(w http.ResponseWriter, r *http.Request) {
	ago := time.Since(stats.startTime)
	mins := int(ago.Minutes())

	m := mins % 60
	hours := (mins - m) / 60

	h := hours % 24
	days := (hours - h) / 24
	hours = hours - days*24

	mins = m
	uptime := fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	s := fmt.Sprintf("relaybroker stats: online for %s,  %d total connections, %d messages sent, %d messages received MrDestructoid ",
		uptime,
		countConns(),
		stats.totalMsgsSent,
		stats.totalMsgsReceived)
	w.Write([]byte(s))
}

func countConns() int {
	var i int
	for _, bot := range bots {
		i += len(bot.sendconns) + len(bot.readconns) + 1
	}
	return i
}

func (stats *brokerStats) resetCD() {
	stats.cd = false
}

func (c *Client) relaybrokerCommand(cha chan string) {

	parser := Parse{}
	for line := range cha {
		msg := parser.Parse(line)
		if !stats.cd {
			if msg.Type == common.MsgPrivmsg {
				text := msg.Text
				trigger := strings.Split(strings.ToLower(text), " ")[0]
				if trigger == "!relaybroker" {
					ago := time.Since(stats.startTime)
					mins := int(ago.Minutes())

					m := mins % 60
					hours := (mins - m) / 60

					h := hours % 24
					days := (hours - h) / 24
					hours = hours - days*24

					mins = m
					uptime := fmt.Sprintf("%dd %dh %dm", days, hours, mins)
					s := fmt.Sprintf("relaybroker stats: online for %s,  %d total connections, %d messages sent, %d messages received MrDestructoid ",
						uptime,
						len(c.bot.sendconns)+len(c.bot.readconns),
						stats.totalMsgsSent,
						stats.totalMsgsReceived)
					c.bot.say("#" + msg.Channel + " : " + s)
					stats.cd = true
					time.AfterFunc(30*time.Second, stats.resetCD)
				}
			}
		}

	}
}
