#!/bin/bash

LOG=/path/to/app.log

PATTERNS="java.lang.OutOfMemoryError"

REDIS_CLI=./redis-cli
REDIS_IP=127.0.0.1
REDIS_PORT=6379

stdbuf -oL tail -F ${LOG} |stdbuf -oL grep -oE "${PATTERNS}" | stdbuf -oL awk -f collector.awk | "${REDIS_CLI}" -h "${REDIS_IP}" -p "${REDIS_PORT}"
