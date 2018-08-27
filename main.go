package main

import (
	"os"
	"time"
)

var (
	brokerPass string

	bots = make(map[string]*bot)

	// sync all bots joins since its ip based and not account based
	joinTicker = time.NewTicker(300 * time.Millisecond)
)

func main() {
	brokerPass = getEnv("BROKERPASS", "relaybroker")

	server := new(Server)
	server.startServer()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
