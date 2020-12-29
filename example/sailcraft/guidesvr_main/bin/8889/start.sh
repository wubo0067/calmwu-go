#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../guidesvr_main guide --ip=$ip --port=8009 --conf=../../conf/8889/config.json --cport=8109 --logpath=../../log &