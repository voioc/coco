#!/bin/sh
cd /www/wwwroot/lemon
/usr/bin/git pull
/bin/sh build.sh

if [ ! -n "$1" ] ;then
    IP_LIST=(`cat /shell/deploy/iphosts`)
else
    IP_LIST=(`echo $1 | tr ',' ' '`)
fi

file_name=/www/wwwroot/lemon/bin/lemon

#for i in ${IP_LIST[*]}
#do
    echo -e "== Update the service == \n"
    #CMD="rsync --include=bin/ --include=config/ --exclude=/* -auvz --delete-after $SRCDIR/* $i::$MOD/shanyin_push/"
    #echo -e "\n == delete the bak file ==\n"
    #if [ ! -d "$file_name"]; then
        /bin/sh /shell/deploy/cmd.sh "$1" "rm -rf /www/wwwroot/lemon/bin/lemon.bak"
        /bin/sh /shell/deploy/cmd.sh "$1" "mv /www/wwwroot/lemon/bin/lemon /www/wwwroot/lemon/bin/lemon.bak" "Back the file... [OK]"
    #fi
    for i in ${IP_LIST[*]}
    do 
	echo -ne "== $i msg:Upload file..."
    	scp -r /www/wwwroot/lemon/bin root@$i:/www/wwwroot/lemon/
    done
    /bin/sh /shell/deploy/cmd.sh "$1" "/etc/init.d/lemon restart" 

    echo -e "[success]\n"
#done

#/bin/sh /letv/sh/rsync_push_push.sh
#/bin/sh /letv/sh/batch_cmd.sh "/etc/init.d/shanyin-push-rpc restart"
