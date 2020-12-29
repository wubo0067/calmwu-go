#!/bin/bash
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../indexsvr_main index --ip=$ip --port=5000 --conf=../../conf/product/config.json --cport=5100 --logpath=../../log --consul=10.161.118.71 &