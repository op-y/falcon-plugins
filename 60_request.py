#! /usr/bin/env python
#-*- coding:utf-8 -*-
#
# Author: yezhiqin@hualala.com
# Date  : 2016/11/11

import os
import json
import time
import requests

class RequestPatrol:
    def __init__(self, config_file):
        self.config_file = config_file
        with open(self.config_file) as file:
            self.data = []
            config = json.load(file)
            self.endpoint = os.uname()[1]
            self.step = config['step']
            self.counterType = config['counterType']
            self.tags = config['tags']
            self.request = config['request']

    # function to dispatch checking items.
    def HandleRequest(self):
        for item in self.request:
            metric = item['metric']
            protocol = item['protocol']
            type = item['type']
            url = item.get('url','')
            method = item.get('method','get')
            ip = item.get('ip','')
            port = item.get('port','')
            ping = item.get('ping','')
            pong = item.get('pong','')
            if protocol == 'http':
                value = self.CheckHttp(type, url, ping, pong)
                self.AddPoint(self.endpoint, metric, self.step, value, self.counterType, self.tags)
            elif protocol == 'tcp':
                value = self.CheckTcp(type, ip, port, ping, pong)
            elif protocol == 'udp':
                value = self.CheckUdp(type, ip, port, ping, pong)
            else:
                return -1
        self.PushData()
    
    # function to check HTTP status.
    def CheckHttp(self, type, url, method="", ping="", pong=""):
        try:
            if type == "status":
                response = requests.request(method, url)
                code = response.status_code
                if code == 200:
                    return 0
                else:
                    return 1
            elif type == "semanteme":
                response = requests.request(method, url)
                content = response.text
                if pong in content:
                    return 0
                else:
                    return 1
            else:
                return -1
        except:
            return -1

    # TODO:: function to check TCP port status.
    def CheckTcp(self, type, ip, port, ping="", pong=""):pass

    # TODO:: function to check UDP port status.
    def CheckUdp(self, type, ip, port, ping="", pong=""):pass

    def AddPoint(self, endpoint, metric, step, value, counterType, tags):
        point = {
                    'endpoint': endpoint,
                    'metric': metric,
                    'timestamp': int(time.time()),
                    'step': step,
                    'value': value,
                    'counterType': counterType,
                    'tags': tags
                }
        self.data.append(point)

    def PushData(self):
        print(json.dumps(self.data))

# start to check
path = os.path.split(os.path.realpath(__file__))[0]
cfg = path+"/request.json"
patrol = RequestPatrol(cfg)
patrol.HandleRequest();
