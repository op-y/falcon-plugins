#!/bin/bash

LOG=/path/to/app.log

PATTERNS="PATTERN0|PATTERN1|PATTERN2"

REDIS_CLI=/path/to/redis-cli
REDIS_IP=127.0.0.1
REDIS_PORT=6379

stdbuf -oL tail -F ${LOG} |stdbuf -oL grep -oE "${PATTERNS}" | stdbuf -oL awk -f check.awk | "${REDIS_CLI}" -h "${REDIS_IP}" -p "${REDIS_PORT}"
