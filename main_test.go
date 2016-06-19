package main

import (
    "testing"
    "reflect"
)

func TestCanReadConfig(t *testing.T) {
    cfg := readConfig("test_data/config.json")
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
