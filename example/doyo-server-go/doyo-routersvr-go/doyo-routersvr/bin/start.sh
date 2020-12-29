#!/bin/bash
sleep 6s
./doyo-routersvr --id=1 --topic=DoyoRouterSvr --conf=../conf/conf.json --logpath=../log --consulhealthcheckaddr=192.168.68.228:7002
