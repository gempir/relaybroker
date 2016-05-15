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
	log    = logging.MustGetLogger("relaybroker")
	format = logging.MustStringFormatter(
		`%{color}[%{time:2006-01-02 15:04:05}] [%{level:.4s}] %{color:reset}%{message}`,
	)
	joins = 0
)

func main() {
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)

	// Read config file
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(err)
	}

	TCPServer()
}
