#!/bin/bash

BIN_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source ${BIN_PATH}/order_count.conf

# query from mysql
result=$(${mysql_client} -h${db_ip} -P${db_port} -u${db_username} -p${db_password} -e"use ${db_database}; ${sql}" | grep -v "-" | grep -v "+" | grep -v "in set" | grep -v "orderCount" | grep -v "orderTotalAmount")

orderCount=$(echo ${result} | awk '{print $1}')
if [[ -z "${orderCount}" || 0 -eq "${orderCount}" ]];then
    orderTotalAmount=0
else
    orderTotalAmount=$(echo ${result} | awk '{print $2}')
fi

# push to falcon
curl -X POST -d "[{\"metric\": \"order.count\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${timestamp},\"step\": ${step},\"value\": ${orderCount},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"}]" ${falcon_api}

usleep 10

curl -X POST -d "[{\"metric\": \"order.amount\", \"endpoint\": \"${endpoint}\", \"timestamp\": ${timestamp},\"step\": ${step},\"value\": ${orderTotalAmount},\"counterType\": \"${counterType}\",\"tags\": \"${tags}\"}]" ${falcon_api}
