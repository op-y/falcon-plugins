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
	"time"

	"github.com/garyburd/redigo/redis"
)

func IncrBy(endpoint string, key string, value int64) {
	conn, err := redis.Dial("tcp", endpoint)
	if err != nil {
		log.Printf("connect to redis  %s FAiL", endpoint)
		return
	}
	defer conn.Close()

	result, err := redis.Int(conn.Do("INCRBY", key, value))
	if err != nil {
		log.Printf("Execute: INCRBY %s %d FAiL", key, value)
		return
	}
	log.Printf("Result: %d, Execute: INCRBY %s %d", result, key, value)
	return
}

func Sadd(endpoint string, key string, value []string) {
	conn, err := redis.Dial("tcp", endpoint)
	if err != nil {
		log.Printf("connect to redis  %s FAiL", endpoint)
		return
	}
	defer conn.Close()

	c := 0
	for _, id := range value {
		result, err := redis.Int(conn.Do("SADD", key, id))
		if err != nil {
			log.Printf("Execute: SADD %s %s FAiL", key, id)
		}
		c += result
	}
	log.Printf("Result: %d, Execute: SADD %s ...", c, key)
	return
}

func Push2Redis(endpoint string, data *AgentData) {
	start := time.Unix(data.TsStart, 0)
	prefix := start.Format("200601021504")

	key := "request-" + prefix
	IncrBy(endpoint, key, data.ReqCnt)

	key = "request200-" + prefix
	IncrBy(endpoint, key, data.Req200Cnt)

	key = "request499-" + prefix
	IncrBy(endpoint, key, data.Req499Cnt)

	key = "request500-" + prefix
	IncrBy(endpoint, key, data.Req500Cnt)

	key = "request502-" + prefix
	IncrBy(endpoint, key, data.Req502Cnt)

	key = "request504-" + prefix
	IncrBy(endpoint, key, data.Req504Cnt)

	key = "ip-" + prefix
	Sadd(endpoint, key, data.IpSet.ToSlice())

	key = "shopid-" + prefix
	Sadd(endpoint, key, data.IdSet.ToSlice())
}
