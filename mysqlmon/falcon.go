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
    Metric      string      `json:"metric"`
    Endpoint    string      `json:"endpoint"`
    Value       interface{} `json:"value"`
    CounterType string      `json:"counterType"`
    Tags        string      `json:"tags"`
    Timestamp   int64       `json:"timestamp"`
    Step        int64       `json:"step"`
}

func (data *FalconData) SetValue(v interface{}) {
    data.Value = v
}

func (data *FalconData) String() string {
    s := fmt.Sprintf("FalconData Metric:%s Endpoint:%s Value:%v CounterType:%s Tags:%s Timestamp:%d Step:%d",
        data.Metric, data.Endpoint, data.Value, data.CounterType, data.Tags, data.Timestamp, data.Step)
    return s
}

func NewFalconData(metric string, endpoint string, value interface{}, counterType string, tags string, timestamp int64, step int64) *FalconData {
    point := &FalconData{
        Metric:      metric,
        Endpoint:    GetEndpoint(endpoint),
        CounterType: counterType,
        Tags:        tags,
        Timestamp:   GetTimestamp(timestamp),
        Step:        step,
    }
    point.SetValue(value)
    return point
}

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

func GetTimestamp(timestamp int64) int64 {
    if timestamp > 0 {
        return timestamp
    } else {
        return time.Now().Unix()
    }
}

func PushToFalcon(api string, data []*FalconData) ([]byte, error) {
    points, err := json.Marshal(data)
    if err != nil {
        log.Printf("fail to convert data to json: %s", err.Error())
        return nil, err
    }

    response, err := http.Post(api, "Content-Type: application/json", bytes.NewBuffer(points))
    if err != nil {
        log.Printf("fail to call falcon API: %s", err.Error())
        return nil, err
    }
    defer response.Body.Close()

    content, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }

    return content, nil
}
