#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../financesvr_main finance --ip=$ip --port=4000 --conf=../../conf/stage/config.json --cport=4100 --logpath=../../log --consul=10.10.81.214 &