#! /usr/bin/env python
#-*- coding:utf-8 -*-
#
# Author: ye.zhiqin@outlook.com
# Date  : 2017/11/6

import json
import logging
import shelve
import time
from falcon import Falcon
from zookeeper import ZooKeeper

logging.basicConfig(level=logging.INFO, filename='canal-monitor.log', filemode='a')

now = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(time.time()))
logging.info("=====Started: %s=====" % now)

# init falcon object
falcon = Falcon()

# current timestamp
tsNow = int(round(time.time() * 1000))

# read last timestamp from dump file
tsRec = shelve.open('canal-monitor.db', writeback=True)

# read data from zookeeper
endpoint="127.0.0.1:2181"
zookeeper = ZooKeeper(endpoint)
zookeeper.connect()

instances = zookeeper.listInstances(node='/otter/canal/destinations')
for instance in instances:
    logging.info("-----%s-----" % instance)
    node = "/otter/canal/destinations/%s/1001/cursor" % instance
    cursor, stat = zookeeper.getCursor(node)
    if cursor is None or cursor == '':
        # status-3: no data in zookeeper
        logging.info("%s: no data in zookeeper" % instance)
        tags = "from=127.0.0.1,instance=%s" % str(instance)
        falcon.add(value=3, tags=tags)
        continue
    # dump cursor data
    cursorObj = json.loads(cursor)
    ts = cursorObj['postion']['timestamp']
    if not tsRec.has_key(str(instance)):
        # status-1: first time to keep timestamp
        logging.info("%s: first time to keeper timestamp %d" % (instance, ts))
        tsRec[str(instance)] = ts
        tags = "from=127.0.0.1,instance=%s" % str(instance)
        falcon.add(value=1, tags=tags)
    else:
        lastTs = tsRec[str(instance)]
        logging.info("Last TS:%d <---> This TS:%d" % (lastTs, ts))
        if ts <= lastTs:
            # status-2: timestamp has no change
            logging.info("timestamp has no change")
            tags = "from=127.0.0.1,instance=%s" % str(instance)
            falcon.add(value=2, tags=tags)
        elif tsNow - ts > 3600000:
            # status-4: before one hour
            logging.info("before one hour")
            tags = "from=127.0.0.1,instance=%s" % str(instance)
            falcon.add(value=4, tags=tags)
        else:
            # status-0: normal
            logging.info("OK")
            tsRec[str(instance)] = ts
            tags = "from=127.0.0.1,instance=%s" % str(instance)
            falcon.add(value=0, tags=tags)

# close zookeeper connection
zookeeper.close()

# close and write back dump file
tsRec.close()

# push monitor data to falcon
falcon.push()
