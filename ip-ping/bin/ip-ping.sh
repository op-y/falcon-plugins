#!/bin/bash

# global variables
readonly TS=$(date +%s)
readonly HOSTNAME=$(hostname)

if [[ $# -ne 2 ]];then
    echo "Parameter Error!"
    exit 1
else
    ip=$1
    group=$2
fi
   
loss=$(ping -c 30 -w 30 -W 30 -q "${ip}" | grep -oE "[0-9]*% packet loss" | awk -F'%' '{print $1}')

echo "${ip} loss percentage: ${loss}%"
curl -X POST -d "[{\"metric\": \"ip.loss\", \"endpoint\": \"$HOSTNAME\", \"timestamp\": $TS,\"step\": 60,\"value\": $loss,\"counterType\": \"GAUGE\",\"tags\": \"group=$group,ip=$ip\"}]" http://127.0.0.1:1988/v1/push
