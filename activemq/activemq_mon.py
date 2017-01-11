#! /usr/bin/env python
#-*- coding:utf-8 -*-
#
# Author: yezhiqin@hualala.com
# Date  : 2017/01/10

import os
import requests
import xml.dom.minidom
import time
import json

class ActivemqMonitor:
    def __init__(self):
        self.data = []
        self.endpoint = os.uname()[1]
        self.metric = "activemq.consumerCount"
        self.step = 60
        self.counterType = "GAUGE"
        self.reqURL = "http://127.0.0.1:8161/admin/xml/queues.jsp"
        self.xmlString = ""
        self.xmlFile = "queues.xml"
        self.agentURL = "http://127.0.0.1:1988/v1/push"

    # function to get xml.
    def getXML(self, url, xmlFile, method="get"):
        try:
            response = requests.request(method, url)
            code = response.status_code
            if code == 200:
                self.xmlString = response.text
                f = open(xmlFile, 'w')
                f.write(response.text)
                f.close()
                return 0
            else:
                return 1
        except:
            return -1

    # TODO:: function to parse queues xml file.
    def processXML(self, xmlString):
        domTree = xml.dom.minidom.parseString(xmlString)
        queues = domTree.getElementsByTagName( "queue" )
        for queue in queues:
            queueName = queue.getAttribute('name')
            consumerCount = int(queue.getElementsByTagName("stats")[0].getAttribute('consumerCount'))
            self.addPoint(queueName, consumerCount)

    def addPoint(self, queueName, consumerCount):
        point = {
                    'endpoint': self.endpoint,
                    'metric': self.metric,
                    'timestamp': int(time.time()),
                    'step': self.step,
                    'value': consumerCount,
                    'counterType': self.counterType,
                    'tags': "queuename=%s" % queueName
                }
        self.data.append(point)

    def push(self):
        response = requests.post(self.agentURL, json.dumps(self.data))
        print(json.dumps(self.data))

    def run(self):
        self.getXML(self.reqURL, self.xmlFile)
        self.processXML(self.xmlString)
        self.push()

# start to check
if __name__ == '__main__':
    monitor = ActivemqMonitor()
    monitor.run();
