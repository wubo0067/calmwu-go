#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../indexsvr_main index --ip=$ip --port=6005 --conf=../../conf/dev-8889/config.json --cport=605 --logpath=../../log &