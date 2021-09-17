#!/bin/sh

# IP_LIST=(
# `cat /shell/deploy/iphosts`
# )

if [ ! -n "$1" ] ;then
    IP_LIST=(`cat /shell/deploy/iphosts`)
else
   IP_LIST=(`echo $1 | tr ',' ' '`)
fi

# echo ${IP_LIST[@]}

CMD=$2
MSG=$3
for i in ${IP_LIST[*]}
do
	if [ -n "$MSG" ]; then
	    echo -ne "== $i msg:$2"
	fi
        ssh  root@$i "$CMD"
	if [ -n "$MSG" ]; then
            echo -e "         [OK]"
        fi
       sleep 1
done
