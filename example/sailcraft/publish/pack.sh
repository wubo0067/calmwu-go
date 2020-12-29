#!/bin/bash

PackSvrName=$1
PackEnv=$2
SailCraftSvrs=("guidesvr_main" "indexsvr_main" "logsvr_main" "csssvr_main" "financesvr_main" "fleetsvr_main" "omsvr_main")

packSvr(){
    echo -e "************$1 pack start!************"
    mkdir $1/log
    mkdir $1/conf
    mkdir $1/bin

    if [[ $1 == "logsvr_main" ]]; then
        mkdir $1/logstorage
    fi

    #到源代码目录编译
    cd ../$1
    make clean all

    if [[ $? -ne 0 ]]; then
        echo "make $1 failed!"
        exit 0
    fi 
    cd -

    cp ../$1/bin/$1 $1/bin
    cp ../$1/bin/*.sh $1/bin
    cp -rf ../$1/bin/$PackEnv $1/bin
    dos2unix $1/bin/$PackEnv/*.sh
    chmod +x -R $1/bin/*

    if [[ $1 == "fleetsvr_main" ]]; then
        cp -rf ../$1/conf/* $1/conf
    else
        cp -rf ../$1/conf/$PackEnv $1/conf
    fi

    if [[ $1 == "csssvr_main" ]]; then
        cp -rf ../$1/conf/GeoLite2-Country.mmdb $1/conf
        cp -rf ../$1/doc $1
    fi  

    if [[ $1 == "financesvr_main" ]]; then
        cp -rf ../$1/tools $1
        cp -rf ../$1/doc $1
    fi      

    echo -e "************$1 pack over!************\n\n"
}

usage(){
echo "usage, example:
        ./pack.sh all stage(test, product)
        ./pack.sh guidesvr_main stage(test, product)"
}

if [[ "$PackEnv" != "stage" && "$PackEnv" != "test" && "$PackEnv" != "product" && "$PackEnv" != "audit" && "$PackEnv" != "domestic" ]]; then
    usage 
    exit 0
fi

if [[ "$PackSvrName" == "all" ]]
then
    echo "********Pack SailCraft Project, Env:$PackEnv********"
    #打包整个项目
    if [[ -f "sailcraft_all_$PackEnv.tgz" ]]; then
        DATE=`date '+%Y-%m-%d_%H:%M:%S'`
        mv "sailcraft_all_$PackEnv.tgz" "sailcraft_bak_all_${PackEnv}_${DATE}.tgz"
    fi

    if [[ -d "./sysconf" ]]; then
        rm -rf ./sysconf
    fi

    if [[ -d "./3rd" ]]; then
        rm -rf ./3rd
    fi    

    #拷贝系统配置
    mkdir sysconf
    cp -rf ../sysconf/$PackEnv ./sysconf
    cp ../sysconf/SensitiveWordPrecision.txt ./sysconf

    mkdir 3rd
    cp ../3rd/cassandra-3.0.13-1.noarch.rpm ./3rd
    cp ../3rd/createPreserveTables ./3rd
    cp ../3rd/dumpPreserveTables ./3rd

    cp ../startsvrs*.sh ./
    cp ../stopsvrs.sh ./   

    for svr in ${SailCraftSvrs[@]};
    do
        if [[ -d "./$svr" ]]; then
            rm -rf ./$svr
        fi
        mkdir ./$svr
        #打包每个独立的服务
        packSvr $svr
    done

    #打包
    tar -czf sailcraft_all_$PackEnv.tgz ${SailCraftSvrs[*]} sysconf *.sh
    echo "********Pack SailCraft:all, Env:$PackEnv completed!********"
else    
    PackSvrExist=false
    for svr in ${SailCraftSvrs[@]};
    do        
        if [[ "$svr" == "$PackSvrName" ]]
        then
            echo "********Pack SailCraft Svr:$PackSvrName, Env:$PackEnv********"

            if [[ -f "sailcraft_${PackSvrName}_${PackEnv}.tgz" ]]; then
                DATE=`date '+%Y-%m-%d_%H:%M:%S'`
                mv "sailcraft_${PackSvrName}_${PackEnv}.tgz" "sailcraft_bak_$PackSvrName_${PackEnv}_${DATE}.tgz"
            fi 

            PackSvrExist=true
            #拷贝系统配置
            rm -rf sysconf
            mkdir sysconf
            cp -rf ../sysconf/$PackEnv ./sysconf
            cp ../sysconf/SensitiveWordPrecision.txt ./sysconf

            rm -rf ./${svr}
            mkdir ./${svr}
            #打包每个独立的服务
            packSvr ${svr}
            tar -czf sailcraft_${PackSvrName}_${PackEnv}.tgz ${svr} sysconf 
            echo "********Pack SailCraft:${PackSvrName}, Env:${PackEnv} completed!********"                    
        fi
    done
    echo $PackSvrExist
    if !($PackSvrExist)
    then
        echo "Pack $PackSvrName is Invalid"
    fi    
fi


