/*
* main.go - the entry of program
*
* history
* --------------------
* 2018/1/11, by Ye Zhiqin, create
*
 */

package main

import (
	//"bytes"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	MAX_UNCHANGED_TIME = 5
	REQUEST_URI        = "/msg/get.htm"
)

type Record struct {
	Finish chan bool
	Agent  *FileAgent
}

var record *Record
var wg sync.WaitGroup

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

	StartAgent()

MAIN:
	for {
		select {
		case <-sysCh:
			log.Printf("system signal: %s", sysCh)
			StopAgent()
			break MAIN
		default:
			time.Sleep(time.Second)
		}
	}

	wg.Wait()
	log.Printf("msgget-monitor-client exit...")
}

/*
* StartAgent - generate the file agent by the configuration
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   No return value
 */
func StartAgent() {
	// initialize agent data
	data := new(AgentData)
	data.Endpoint = config.Endpoint
	data.Tags = config.Tags
	data.CounterType = config.CounterType
	data.Step = config.Step
	data.Pattern = REQUEST_URI

	data.TsStart = 0
	data.TsEnd = 0
	data.TsUpdate = 0

	data.ErrorCnt = 0
	data.ReqCnt = 0
	data.Req200Cnt = 0
	data.Req499Cnt = 0
	data.Req500Cnt = 0
	data.Req502Cnt = 0
	data.Req504Cnt = 0
	data.IpSet = nil
	data.IdSet = nil

	// initialize agent
	agent := new(FileAgent)
	agent.Filename = config.Path
	agent.Delimiter = config.Delimiter
	agent.Inotify = config.Inotify
	agent.SelfCheck = config.SelfCheck

	agent.File = nil
	agent.FileInfo = nil
	agent.LastOffset = 0
	agent.UnchangeTime = 0

	agent.Data = data

	// initialize agent record
	record := new(Record)
	ch := make(chan bool, 1)
	record.Finish = ch
	record.Agent = agent

	// launch file agent
	wg.Add(1)
	log.Printf("wg: %v", wg)
	if record.Agent.Inotify {
		go TailWithInotify(record.Agent, record.Finish)
	} else {
		go TailWithCheck(record.Agent, record.Finish)
	}
}

/*
* StopAgent - recall the file agent when program exit or configuration changed
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   No return value
 */
func StopAgent() {
	record.Finish <- true
	close(record.Finish)

	record = nil
}
