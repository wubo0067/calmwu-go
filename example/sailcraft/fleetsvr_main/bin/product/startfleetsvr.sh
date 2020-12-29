> nohup.out
ip=`ifconfig eth0|sed -n 2p|awk  '{ print $2 }'|tr -d 'addr:'`
nohup ../fleetsvr_main fleetsvr --ip=$ip --port=3000 --conf=../../conf/ --sysconf=../../../sysconf/product/ --cport=3100 --logpath=../../log --consul=10.161.118.71 & echo $! > ./pid