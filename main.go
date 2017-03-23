package main

import (
	"encoding/json"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/op/go-logging"
)

var (
	cfg config
	// Log logger from go-logging
	Log logging.Logger

	bots = make(map[string]*bot)

	// sync all bots joins since its ip based and not account based
	joinTicker = time.NewTicker(300 * time.Millisecond)
)

type config struct {
	BrokerPort string `json:"broker_port"`
	BrokerPass string `json:"broker_pass"`
	APIHost    string `json:"api_host"`
	APIPath    string `json:"api_path"`
}

func main() {
	loggerArgs := os.Args
	var level = logging.INFO
	if len(loggerArgs) > 1 {
		switch loggerArgs[1] {
		case "debug":
			level = logging.DEBUG
		case "error":
			level = logging.ERROR
		default:
			level = logging.INFO
		}
	} else {
		level = logging.INFO
	}
	Log = initLogger(level)
	Log.Infof("running in %s mode, switch by typing ./relaybroker debug/error", level.String())
	var err error
	cfg, err = readConfig("config.json")
	if err != nil {
		Log.Fatal(err)
	}

	Log.Infof("starting up on port %s", cfg.BrokerPort)
	server := new(Server)

	server.startServer()

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

func readConfig(path string) (config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	return unmarshalConfig(file)
}

func unmarshalConfig(file []byte) (config, error) {
	err := json.Unmarshal(file, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
