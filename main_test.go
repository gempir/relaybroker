package main

import (
	"reflect"
	"testing"
)

func TestCanInitLogger(t *testing.T) {
	log := initLogger()
	if reflect.TypeOf(log).String() != "logging.Logger" {
		t.Fatal("logger invalid type")
	}
}

func TestCanReadConfig(t *testing.T) {
	cfg, err := readConfig("test_data/config.json")
	if err != nil {
		t.Fatal("error reading config", err)
	}
	if cfg.BrokerPass != "test" || cfg.BrokerPort != 3333 {
		t.Fatal("invalid config")
	}
}

func TestCanNotReadConfig(t *testing.T) {
	_, err := readConfig("test_data/invalid_file.json")
	if err == nil {
		t.Fatal("Invalid file but no error thrown", err)
	}
}

func TestCanUnmarshal(t *testing.T) {
	file := []byte(`{"broker_port": 3333, "broker_pass": "test"}`)
	cfg, err := unmarshalConfig(file)
	if err != nil {
		t.Fatal("failed to unmarshal config", err)
	}
	if cfg.BrokerPass != "test" || cfg.BrokerPort != 3333 {
		t.Fatal("invalid config")
	}
}

func TestCanNotUnmarshal(t *testing.T) {
	file := []byte(`{myInvalidJson}`)
	_, err := unmarshalConfig(file)
	if err == nil {
		t.Fatal("Didn't fail unmarshaling but should have")
	}
}
