package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var ErrConn = errors.New("failed to connect redis")

func GetInfo(endpoint string, timeout time.Duration) (string, error) {
	opt := redis.DialConnectTimeout(timeout)
	conn, err := redis.Dial("tcp", endpoint, opt)
	if err != nil {
		log.Printf("failed to connect to redis %s: %s", endpoint, err.Error())
		return "", ErrConn
	}
	defer conn.Close()

	info, err := redis.String(conn.Do("INFO"))
	if err != nil {
		log.Printf("failed to get info from redis %s: %s", endpoint, err.Error())
		return "", err
	}
	return info, nil
}

func GetMaxMemory(endpoint string, timeout time.Duration) int64 {
	opt := redis.DialConnectTimeout(timeout)
	conn, err := redis.Dial("tcp", endpoint, opt)
	if err != nil {
		log.Printf("failed to connect to redis %s: %s", endpoint, err.Error())
		return 0
	}
	defer conn.Close()

	mm, err := redis.Strings(conn.Do("config", "get", "maxmemory"))
	if err != nil {
		log.Printf("failed to get max memory from redis %s: %s", endpoint, err.Error())
		return 0
	}

	if len(mm) != 2 {
		log.Printf("unexpect max memory result.")
		return 0
	}
	max, err := strconv.ParseInt(mm[1], 10, 64)
	if err != nil {
		log.Printf("failed to parse max memory: %s", err.Error())
		return 0
	}
	return max
}

func GetMaxClients(endpoint string, timeout time.Duration) int64 {
	opt := redis.DialConnectTimeout(timeout)
	conn, err := redis.Dial("tcp", endpoint, opt)
	if err != nil {
		log.Printf("failed to connect to redis %s: %s", endpoint, err.Error())
		return 0
	}
	defer conn.Close()

	mc, err := redis.Strings(conn.Do("config", "get", "maxclients"))
	if err != nil {
		log.Printf("failed to get max client from redis %s: %s", endpoint, err.Error())
		return 0
	}

	if len(mc) != 2 {
		log.Printf("unexpect max clients result.")
		return 0
	}
	max, err := strconv.ParseInt(mc[1], 10, 64)
	if err != nil {
		log.Printf("failed to parse max client: %s", err.Error())
		return 0
	}
	return max
}

func ParseMetrics(info string) (version, max_memory, used_memory, connected_clients, blocked_clients int64) {
	version = int64(3)
	max_memory = int64(0)
	used_memory = int64(0)
	connected_clients = int64(0)
	blocked_clients = int64(0)

	elements := strings.Split(info, "\n")
	for _, elem := range elements {
		if strings.HasPrefix(elem, "redis_version:") {
			kv := strings.Split(elem, ":")
			if strings.Compare(kv[1], "2") <= 0 {
				version = 2
			}
		} else if strings.HasPrefix(elem, "maxmemory:") {
			kv := strings.Split(elem, ":")
			v, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
			if err != nil {
				log.Printf("failed to parse maxmemory: %s", err.Error())
				continue
			}
			max_memory = v
		} else if strings.HasPrefix(elem, "used_memory:") {
			kv := strings.Split(elem, ":")
			v, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
			if err != nil {
				log.Printf("failed to parse used_memory: %s", err.Error())
				continue
			}
			used_memory = v
		} else if strings.HasPrefix(elem, "connected_clients:") {
			kv := strings.Split(elem, ":")
			v, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
			if err != nil {
				log.Printf("failed to parse connected_clients: %s", err.Error())
				continue
			}
			connected_clients = v
		} else if strings.HasPrefix(elem, "blocked_clients:") {
			kv := strings.Split(elem, ":")
			v, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
			if err != nil {
				log.Printf("failed to parse blocked_clients: %s", err.Error())
				continue
			}
			blocked_clients = v
		} else {
			continue
		}
	}
	return
}

func CheckRedis(endpoint string, timeout time.Duration) (ok bool, version, max_memory, used_memory, memory_used_percent, max_clients, connected_clients, blocked_clients int64) {
	ok = true
	version = int64(3)
	max_memory = int64(0)
	used_memory = int64(0)
	memory_used_percent = int64(0)
	max_clients = int64(0)
	connected_clients = int64(0)
	blocked_clients = int64(0)

	info, err := GetInfo(endpoint, timeout)
	if err == ErrConn {
		ok = false
		return
	}
	if err != nil {
		log.Printf("failed to get info")
		return
	}

	version, max_memory, used_memory, connected_clients, blocked_clients = ParseMetrics(info)

	if version <= 2 {
		max_memory = GetMaxMemory(endpoint, timeout)
	}

	max_clients = GetMaxClients(endpoint, timeout)

	memory_used_percent = int64(0)
	if max_memory > 0 {
		memory_used_percent = used_memory * int64(100) / max_memory
	}

	return
}
