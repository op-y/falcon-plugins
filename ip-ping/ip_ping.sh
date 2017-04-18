#!/bin/bash

# global variables
readonly TS=$(date +%s)
readonly HOSTNAME=$(hostname)

function push() {
    if [[ $# -ne 2 ]];then
        exit 1
    else
        local ip=$1
        local loss=$2
    fi
    
    echo "${ip} loss percentage: ${loss}%"
    curl -X POST -d "[{\"metric\": \"vip.loss\", \"endpoint\": \"$HOSTNAME\", \"timestamp\": $TS,\"step\": 300,\"value\": $loss,\"counterType\": \"GAUGE\",\"tags\": \"module=nginx,vip=$ip\"}]" http://127.0.0.1:1988/v1/push
    echo ""
}

while read ip || [[ -n "${ip}" ]];do
    loss=$(ping -c 5 -w 5 -W 5 -q "${ip}" | grep -oE "[0-9]*% packet loss" | awk -F'%' '{print $1}')
    push "${ip}" "${loss}"
done < ip.list
