#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
../fleetsvr_main fleetsvr --ip=$ip --port=3000 --conf=../../conf/ --sysconf=../../../sysconf/audit/ --cport=3100 --logpath=../../log --consul=$ip