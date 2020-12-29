#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
../financesvr_main finance --ip=$ip --port=4000 --conf=../../conf/audit/config.json --cport=4100 --logpath=../../log --consul=$ip