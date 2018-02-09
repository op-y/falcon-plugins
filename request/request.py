#! /usr/bin/env python
#-*- coding:utf-8 -*-
#
# Author: yezhiqin@hualala.com
# Date  : 2016/11/11 create
# Date  : 2017/05/29 modify

import commands
import json
import os
import requests
import time

class Checker:
    def __init__(self, cfg):
        with open(cfg) as file:
            config = json.load(file)
            self.endpoint = os.uname()[1]
            self.step = config['step']
            self.counterType = config['counterType']
            self.tags = config['tags']
            self.request = config['request']
            self.data = []

    # function to dispatch request items.
    def handleRequest(self):
        for item in self.request:
            metric    = item['metric']
            protocol  = item['protocol']
            judgement = item['judgement']
            url       = item.get('url','')
            method    = item.get('method','GET')
            ip        = item.get('ip','127.0.0.1')
            port      = item.get('port','')
            keyword   = item.get('keyword','')
            if protocol == 'http':
                value = self.checkHTTP(metric, judgement, url, method, keyword)
                self.addPoint(self.endpoint, metric, self.step, value, self.counterType, self.tags)
                print "=========="
            elif protocol == 'grpc':
                value = self.checkGRPC(metric, judgement, ip, port, method, keyword)
                self.addPoint(self.endpoint, metric, self.step, value, self.counterType, self.tags)
                print "=========="
            else:
                print "%s Unknown Protocol!" % metric
                print "=========="
        self.push2falcon()
        print "~~~~~End~~~~~"
    
    # function to check HTTP status.
    def checkHTTP(self, metric, judgement, url, method="", keyword=""):
        try:
            if judgement == "status":
                response = requests.request(method, url, timeout=5)
                code = response.status_code
                if code == 200 or code == 503 or code == 404:
                    print "%s is OK" % metric
                    print "    The status code is %d" % code
                    return 0
                else:
                    print "%s is abnormal!" % metric
                    print "    The status code is %d" % code
                    return 1
            elif judgement == "semanteme":
                response = requests.request(method, url, timeout=5)
                code = response.status_code
                if code == 200 or code == 503 or code == 404:
                    print "%s is OK." % metric
                    print "    The status code is %d" % code
                else:
                    print "%s is abnormal!" % metric
                    print "    The status code is %d" % code
                    return 1
                content = response.text
                if keyword not in content:
                    print "%s is OK." % metric
                    print "    The response text is: %s" % content
                    return 0
                else:
                    print "%s is abnormal!" % metric
                    print "    The response text is: %s" % content
                    return 2
            else:
                print "Unknown Judgement!"
                return -1
        except Exception,e:
            print "Request Error!"
            print e
            return -1

    # function to check GRPC status.
    def checkGRPC(self, metric, judgement, ip, port, method="", keyword=""):
        try:
            if judgement == "status":
                cmd = "./HealthCheck %s:%s" % (ip, port)
                status, output = commands.getstatusoutput(cmd)
                if status == 0:
                    print "%s is OK" % metric
                    print "    The status code is %d" % status
                    return 0
                else:
                    print "%s is abnormal!" % metric
                    print "    The status code is %d" % status
                    return 1
            elif judgement == "semanteme":
                pass
            else:
                print "Unknown Judgement!"
                return -1
        except Exception,e:
            print "Request Error!"
            print e
            return -1

    def addPoint(self, endpoint, metric, step, value, counterType, tags):
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

    def push2falcon(self):
        print "Push data to falcon: %s" % json.dumps(self.data)
        response = requests.post("http://127.0.0.1:1988/v1/push", json.dumps(self.data), timeout=5)
        code = response.status_code
        text = response.text
        print "Push data to falcon, status code is %d and response text is %s" % (code,text)

    def send2Wechat(self, tos, content):
        url="http://127.0.0.1:9090/msg"
        data={'tos':tos,'content':content}
        response = requests.post(url, data=data)
        code = response.status_code
        text = response.text
        print "Send message to wechat, status code is %d and response text is %s" % (code,text)

# start to check
path = os.path.split(os.path.realpath(__file__))[0]
cfg = path+"/request.json"
checker = Checker(cfg)
checker.handleRequest();
