#! /usr/bin/env python
#-*- coding:utf-8 -*-
#
# Author: ye.zhiqin@outlook.com
# Date  : 2017/11/6

import json
import logging
import requests
import time

class Falcon:
    def __init__(self, url="http://127.0.0.1:1988/v1/push"):
        logging.basicConfig()
        self.url=url
        self.points = []

    def add(self, value, tags, metric="canal.status", endpoint="canal.app", step=900, counterType="GAUGE"):
        point = {
            'endpoint': endpoint,
            'metric': metric,
            'timestamp': int(time.time()),
            'step': step,
            'value': value,
            'counterType': counterType,
            'tags': tags
        }
        self.points.append(point)

    def push(self):
        logging.info("push data to falcon: %s" % json.dumps(self.points))
        response = requests.post(self.url, json.dumps(self.points), timeout=5)
        code = response.status_code
        text = response.text
        logging.info("push data to falcon, status code is %d and response text is %s" % (code,text))
