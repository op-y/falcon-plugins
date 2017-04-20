#!/bin/bash

# check 1st time.
if [ $(ps aux | grep -v "grep" | grep "apache-activemq" | wc -l) -eq 0 ]; then
    sleep 5
else
    exit 0
fi

# check 2nd time.
if [ $(ps aux | grep -v "grep" | grep "apache-activemq" | wc -l) -eq 0 ]; then
    sleep 5
else
    exit 0
fi

# check 3rd time.
if [ $(ps aux | grep -v "grep" | grep "apache-activemq" | wc -l) -eq 0 ]; then
    systemctl stop keepalived
    exit 1
else
    exit 0
fi
