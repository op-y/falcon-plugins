#! /usr/bin/env python
#-*- coding:utf-8 -*-

import json
import time
import requests

class RequestPatrol:
    def __init__(self, config_file):
        self.config_file = config_file
        with open(self.config_file) as file:
            config = json.load(file)
            self.agentApi = config['agentApi']
            self.endpoint = config['endpoint']
            self.step = config['step']
            self.counterType = config['counterType']
            self.tags = config['tags']
            self.request = config['request']

    def HandleRequest(self):
        for item in self.request:
            metric = item['metric']
            protocol = item['protocol']
            type = item['type']
            url = item.get('url','')
            ip = item.get('ip','')
            port = item.get('port','')
            ping = item.get('ping','')
            pong = item.get('pong','')
            if protocol == 'http':
                value = self.CheckHttp(type, url, ping, pong)
                self.PushPoint(self.endpoint, metric, self.step, value, self.counterType, self.tags)
            elif protocol == 'tcp':
                value = self.CheckTcp(type, ip, port, ping, pong)
            elif protocol == 'udp':
                value = self.CheckUdp(type, ip, port, ping, pong)
            else:pass
    
    def CheckHttp(self, type, url, ping="", pong=""):
        response = requests.post(url)
        if type == "status":
            code = response.status_code
            if code == 200:
                return 1
            else:
                return 0
        elif type == "semanteme":
            content = response.text
            if pong in content:
                return 1
            else:
                return 0
        else:
            return -1

    def CheckTcp(self, type, ip, port, ping="", pong=""):pass

    def CheckUdp(self, type, ip, port, ping="", pong=""):pass

    def PushPoint(self, endpoint, metric, step, value, counterType, tags):
        point = [{
                    'metric': endpoint,
                    'endpoint': metric,
                    'timestamp': int(time.time()),
                    'step': step,
                    'value': value,
                    'counterType': counterType,
                    'tags': tags
                }]
        r = requests.post(self.agentApi, data=json.dumps(point))

if __name__ == '__main__':
    patrol = RequestPatrol('request_patrol.json')
    patrol.HandleRequest();
