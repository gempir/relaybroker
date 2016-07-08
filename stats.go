package main

import (
	"fmt"
	"net/http"
	"time"
)

type brokerStats struct {
	totalMsgsSent     int
	totalMsgsReceived int
	startTime         time.Time
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
	botCount := fmt.Sprintf("%d bot", len(bots))
	if len(bots) > 1 {
		botCount += "s"
	}
	s := fmt.Sprintf("relaybroker stats: online for %s, %s connected, %d total connections, %d messages sent, %d messages received MrDestructoid ",
		uptime,
		botCount,
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
