/*
* redis.go - functions to push data to redis
*
* history
* --------------------
* 2017/1/11, by Ye Zhiqin, create
*
 */

package main

import (
	"log"

	"github.com/garyburd/redigo/redis"
)

func Get(endpoint string, key string) int64 {
	conn, err := redis.Dial("tcp", endpoint)
	if err != nil {
		log.Printf("connect to redis  %s FAiL", endpoint)
		return -1
	}
	defer conn.Close()

	result, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		log.Printf("Execute: GET %s FAiL", key)
		return -1
	}
	log.Printf("Result: %d, Execute: GET %s", result, key)
	return result
}

func Scard(endpoint string, key string) int64 {
	conn, err := redis.Dial("tcp", endpoint)
	if err != nil {
		log.Printf("connect to redis  %s FAiL", endpoint)
		return -1
	}
	defer conn.Close()

	result, err := redis.Int64(conn.Do("SCARD", key))
	if err != nil {
		log.Printf("Execute: SCARD %s FAiL", key)
		return -1
	}
	log.Printf("Result: %d, Execute: SCARD %s", result, key)
	return result
}

func Del(endpoint string, key string) {
	conn, err := redis.Dial("tcp", endpoint)
	if err != nil {
		log.Printf("connect to redis  %s FAiL", endpoint)
		return
	}
	defer conn.Close()

	result, err := redis.Int(conn.Do("DEL", key))
	if err != nil {
		log.Printf("Execute: DEL %s FAiL", key)
		return
	}
	log.Printf("Result: %d, Execute: DEL %s", result, key)
	return
}

func DelFromRedis(endpoint string, prefix string) {
	key := "request-" + prefix
	Del(endpoint, key)

	key = "request200-" + prefix
	Del(endpoint, key)

	key = "request499-" + prefix
	Del(endpoint, key)

	key = "request500-" + prefix
	Del(endpoint, key)

	key = "request502-" + prefix
	Del(endpoint, key)

	key = "request504-" + prefix
	Del(endpoint, key)

	key = "ip-" + prefix
	Del(endpoint, key)

	key = "shopid-" + prefix
	Del(endpoint, key)
}

func ReadFromRedis(endpoint string, prefix string, data *AggrData) {
	key := "request-" + prefix
	data.ReqSum = Get(endpoint, key)

	key = "request200-" + prefix
	data.Req200Sum = Get(endpoint, key)

	key = "request499-" + prefix
	data.Req499Sum = Get(endpoint, key)

	key = "request500-" + prefix
	data.Req500Sum = Get(endpoint, key)

	key = "request502-" + prefix
	data.Req502Sum = Get(endpoint, key)

	key = "request504-" + prefix
	data.Req504Sum = Get(endpoint, key)

	key = "ip-" + prefix
	data.IpSum = Scard(endpoint, key)

	key = "shopid-" + prefix
	data.IdSum = Scard(endpoint, key)
}
