#!/bin/bash
#you have to install pcstat
if [ ! -f /data0/brokerproxy/pcstat ]
then
    echo "You haven't installed pcstat yet"
    echo "run \"go get github.com/tobert/pcstat\" to install"
    exit
fi
#find the top 10 processs' cache file
ps -e -o pid,rss|sort -nk2 -r|head -10 |awk '{print $1}'>/tmp/cache.pids
#find all the processs' cache file
#ps -e -o pid>/tmp/cache.pids
if [ -f /tmp/cache.files ]
then
    echo "the cache.files is exist, removing now "
    rm -f /tmp/cache.files
fi
while read line
do
    lsof -p $line 2>/dev/null|awk '{print $9}' >>/tmp/cache.files 
done</tmp/cache.pids
if [ -f /tmp/cache.pcstat ]
then
    echo "the cache.pcstat is exist, removing now"
    rm -f /tmp/cache.pcstat
fi
for i in `cat /tmp/cache.files`
do
    if [ -f $i ]
    then
        echo $i >>/tmp/cache.pcstat
    fi
done
/data0/brokerproxy/pcstat  `cat /tmp/cache.pcstat`
rm -f /tmp/cache.{pids,files,pcstat}
