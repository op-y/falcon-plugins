#! /usr/bin/env python
#-*- coding:utf-8 -*-

import json
import os
import pickle
import re
import time

class LogPatrol:
    def __init__(self, config_file):
        self.config_file = config_file
        with open(self.config_file) as file:
            config = json.load(file)
            self.endpoint = config['endpoint']
            self.step = config['step']
            self.counterType = config['counterType']
            self.tags = config['tags']
            self.logs = config['logs']

    def HandleLog(self):
        for log in self.logs:
            file = log['file']
            split_ok = self.SplitLog(file)
            if not split_ok:
                continue
            items = log['items']
            for item in items:
                name = item['name']
                type = item['type']
                keyword = item['keyword']
                self.AnalyzeItem(file, name, type, keyword)

    def SplitLog(self, log):
        meta_prefix = 'meta'
        temp_prefix = 'temp'
        log_meta = meta_prefix+"."+log
        log_temp = temp_prefix+"."+log
    
        if not os.path.exists(log):
            return False
        if not os.path.isfile(log):
            return False
    
        if not os.path.exists(log_meta):
            # calculate log file line number.
            sys_cmd = "wc -l %s | awk '{print $1}'" % (log)
            sys_handler = os.popen(sys_cmd)
            total_num = sys_handler.read()
            total = int(total_num)
            processed = total
            # initialize log meta file.
            meta = {'total':total, 'processed':processed}
            meta_handler = open(log_meta, 'w')
            pickle.dump(meta, meta_handler) 
            meta_handler.close()
            return False
        else:
            # load info from log meta file.
            meta_handler = open(log_meta, 'r')
            meta = pickle.load(meta_handler)
            meta_handler.close()
            # calculate log file line number again. 
            sys_cmd = "wc -l %s | awk '{print $1}'" % (log)
            sys_handler = os.popen(sys_cmd)
            new_total_num = sys_handler.read()
            total = int(new_total_num)
            if total > meta['total']:
                # calculate the addition line number.
                diff = total - meta['total']
                # truncate the addition log to temp log file.
                sys_cmd = "tail -n %d %s > %s" % (diff, log, log_temp)
                sys_handler = os.system(sys_cmd)
                # update the info in log meta file.
                meta = {'total':total, 'processed':total}
                meta_handler = open(log_meta, 'w')
                pickle.dump(meta, meta_handler) 
                meta_handler.close()
                return True
            else:
                # log rotated, update the info in log meta file.
                meta = {'total':total, 'processed':total}
                meta_handler = open(log_meta, 'w')
                pickle.dump(meta, meta_handler) 
                meta_handler.close()
                return False

    def AnalyzeItem(self, log, name, type, keyword):
        temp_prefix = 'temp'
        log_temp = temp_prefix+"."+log
        if not os.path.exists(log_temp):
            return
        if not os.path.isfile(log_temp):
            return

        if type == 'regular':
            if os.path.getsize(log_temp) == 0:
                self.PushPoint(self.endpoint, name+".cnt", self.step, 0, self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".cps", self.step, 0, self.counterType, self.tags)
            else:
                result = self.MatchRegular(log_temp, keyword, self.step)
                self.PushPoint(self.endpoint, name+".cnt", self.step, result['cnt'], self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".cps", self.step, result['cps'], self.counterType, self.tags)
        elif type == 'data':
            if os.path.getsize(log_temp) == 0:
                self.PushPoint(self.endpoint, name+".sum", self.step, 0, self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".max", self.step, 0, self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".min", self.step, 0, self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".avg", self.step, 0, self.counterType, self.tags)
            else:
                result = self.MatchData(log_temp, keyword)
                self.PushPoint(self.endpoint, name+".sum", self.step, result['sum'], self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".max", self.step, result['max'], self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".min", self.step, result['min'], self.counterType, self.tags)
                self.PushPoint(self.endpoint, name+".avg", self.step, result['avg'], self.counterType, self.tags)
        else:
            return

    def MatchRegular(self, file, keyword, step):
        cnt = 0
        pattern = re.compile(keyword);
        for line in open(file):
            search_obj = re.search(pattern, line)
            if search_obj:
                cnt += 1
        cps = float(cnt) / step
        return {'cnt':cnt, 'cps':cps}

    def MatchData(self, file, keyword):
        cnt = 0
        sum = 0
        max = 0
        min = float("inf")
        avg = 0
        pattern = re.compile(keyword)
        for line in open(file):
            search_obj = re.search(pattern, line)
            if search_obj:
                cnt += 1
                value = float(search_obj.group(1))
                sum += value
                if value > max:
                    max = value
                if value < min:
                    min = value
        if min == float("inf"):
            min = 0
        if cnt != 0:
            avg = sum / cnt
        return {'sum':sum, 'max':max, 'min':min, 'avg':avg}

    def PushPoint(self, endpoint, metric, step, value, counterType, tags):
        point = [{
                    'endpoint': endpoint,
                    'metric': metric,
                    'timestamp': int(time.time()),
                    'step': step,
                    'value': value,
                    'counterType': counterType,
                    'tags': tags
                }]
        print(json.dumps(point))

path = os.path.split(os.path.realpath(__file__))[0]
cfg = path+"/log_patrol.json"
patrol = LogPatrol(cfg)
patrol.HandleLog();
