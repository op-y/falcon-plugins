set -x

WORKSPACE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd ${WORKSPACE}

HOST=127.0.0.1

REDIS_CLI=./redis-cli
REDIS_IP=127.0.0.1
REDIS_PORT=6379

OOM=$(${REDIS_CLI} -h ${REDIS_IP} -p ${REDIS_PORT} "GET" "oom.key")

if [[ -z ${OOM} ]];then
    OOM=0
fi

if [ "${OOM}" -gt "0" ];then
    echo "OOM=${OOM} Restart APP!"
    docker restart container-name
fi

${REDIS_CLI} -h ${REDIS_IP} -p ${REDIS_PORT} "DEL" "oom.key"
