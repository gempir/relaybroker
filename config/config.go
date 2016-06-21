package config

import (
    "io/ioutil"
    "encoding/json"
)

type Config struct {
	Broker_port string `json:"broker_port"`
	Broker_pass string `json:"broker_pass"`
}

func ReadConfig(path string) (Config, error) {
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
