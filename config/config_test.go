package config

import (
    "testing"
)

func TestCanReadConfig(t *testing.T) {
	cfg, err := ReadConfig("../test_data/config.json")
    if err != nil {
        t.Fatal("error reading config", err)
    }
	if cfg.Broker_pass != "test" || cfg.Broker_port != "3333" {
		t.Fatal("invalid config")
	}
}

func TestCanUnmarshal(t *testing.T) {
    file := []byte(`{"broker_port": "3333","broker_pass": "test"}`)
    cfg, err := unmarshalConfig(file)
    if err != nil {
        t.Fatal("failed to unmarshal config")
    }
    if cfg.Broker_pass != "test" || cfg.Broker_port != "3333" {
        t.Fatal("invalid config")
    }
}
