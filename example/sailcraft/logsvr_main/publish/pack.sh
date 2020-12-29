#!/bin/bash

rm *.tgz
rm -rf log logstorage bin
mkdir log
mkdir logstorage
mkdir bin
cp ../bin/* ./bin
chmod +x ./bin/*
tar -czf logsvr.tgz log logstorage bin

