package main

import (
	"github.com/op/go-logging"
	"github.com/gempir/relaybroker/config"
	"os"
)

var (
	cfg config.Config
	log    logging.Logger
)

func main() {
	log = initLogger()
	cfg, err := config.ReadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Info("starting up on port", cfg.Broker_port)
}

func initLogger() logging.Logger {
	var logger *logging.Logger
	logger = logging.MustGetLogger("relaybroker")
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(`%{color}[%{time:2006-01-02 15:04:05}] [%{level:.4s}] %{color:reset}%{message}`)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)
	return *logger
}
