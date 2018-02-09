/*
* falcon.go - the data structure of open falcon and related functions
*
* history
* --------------------
* 2017/8/18, by Ye Zhiqin, create
*
* DESCRIPTION
* This file contains the definition of open falcon data structure
* and the function to push data to open falcon
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type FalconData struct {
	Endpoint    string      `json:"endpoint"`
	Metric      string      `json:"metric"`
	Tags        string      `json:"tags"`
	CounterType string      `json:"counterType"`
	Step        int64       `json:"step"`
	Timestamp   int64       `json:"timestamp"`
	Value       interface{} `json:"value"`
}

/*
* SetValue - set FalconData value
*
* RECEIVER: *FalconData
*
* PARAMS:
*   - v: value
*
* RETURNS:
*   No return value
 */
func (data *FalconData) SetValue(v interface{}) {
	data.Value = v
}

/*
* String - generate a new FalconData
*
* RECEIVER: *FalconData
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   - string: string to display
 */
func (data *FalconData) String() string {
	s := fmt.Sprintf("FalconData Endpoint:%s Metric:%s Tags:%s CounterType:%s Step:%d Timestamp:%d Value:%v",
		data.Endpoint, data.Metric, data.Tags, data.CounterType, data.Step, data.Timestamp, data.Value)
	return s
}

/*
* NewFalconData - generate a new FalconData
*
* PARAMS:
*   - metric
*   - endpoint
*   - value
*   - counterType
*   - timestamp
*   - step
*
* RETURNS:
*   - *FalconData
 */
func NewFalconData(endpoint string, metric string, tags string, counterType string, step int64, timestamp int64, value interface{}) *FalconData {
	point := &FalconData{
		Endpoint:    GetEndpoint(endpoint),
		Metric:      metric,
		Tags:        tags,
		CounterType: counterType,
		Step:        step,
		Timestamp:   GetTimestamp(timestamp),
	}
	point.SetValue(value)
	return point
}

/*
* GetEndpoint - generate endpoint value
*
* PARAMS:
*   - endpoint
*
* RETURNS:
*   - endpoint: if endpoint is avaliable
*   - hostname: if endpoint is empty
*   - localhost: if endpoint is empty and can't get hostname
 */
func GetEndpoint(endpoint string) string {
	if endpoint != "" {
		return endpoint
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return hostname
}

/*
* GetTimestamp - generate timestamp value
*
* PARAMS:
*   - timestamp
*
* RETURNS:
*   - timestamp: if timestamp > 0
*   - now: if timestamp <= 0
 */
func GetTimestamp(timestamp int64) int64 {
	if timestamp > 0 {
		return timestamp
	} else {
		return time.Now().Unix()
	}
}

/*
* PushData - push data to open falcon
*
* PARAMS:
*   - api: url of agent or transfer
*   - data: an array of FalconData
*
* RETURNS:
*   - []byte, nil: if succeed
*   - nil, error: if fail
 */
func PushData(api string, data []*FalconData) ([]byte, error) {
	points, err := json.Marshal(data)
	if err != nil {
		log.Printf("data marshaling FAIL: %v", err)
		return nil, err
	}

	response, err := http.Post(api, "Content-Type: application/json", bytes.NewBuffer(points))
	if err != nil {
		log.Printf("api call FAIL: %v", err)
		return nil, err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

/*
* Push2Falcon - push file agent data
*
* PARAMS:
*   - api: falcon agent address
*   - data: AggrData
 */
func Push2Falcon(api string, data *AggrData) {
	var d []*FalconData

	// request.count.sum
	point := NewFalconData(data.Endpoint, "request.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.ReqSum)
	d = append(d, point)

	// request.200.count.sum
	point = NewFalconData(data.Endpoint, "request.200.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.Req200Sum)
	d = append(d, point)

	// request.499.count.sum
	point = NewFalconData(data.Endpoint, "request.499.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.Req499Sum)
	d = append(d, point)

	// request.500.count.sum
	point = NewFalconData(data.Endpoint, "request.500.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.Req500Sum)
	d = append(d, point)

	// request.502.count.sum
	point = NewFalconData(data.Endpoint, "request.502.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.Req502Sum)
	d = append(d, point)

	// request.504.count.sum
	point = NewFalconData(data.Endpoint, "request.504.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.Req504Sum)
	d = append(d, point)

	// ip.count.sum
	point = NewFalconData(data.Endpoint, "ip.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.IpSum)
	d = append(d, point)

	// shopid.count.sum
	point = NewFalconData(data.Endpoint, "shopid.count.sum", data.Tags, data.CounterType, data.Step, data.Timestamp, data.IdSum)
	d = append(d, point)

	//log.Printf("falcon points: %v", d)
	response, err := PushData(api, d)
	if err != nil {
		log.Printf("push data to falcon FAIL: %s", err.Error())
	}
	log.Printf("push data to falcon succeed: %s", string(response))
	return
}
