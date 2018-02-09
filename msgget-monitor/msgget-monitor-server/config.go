/*
* config.go - the structure of configuration and related functions
*
* history
* --------------------
* 2018/1/11, by Ye Zhiqin, create
*
 */
package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	RedisServer string `yaml:"redis"`
	FalconAgent string `yaml:"agent"`
	Endpoint    string `yaml:"endpoint"`
	Tags        string `yaml:"tags"`
	CounterType string `yaml:"counterType"`
	Step        int64  `yaml:"step"`
}

var config *Config

/*
* LoadConfig - load configuration file to Config struct
*
* PARAMS:
* No paramter
*
* RETURNS:
* nil, if error ocurred
* *Config, if succeed
 */
func LoadConfig() *Config {
	cfg := new(Config)
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("configuration file reading FAIL: %s", err.Error())
		return nil
	}
	if err := yaml.Unmarshal(buf, cfg); err != nil {
		log.Printf("yaml file unmarshal FAIL: %s", err.Error())
		return nil
	}
	log.Printf("config: %v", cfg)

	// check configuration
	if cfg.RedisServer == "" {
		log.Printf("redis server address should not EMPTY!")
		return nil
	}

	if cfg.FalconAgent == "" {
		log.Printf("falcon agent address should not EMPTY!")
		return nil
	}

	return cfg
}
