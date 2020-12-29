#!/bin/bash

echo "**********now pack: $1**********"

if [[ "$1" == "stage" || "$1" == "test" || "$1" == "product" ]]
then

    if [[ ! -d "../bin/$1" || ! -d "../conf/$1" ]]
    then
        echo "../bin/$1 or ../conf/$1 directory is not exist!"
    else
        if [[ -f "indexsvr_$1.tgz" ]]; then
            DATE=`date '+%Y-%m-%d %H:%M:%S'`
            mv "indexsvr_$1.tgz" "indexsvr_$1_$DATE.tgz"
        fi

        rm -rf log conf bin indexsvr_$1.tgz
        mkdir log
        mkdir conf
        mkdir bin

        cd ..
        make clean all
        cd -

        cp ../bin/indexsvr_main ./bin
        cp -rf ../bin/$1 ./bin
        cp -rf ../conf/$1 ./conf
        dos2unix ./bin/$1/*.sh
        chmod +x ./bin/* ./bin/$1/*
        tar -czf indexsvr_$1.tgz log conf bin

        echo "**********now pack: $1 completed**********"
    fi
else
echo "pack $1 is unknown!"
fi