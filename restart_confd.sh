#!/bin/bash

# 线上环境重启confd脚本
echo "Restart Confd..."
ps aux | grep "confd/bin/confd" | grep -v 'grep' | awk '{print "kill -9 "$2}' | sh 
/usr/local/go/src/github.com/mafengwo/confd/bin/confd -backend etcd -interval 10 -node http://127.0.0.1:10001 -redisqueue 192.168.3.40:6379 >> /var/log/confd.log 2>&1 &

