> nohup.out
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../fleetsvr_main fleetsvr --ip=$ip --port=4002 --conf=../../conf/ --sysconf=../../../sysconf/stage/ --cport=202 --logpath=../../log --consul=10.10.81.214 & echo $! > ./pid