#!/bin/bash

# 开发环境环境重启confd脚本
echo "Restart Confd..."
ps aux | grep "confd/bin/confd" | grep -v 'grep' | awk '{print "kill -9 "$2}' | sh
/usr/local/go/src/github.com/mafengwo/confd/bin/confd -backend etcd -interval 2 -node http://127.0.0.1:30100 >> /var/log/confd.log 2>&1 &