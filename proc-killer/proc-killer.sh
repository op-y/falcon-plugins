#!/bin/bash

RETRY=3
HIT=0
THRESHOLD=50

PNAME="/path/to/proc/procname"
PROC="proc"

# get process id
PID=$(ps -ef | grep "${PNAME}" | grep -v "grep" | awk '{print $2}')

if [[ -z ${PID} ]];then
    echo "Can not find the process ID!"
    exit 1
fi

if [[ 1 -eq ${PID} ]];then
    echo "Process ID 1 is unexpected!"
    exit 1
fi

# get process cpu usage
for i in $(seq 1 ${RETRY});do
    CPU=$(ps -p "${PID}" -o pcpu --no-header)
    sleep 1
    echo "Try: ${i} CPU: ${CPU}"
    FLAG=$(echo "${CPU}>${THRESHOLD}" | bc)
    if [[ 1 -eq ${FLAG} ]];then
        HIT=$((HIT+1))
    fi
done

# kill process when cpu is busy
if [[ ${RETRY} -eq ${HIT} ]];then
    kill ${PID}
    sleep 5
    COMMAND=$(ps -p ${PID} -o comm --no-header)
    if [[ ${PROC} == ${COMMAND} ]];then
        kill ${PID}
        sleep 5
        COMMAND=$(ps -p ${PID} -o comm --no-header)
        if [[ ${PROC} == ${COMMAND} ]];then
            kill -9 ${PID}
            echo "kill process ${PID} force. Please check!!!"
            exit 2
        else
            echo "kill process ${PID} normally."
            exit 0
        fi
    else
        echo "kill process ${PID} normally."
        exit 0
    fi
fi

echo "CPU is not busy, killer go away."
exit 0
