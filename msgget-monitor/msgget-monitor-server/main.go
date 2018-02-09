/*
* main.go - the entry of program
*
* history
* --------------------
* 2018/1/12, by Ye Zhiqin, create
*
 */

package main

import (
	//"bytes"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AggrData struct {
	Endpoint    string
	Tags        string
	CounterType string
	Step        int64

	Timestamp int64

	ReqSum    int64
	Req200Sum int64
	Req499Sum int64
	Req500Sum int64
	Req502Sum int64
	Req504Sum int64
	IpSum     int64
	IdSum     int64
}

func Timeup(data *AggrData, t time.Time) {
	d2m, err := time.ParseDuration("-2m")
	if err != nil {
		log.Printf("calculate -2m duration FAIL: %s", err.Error())
		return
	}

	d62m, err := time.ParseDuration("-62m")
	if err != nil {
		log.Printf("calculate -62m duration FAIL: %s", err.Error())
		return
	}

	t2m := t.Add(d2m)
	t62m := t.Add(d62m)

	dp := t62m.Format("200601021504")
	rp := t2m.Format("200601021504")
	data.Timestamp = t2m.Unix()

	DelFromRedis(config.RedisServer, dp)
	ReadFromRedis(config.RedisServer, rp, data)
	Push2Falcon(config.FalconAgent, data)

	data.ReqSum = 0
	data.Req200Sum = 0
	data.Req499Sum = 0
	data.Req500Sum = 0
	data.Req502Sum = 0
	data.Req504Sum = 0
	data.IpSum = 0
	data.IdSum = 0
}

// main
func main() {
	sysCh := make(chan os.Signal, 1)
	signal.Notify(sysCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer close(sysCh)

	// set log format
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// load configuration
	config = LoadConfig()
	if config == nil {
		log.Printf("configuration loading FAIL, please check the config.yaml")
		os.Exit(-1)
	}

	// create one minute ticker
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	data := new(AggrData)
	data.Endpoint = config.Endpoint
	data.Tags = config.Tags
	data.CounterType = config.CounterType
	data.Step = config.Step

MAIN:
	for {
		select {
		case <-sysCh:
			log.Printf("system signal: %s", sysCh)
			break MAIN
		case t := <-ticker.C:
			Timeup(data, t)
		default:
			time.Sleep(time.Second)
		}
	}

	log.Printf("msgger-monitor-server exit...")
}
