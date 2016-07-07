package main

import (
	"encoding/json"
	"io/ioutil"
	_ "net/http/pprof"
	"os"

	"github.com/op/go-logging"
)

var (
	cfg config
	// log logger from go-logging
	log logging.Logger
)

type config struct {
	BrokerPort int    `json:"broker_port"`
	BrokerPass string `json:"broker_pass"`
}

func main() {

	log = initLogger()
	cfg, err := readConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	exitCode := TCPServer(cfg.BrokerPort, cfg.BrokerPass)
	log.Error("Exit code: ", exitCode)
}

func initLogger() logging.Logger {
	var logger *logging.Logger
	logger = logging.MustGetLogger("relaybroker")
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02 15:04:05.000} %{shortfile:-15s} %{level:.4s}%{color:reset} %{message}`,
	)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)
	return *logger
}

func readConfig(path string) (config, error) {
	var cfg config
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	return unmarshalConfig(file)
}

func unmarshalConfig(file []byte) (config, error) {
	var cfg config
	err := json.Unmarshal(file, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
