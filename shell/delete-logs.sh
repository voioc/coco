#!/bin/bash

if [ -f /tmp/delete-log ];then
    rm /tmp/delete-log
    # echo 111;
    exit 0
else
    touch /tmp/delete-log
    find /www/server/total/logs -name "*.log" -mtime +0 -exec rm -rf {} \;
    echo > /www/server/nginx/on;
    echo > /log/kunyu77.com.log;
    echo > /log/lemon_access.log;
    # echo 222;
fi
