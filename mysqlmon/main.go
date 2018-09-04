package main

import (
    "fmt"
    "log"
    "net"
    "time"
)

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    fmt.Println("=====mysqlmon=====")
    fmt.Printf("Start:%s\n", time.Now().Format("200601021504"))

    DBList = LoadDBList()
    if DBList == nil {
        panic("fail to load mysql instance configuration")
    }

    api := "http://127.0.0.1:1988/v1/push"
    from := GetEndpoint("")
    endpoint := "mysqlmon"
    metric := "mysql.alive"
    ts := time.Now().Unix()

    var data []*FalconData

    for _, instance := range DBList.Instances {
        if instance.Enabled {
            tags := fmt.Sprintf("from=%s,ip=%s,port=%d", from, instance.IP, instance.Port)
            if ok := DialMySQL(instance.Name); ok {
                fmt.Printf("MySQL %s is ok.\n", instance.Name)
                point := NewFalconData(metric, endpoint, 1, "GAUGE", tags, ts, 60)
                data = append(data, point)
            } else {
                fmt.Printf("MySQL %s is DOWN!!!\n", instance.Name)
                point := NewFalconData(metric, endpoint, 0, "GAUGE", tags, ts, 60)
                data = append(data, point)
            }
        }
    }

    response, err := PushToFalcon(api, data)
    if err != nil {
        log.Printf("fail to push falcon data: %s", err.Error)
    }
    fmt.Printf("push data to falcon successfully: %s\n", string(response))

    fmt.Printf("End:%s\n", time.Now().Format("200601021504"))
    return
}

func DialMySQL(address string) bool {
    conn, err := net.DialTimeout("tcp", address, time.Second*5)
    if err != nil {
        log.Printf("fail to connect mysql: %s", err.Error())
        return false
    }
    fmt.Printf("connect to mysql successfully: %s\n", conn.RemoteAddr().String())
    conn.Close()
    return true
}
