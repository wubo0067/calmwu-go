#!/bin/bash

ip="10.10.81.214:4000"
version=""
zoneid=""

while getopts 'v:z:' OPT; do
    case $OPT in
        v)
            version=$OPTARG;;
        z)
            zoneid=$OPTARG;;
        ?)
            echo "Usage: `basename $0` -v 1 -z 1"
            exit
    esac
done

if [[ "$ip" = "" ]]
then
    echo "$0 -v 1 -z 1"
    exit
fi

if [[ "$zoneid" = "" ]]
then
    echo "$0 -v 1 -z 1"
    exit
fi

./shopConfig.exe --type=resource --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=cardpack --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=refresh --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=recharge --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=signin --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=vip --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=newplayerbenefit --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=supergift --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=mission --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=exchange --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=cdkey --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip

./shopConfig.exe --type=firstrecharge --configpath=../doc/Shop --zoneid=$zoneid --version=$version --svrip=$ip