#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../financesvr_main finance --ip=$ip --port=5000 --conf=../../conf/dev-8889/config.json --cport=5100 --logpath=../../log &

