package main

import (
	"github.com/op/go-logging"
)

var (
	log    = logging.MustGetLogger("relaybroker")
	format  = logging.MustStringFormatter(
		`%{color}[%{time:2006-01-02 15:04:05}] [%{level:.4s}] %{color:reset}%{message}`,
	)
)

func main() {
	TCPServer()
}
