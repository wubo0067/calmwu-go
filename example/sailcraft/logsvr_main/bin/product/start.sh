#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../logsvr_main --ip=$ip --port=6000 --storagepath=../../logstorage --logpath=../../log --consul=10.161.118.71 --cport=6100 &