package main

import (
	"github.com/op/go-logging"
	"os"
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	Broker_port string `json:"broker_port"`
	Broker_pass string `json:"broker_pass"`
}

var (
	config Config
	log logging.Logger
)

func main() {
	log = initLogger()
	config = readConfig("config.json")
	log.Info("starting up...")
}

func initLogger() logging.Logger {
	var logger *logging.Logger
	logger = logging.MustGetLogger("relaybroker")
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	format   := logging.MustStringFormatter(`%{color}[%{time:2006-01-02 15:04:05}] [%{level:.4s}] %{color:reset}%{message}`)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)
	return *logger
}

func readConfig(path string) Config {
	var cfg Config
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
