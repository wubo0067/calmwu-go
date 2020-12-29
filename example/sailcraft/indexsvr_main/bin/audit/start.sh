#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
../indexsvr_main index --ip=$ip --port=5000 --conf=../../conf/audit/config.json --cport=5100 --logpath=../../log --consul=$ip