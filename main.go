package main

import (
	"os"
	"time"

	"github.com/op/go-logging"
)

var (
	// Log logger from go-logging
	Log        logging.Logger
	brokerPass string
	logLevel   logging.Level

	bots = make(map[string]*bot)

	// sync all bots joins since its ip based and not account based
	joinTicker = time.NewTicker(300 * time.Millisecond)
)

func main() {
	brokerPass = getEnv("BROKERPASS", "relaybroker")
	logLevel = getLogLevel(getEnv("LOGLEVEL", "info"))

	Log = initLogger(logLevel)
	server := new(Server)
	server.startServer()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getLogLevel(level string) logging.Level {
	switch level {
	case "debug":
		return logging.DEBUG
	case "error":
		return logging.ERROR
	default:
		return logging.INFO
	}
}

func initLogger(level logging.Level) logging.Logger {
	var logger *logging.Logger
	logger = logging.MustGetLogger("relaybroker")
	logging.SetLevel(level, "relaybroker")
	backend := logging.NewLogBackend(os.Stdout, "", 0)

	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02 15:04:05.000} %{level:.4s} %{shortfile}%{color:reset} %{message}`,
	)
	logging.SetFormatter(format)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, "relaybroker")
	logging.SetBackend(backendLeveled)
	return *logger
}
