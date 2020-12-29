#!/bin/bash

go build doyopush.go

/data/expect_tool/demo_test/scp.sh 47.74.150.95 calm H4zzFLjCaN7JBE30obxc ./doyopush /data/offline_media 22200