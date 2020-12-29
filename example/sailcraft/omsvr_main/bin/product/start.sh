#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../omsvr_main omsvr --ip=$ip --port=2000 --conf=../../conf/product/config.json --cport=2100 --logpath=../../log --consul=$ip &