package main

import (
    "io/ioutil"
    "log"

    "gopkg.in/yaml.v2"
)

type MySQLs struct {
    Instances []Instance `yaml:"instances"`
}

type Instance struct {
    Name    string `yaml:"name"`
    IP      string `yaml:"ip"`
    Port    int    `yaml:"port"`
    Enabled bool   `yaml:"enabled"`
}

var DBList *MySQLs

func LoadDBList() *MySQLs {
    dblist := new(MySQLs)

    buf, err := ioutil.ReadFile("instance.yaml")
    if err != nil {
        log.Printf("fail to read instance.yaml: %s", err.Error())
        return nil
    }
    if err := yaml.Unmarshal(buf, dblist); err != nil {
        log.Printf("fail to unmarshal dblist: %s", err.Error())
        return nil
    }
    log.Printf("load dblist successfully: %v", dblist)

    return dblist
}
