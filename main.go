package main

import (
	"encoding/json"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"strconv"
)

var (
	cfg Config
	log logging.Logger
)

type Config struct {
	Broker_port string `json:"broker_port"`
	Broker_pass string `json:"broker_pass"`
}

func main() {
	log = initLogger()
	cfg, err := readConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	log.Info("starting up on port", cfg.Broker_port)
	server := new(Server)
	port, err := strconv.Atoi(cfg.Broker_port)
	if err != nil {
		panic("can't parse broker port")
	}
	server.startServer(port)
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

func readConfig(path string) (Config, error) {
	var cfg Config
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	return unmarshalConfig(file)
}

func unmarshalConfig(file []byte) (Config, error) {
	var cfg Config
	err := json.Unmarshal(file, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
