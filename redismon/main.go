package main

import (
	"fmt"
	"log"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	Cfg = LoadConfig()
	if Cfg == nil {
		panic("fail to load application configuration")
	}
}

func main() {
	fmt.Println("=====redismon=====")
	fmt.Printf("Start:%s\n", time.Now().Format("200601021504"))

	lepus, err := Connect(Cfg.Lepus)
	if err != nil {
		log.Printf("failed to connect to the lepus: %s", err.Error())
		return
	}
	defer lepus.Close()

	manager, err := Connect(Cfg.Manager)
	if err != nil {
		log.Printf("failed to connect to the db_manager: %s", err.Error())
		return
	}
	defer manager.Close()

	instances, err := GetInstances(lepus)
	if err != nil {
		log.Printf("failed to get redis instance from lepus: %s", err.Error())
		return
	}

	count := time.Duration(Cfg.Timeout)

	fep := Cfg.Falcon.Endpoint
	step := Cfg.Falcon.Step

	ts := time.Now().Unix()

	for _, instance := range instances {
		timeout := time.Second * count
		endpoint := fmt.Sprintf("%s:%s", instance.Host, instance.Port)
		host := instance.Host
		port := instance.Port
		itags := instance.Tags
		role := instance.Role
		fmt.Printf("Endpoint: %s\n", endpoint)
		fmt.Printf("Role: %s\n", role)
		ok, version, max_memory, used_memory, memory_used_percent, max_clients, connected_clients, blocked_clients := CheckRedis(endpoint, timeout)
		is_alive := int64(1)
		if !ok {
			is_alive = int64(0)
		}
		fmt.Printf("alive: %t\nversion: %d\nmax_memory: %d, used_memory: %d, memory_used_percent: %d\nmax_clients: %d, connected_clients: %d, blocked_clients: %d\n\n", ok, version, max_memory, used_memory, memory_used_percent, max_clients, connected_clients, blocked_clients)

		// prepare and push falcon data
		from := GetEndpoint("")
		tags := fmt.Sprintf("from=%s,host=%s,port=%s", from, instance.Host, instance.Port)

		data := []*FalconData{}
		point := NewFalconData("redis.alive", fep, is_alive, "GAUGE", tags, ts, step)
		data = append(data, point)
		point = NewFalconData("redis.version", fep, version, "GAUGE", tags, ts, step)
		data = append(data, point)
		point = NewFalconData("redis.max.memory", fep, max_memory, "GAUGE", tags, ts, step)
		data = append(data, point)
		point = NewFalconData("redis.used.memory", fep, used_memory, "GAUGE", tags, ts, step)
		data = append(data, point)
		point = NewFalconData("redis.memory.used.percent", fep, memory_used_percent, "GAUGE", tags, ts, step)
		data = append(data, point)
		point = NewFalconData("redis.max.clients", fep, max_clients, "GAUGE", tags, ts, step)
		data = append(data, point)
		point = NewFalconData("redis.connected.clients", fep, connected_clients, "GAUGE", tags, ts, step)
		data = append(data, point)
		point = NewFalconData("redis.blocked.clients", fep, blocked_clients, "GAUGE", tags, ts, step)
		data = append(data, point)
		response, err := PushToFalcon(Cfg.Falcon.API, data)
		if err != nil {
			log.Printf("failed to push falcon data: %s", err.Error())
			continue
		}
		fmt.Printf("push data to falcon successfully: %s\n", string(response))

		if err := UpdateStatus(manager, itags, host, port, role, is_alive, memory_used_percent, max_clients, connected_clients, blocked_clients); err != nil {
			log.Printf("failed to save monitor data: %s", err.Error())
			continue
		}
	}
	fmt.Printf("End:%s\n", time.Now().Format("200601021504"))
	return
}
