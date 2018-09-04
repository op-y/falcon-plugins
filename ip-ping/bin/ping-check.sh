#!/bin/bash

BIN_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONF_PATH=${BIN_PATH}/../conf
LOG_PATH=${BIN_PATH}/../log

echo "=====Start to Ping====="

for file in $(ls ${CONF_PATH});do
    group=${file##list-}
    echo "*****file:${file} group:${group}*****"
    
    while read ip || [[ -n "${ip}" ]];do
        echo "ping ${ip} ......" 
        sh -x ${BIN_PATH}/ip-ping.sh "${ip}" "${group}" &> ${LOG_PATH}/${ip}.log &
    done < ${CONF_PATH}/${file}
done

echo ">>>>>End of Ping<<<<<"
