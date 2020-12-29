#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
../guidesvr_main guide --ip=$ip --port=8000 --conf=../../conf/audit/config.json --cport=8100 --logpath=../../log --consul=$ip