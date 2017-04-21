#!/bin/bash
#
# Author: yezhiqin@hualala.com
# Date: 2017/3/13
# Brief:
#   Get monitor data from log file.
# Globals:
#   TS
#   HOSTNAME
#   TIMESTAMP
#   PATTERN
#   APP_LOG
#   APP_TMP_LOG
#   MAX
#   MIN
#   AVG
#   TIME
# Arguments:
#   None
# Return:
#   None

# set

# global variables
readonly TS=$(date +%s -d "1 minute ago")
readonly HOSTNAME=$(hostname)

readonly TIMESTAMP=$(date '+%Y-%m-%d %H:%M:' -d '1 minute ago')
readonly PATTERN="rpc cost [[0-9]*]ms"

readonly APP_LOG=/path/to/app.log
readonly APP_TMP_LOG=app.1m.log

declare TIME=0

# usage
function usage() {
    echo "Usage: cost_monitor.sh"
}

#################################################
# Brief:
#   new day judgement.
# Globals:
#   TIMESTAMP
# Arguments:
#   None
# Returns:
#   0: date not changed
#   1: date changed
#################################################
function change_date() {
    if [[ "${TIMESTAMP}" =~ " 23:59:" ]];then
        return 1
    else
        return 0
    fi
}

#################################################
# Brief:
#   pick 1 minute log from log file.
# Globals:
#   TIMESTAMP
#   URL
# Arguments:
#   $1: in_file
#   $2: out_file
# Returns:
#   None
#################################################
function pick_log() {
    local in_file
    local out_file
    
    if [[ $# -ne 2 ]];then
        echo "Need 2 parameters!"
        exit 1
    else
        in_file="$1"
        out_file="$2"
    fi
    
    if [[ ! -f "${in_file}" ]];then
        echo "No such target log: ${in_file}"
        > "${out_file}"
        return 2
    fi
    
    grep "${TIMESTAMP}" "${in_file}" | grep -oE "${PATTERN}" > "${out_file}"
}

#################################################
# Brief:
#   calculate client cost.
# Globals:
#   MAX
#   MIN
#   AVG
# Arguments:
#   $1: log_file
# Returns:
#   None
#################################################
function calculate() {
    local log_file
    local result
    
    if [[ $# -ne 1 ]];then
        echo "Need 1 parameters!"
        exit 1
    else
        log_file="$1"
    fi
    
    if [[ -f ${log_file} && -s ${log_file} ]];then
        result=$(grep -oE "[0-9]*" ${log_file} | awk 'BEGIN{max=0;min=99999;avg=0;count=0;sum=0;} {count+=1;sum+=$1;if(max<$1)max=$1;if(min>$1)min=$1;} END{if(count!=0) avg=sum/count;print max" "min" "avg" "count" "sum}')
        MAX=$(echo "${result}" | awk '{print $1}')
        MIN=$(echo "${result}" | awk '{print $2}')
        AVG=$(echo "${result}" | awk '{print $3}')
    else
        MAX=0
        MIN=0
        AVG=0
    fi
}

#################################################
# Brief:
#   push data to falcon
# Globals:
#   HOSTNAME
#   MAX
#   MIN
#   AVG
#   TS
#   TIME
# Returns:
#   None
#################################################
function push() {
    if [[ -z "${MAX}" ]];then
        echo "Max: None"
    else
        echo "Max: ${MAX}"
        curl -X POST -d "[{\"metric\": \"client.cost.max\", \"endpoint\": \"$HOSTNAME\", \"timestamp\": $TS,\"step\": 60,\"value\": $MAX,\"counterType\": \"GAUGE\",\"tags\": \"module=app,container=1\"}]" http://127.0.0.1:1988/v1/push
    fi
    
    if [[ -z "${MIN}" ]];then
        echo "Min: None"
    else
        echo "Min: ${MIN}"
        curl -X POST -d "[{\"metric\": \"client.cost.min\", \"endpoint\": \"$HOSTNAME\", \"timestamp\": $TS,\"step\": 60,\"value\": $MIN,\"counterType\": \"GAUGE\",\"tags\": \"module=app,container=1\"}]" http://127.0.0.1:1988/v1/push
    fi
    
    if [[ -z "${AVG}" ]];then
        echo "Avg: None"
    else
        echo "Avg: ${AVG}"
        curl -X POST -d "[{\"metric\": \"client.cost.avg\", \"endpoint\": \"$HOSTNAME\", \"timestamp\": $TS,\"step\": 60,\"value\": $AVG,\"counterType\": \"GAUGE\",\"tags\": \"module=app,container=1\"}]" http://127.0.0.1:1988/v1/push
    fi
}

# main
function main() {
    local start
    local end
    local is_new_day
    
    start=$(date +%s)
    
    # new date judgement
    change_date
    is_new_day=$?
    
    if [[ 1 -eq "${is_new_day}" ]];then
        MAX=0
        MIN=0
        AVG=0
    else
        pick_log "${APP_LOG}" "${APP_TMP_LOG}"
        calculate "${APP_TMP_LOG}"
    fi
    
    end=$(date +%s)
    TIME=$((end-start))
    
    # node push data to redis and falcon
    push
}

main "$@"
