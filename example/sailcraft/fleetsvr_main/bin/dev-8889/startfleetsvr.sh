> nohup.out
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../fleetsvr_main fleetsvr --ip=$ip --port=3002 --conf=../../conf/ --sysconf=../../../sysconf/dev-8889/ --cport=202 --logpath=../../log & echo $! > ./pid