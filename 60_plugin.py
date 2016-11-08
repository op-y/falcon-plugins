#! /usr/bin/env python
#-*- coding:utf-8 -*-

import json
import time

data = [
        {
            'metric': 'plugins.1.test',
            'endpoint': 'jieqianhua-host3',
            'timestamp': int(time.time()),
            'step': 60,
            'value': 1,
            'counterType': 'GAUGE',
            'tags': 'company=duolaidian,department=dev,product=mendianbao'
            },
        {
            'metric': 'plugins.0.test',
            'endpoint': 'jieqianhua-host3',
            'timestamp': int(time.time()),
            'step': 60,
            'value': 0,
            'counterType': 'GAUGE',
            'tags': 'company=duolaidian,department=dev,product=mendianbao'
            }
        ]

print(json.dumps(data))
