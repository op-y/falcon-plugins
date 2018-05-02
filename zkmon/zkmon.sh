#!/bin/bash

#!/bin/bash
#
# Author: ye.zhiqin@outlook.com
# Date: 2018/4/20
# Brief:
#   Launch the processes provided by proc file.
# Globals:
#   ZK_PROC
#   ZK_ADDR
#   ZK_PORT
#   FALCON_API
# Arguments:
#   None
# Return:
#   None

# set
set -e
set -u
set -o pipefail

# global variables
ZK_PROC="QuorumPeerMain"
ZK_ADDR="127.0.0.1"
ZK_PORT="2181"
FALCON_API=http://127.0.0.1:1988/v1/push

# variables
endpoint=$(hostname)
ts=$(date +%s)
step=60
counterType="GAUGE"
tags="module=zookeeper"

# get zookeeper process id
pid=$(ps -ef | grep "${ZK_PROC}" | grep -v "grep" | awk '{print $2}')
if [[ -z ${pid} ]];then
    echo "can not find the process id!"
    exit 1
else
    echo "zookeeper process id is: ${pid}"
fi

if [[ 1 -eq ${pid} ]];then
    echo "process ID 1 is unexpected!"
    exit 1
fi

# get zookeeper cpu usage
zk_cpu=$(ps -p "${pid}" -o pcpu --no-header)
echo "zookeeper CPU is: ${zk_cpu}"

# get zookeeper mem usage
mem_K=$(ps -p "${pid}" -o rss --no-header)
mem_M=$(echo "scale=2;${mem_K} / 1024.0" | bc)
zk_mem=$(echo "scale=2;${mem_M} / 1024.0" | bc | awk '{printf "%.2f", $0}')
echo "zookeeper MEM is: ${zk_mem}"

# is zookeeper ok?
ruok_resp=$(echo ruok | nc ${ZK_ADDR} ${ZK_PORT})
if [[ "imok" == ${ruok_resp} ]];then
    zk_ok=0
    echo "zookeeper stat is: ${zk_ok}"
else
    zk_ok=1
    echo "zookeeper stat is: ${zk_ok}"
fi

# get zookeeper data
echo conf | nc ${ZK_ADDR} ${ZK_PORT} > conf.txt
echo stat | nc ${ZK_ADDR} ${ZK_PORT} > stat.txt
echo wchs | nc ${ZK_ADDR} ${ZK_PORT} > wchs.txt

zk_max_cons=$(grep -oE "maxClientCnxns=[0-9]*" conf.txt | awk -F"=" '{print $NF}')
echo "zookeeper max connection number is: ${zk_max_cons}"

zk_cons=$(grep -oE "Connections: [0-9]*" stat.txt | awk '{print $NF}')
echo "zookeeper connection number is: ${zk_cons}"

zk_free_cons=$((zk_max_cons - zk_cons))
echo "zookeeper free connection number is: ${zk_free_cons}"

mode=$(grep -oE "Mode: [a-z]*" stat.txt | awk '{print $NF}')
if [[ "leader" == ${mode} ]];then
    zk_leader=0
    echo "zookeeper mode is: ${zk_leader}"
else
    zk_leader=1
    echo "zookeeper mode is: ${zk_leader}"
fi

zk_wchs=$(grep -oE "watches:[0-9]*" wchs.txt | awk -F":" '{print $NF}')
echo "zookeeper watch count is: ${zk_wchs}"


# push to falcon
curl -X POST -d "[{\"metric\": \"zk.cpu\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${ts},\"step\": ${step},\"value\": ${zk_cpu},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"},{\"metric\": \"zk.mem\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${ts},\"step\": ${step},\"value\": ${zk_mem},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"},{\"metric\": \"zk.ok\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${ts},\"step\": ${step},\"value\": ${zk_ok},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"},{\"metric\": \"zk.cons.max\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${ts},\"step\": ${step},\"value\": ${zk_max_cons},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"},{\"metric\": \"zk.cons\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${ts},\"step\": ${step},\"value\": ${zk_cons},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"},{\"metric\": \"zk.cons.free\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${ts},\"step\": ${step},\"value\": ${zk_free_cons},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"},{\"metric\": \"zk.lead\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${ts},\"step\": ${step},\"value\": ${zk_leader},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"}]" ${FALCON_API}
