package main

import (
	"reflect"
	"testing"
)

func TestCanReadConfig(t *testing.T) {
	cfg, err := readConfig("test_data/config.json")
    if err != nil {
        t.Error("error reading config", err)
    }
	if cfg.Broker_pass != "test" || cfg.Broker_port != "3333" {
		t.Error("invalid config")
	}
}

func TestCanInitLogger(t *testing.T) {
	log := initLogger()
	if reflect.TypeOf(log).String() != "logging.Logger" {
		t.Error("logger invalid type")
	}
}
