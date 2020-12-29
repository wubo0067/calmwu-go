#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
../logsvr_main --ip=$ip --port=6000 --storagepath=../../logstorage --logpath=../../log --consul=$ip --cport=6100