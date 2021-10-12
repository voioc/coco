#!/bin/sh
#获取单个进程的状态，参数1：程序名，参数2：重启脚本
function getSingleProcessStats()
{
    LOG_TIME="`/bin/date +'%Y-%m-%d %H:%M:%S'`"
    CMD=$(/bin/ps -ef | /bin/grep $1 | /bin/grep -v 'grep'|/bin/grep -v $0|/usr/bin/wc -l)
    if [ $CMD -ne 0 ];then
        echo "[$LOG_TIME][$1] OK"
    else
        echo "[$LOG_TIME][$1] ERROR"
        RE_CMD=$(${@:2})
        if [ $? -eq 0 ];then
            echo "[$LOG_TIME][$1] RESTART SERVICE OK"
        else
            echo "[$LOG_TIME][$1] RESTART SERVICE FAILED"
        fi
    fi
}
if [ $# -ne 2 ]; then
    printf "\033[1;33musage: $0 process_cmdline restart_script\033[m\n"
    printf "\033[1;33mexample: `pwd`/proc_monitor.sh \"/usr/sbin/rinetd\" \"/usr/sbin/rinetd\"\033[m\n"
    printf "\033[1;33mplease install process_monitor.sh into crontab by \"* * * * *\"\033[m\n"
    exit 1
else
    getSingleProcessStats $1 $2
fi
