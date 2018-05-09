#! /usr/bin/env python
#-*- coding:utf-8 -*-
#
# Author: ye.zhiqin@outlook.com
# Date  : 2017/11/6

import logging
from kazoo.client import KazooClient
from kazoo.retry import KazooRetry

class ZooKeeper:
    def __init__(self, endpoint):
        logging.basicConfig()
        self.endpoint=endpoint
        self.zk=None

    def connect(self):
        retry = KazooRetry(max_tries=-1, max_delay=5)
        self.zk = KazooClient(hosts=self.endpoint, connection_retry=retry)
        self.zk.start()

    def close(self):
        self.zk.stop()

    def listInstances(self, node='/otter/canal/destinations'):
        if self.zk.exists(node):
            instances = self.zk.get_children(node)
            return instances
        else:
            return None

    def getCursor(self, node):
        if self.zk.exists(node):
            return self.zk.get(node)
        else:
            return (None, None)

# __main__
if __name__=='__main__':
    endpoint="127.0.0.1:2181"
    zookeeper = ZooKeeper(endpoint)
    zookeeper.connect()
    instances = zookeeper.listInstances(node='/otter/canal/destinations')
    print instances
    cursor, stat = zookeeper.getCursor(node='/otter/canal/destinations/db_order/1001/cursor')
    print cursor
    print stat
    zookeeper.close()
