package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/op/go-logging"
)

func TestCanInitLogger(t *testing.T) {
	log := initLogger(0)
	assert.IsType(t, logging.Logger{}, log)
}

func TestCanReadConfig(t *testing.T) {
	cfg, err := readConfig("test_data/config.json")
	if err != nil {
		t.Fatal("error reading config", err)
	}
	assert.Equal(t, "test", cfg.BrokerPass)
	assert.Equal(t, "3333", cfg.BrokerPort)
}

func TestCanNotReadConfig(t *testing.T) {
	_, err := readConfig("test_data/invalid_file.json")
	if err == nil {
		t.Fatal("Invalid file but no error thrown", err)
	}
}

func TestCanUnmarshal(t *testing.T) {
	file := []byte(`{"broker_port": "3333","broker_pass": "test"}`)
	cfg, err := unmarshalConfig(file)
	if err != nil {
		t.Fatal("failed to unmarshal config")
	}
	assert.Equal(t, "test", cfg.BrokerPass)
	assert.Equal(t, "3333", cfg.BrokerPort)
}

func TestCanNotUnmarshal(t *testing.T) {
	file := []byte(`{myInvalidJson}`)
	_, err := unmarshalConfig(file)
	if err == nil {
		t.Fatal("Didn't fail unmarshaling but should have")
	}
}
