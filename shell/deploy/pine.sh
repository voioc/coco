#!/bin/sh
cd /www/wwwroot/pine
/usr/bin/git pull
/bin/sh build.sh

if [ ! -n "$1" ] ;then
    IP_LIST=(`cat /shell/deploy/iphosts`)
else
    IP_LIST=(`echo $1 | tr ',' ' '`)
fi

# echo ${IP_LIST[@]}
file_name=/www/wwwroot/pine/bin/pine

echo -e "\n"
echo -e "== Update the service == "
/bin/sh /shell/deploy/cmd.sh "$1"  "rm -rf /www/wwwroot/pine/bin/pine.bak" "Delete backup file"
/bin/sh /shell/deploy/cmd.sh "$1" "mv /www/wwwroot/pine/bin/pine /www/wwwroot/pine/bin/pine.bak" "Backup the file... [OK]"

for i in ${IP_LIST[*]}
do 
    echo -ne "== $i msg:Upload file..."
    scp -r /www/wwwroot/pine/bin root@$i:/www/wwwroot/pine/
    echo -e "\n"
done

# restart service
/bin/sh /shell/deploy/cmd.sh "$1" "/etc/init.d/pine restart" "restart the service"

echo -e "[success]\n"

