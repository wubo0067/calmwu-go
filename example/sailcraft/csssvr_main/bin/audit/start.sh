#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
../csssvr_main store --ip=$ip --port=9000 --conf=../../conf/audit/config.json --cport=9100 --logpath=../../log --consul=$ip