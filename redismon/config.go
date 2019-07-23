package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var Cfg *Config

type Config struct {
	Lepus   string       `json:"lepus"`
	Manager string       `json:"manager"`
	Timeout int          `json:"timeout"`
	Falcon  FalconConfig `json:"falcon"`
}

type FalconConfig struct {
	API      string `json:"api"`
	Endpoint string `json:"endpoint"`
	Step     int64  `json:"step"`
}

// Load configuration from config.json
func LoadConfig() *Config {
	cfg := new(Config)

	buf, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("fail to read config.json: %s", err.Error())
		return nil
	}
	if err := json.Unmarshal(buf, cfg); err != nil {
		log.Printf("fail to unmarshal config.json: %s", err.Error())
		return nil
	}
	log.Printf("load application configuration successfully: %v", cfg)

	return cfg
}
