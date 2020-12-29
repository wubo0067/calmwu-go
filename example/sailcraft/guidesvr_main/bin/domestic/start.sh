#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../guidesvr_main guide --ip=$ip --port=8000 --conf=../../conf/domestic/config.json --cport=8100 --logpath=../../log --consul=$ip &