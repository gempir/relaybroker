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

	botCount := fmt.Sprintf("%d bot", len(bots))
	if len(bots) > 1 {
		botCount += "s"
	}
	s := fmt.Sprintf("relaybroker stats: online for %s, %s connected, %d total connections, %d messages sent, %d messages received MrDestructoid ",
		getUptime(),
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

func getUptime() string {
	ago := time.Since(stats.startTime)
	totalSeconds := int(ago.Seconds())
	s := totalSeconds % 60
	mins := (totalSeconds - s) / 60
	seconds := totalSeconds - mins*60
	m := mins % 60
	mins = m
	hours := (mins - m) / 60

	h := hours % 24
	days := (hours - h) / 24
	hours = hours - days*24

	if mins+hours+days < 1 {
		// up for less than 1 minute
		return fmt.Sprintf("%ds", seconds)
	} else if hours+days < 1 {
		// up for less than 1 hour
		return fmt.Sprintf("%dm %ds", mins, seconds)
	} else if days < 1 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, seconds)
	}
	return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
}
