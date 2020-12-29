#!/bin/bash
rm -rf doyores doyoreq
go build -o doyoreq -tags static doyoreq.go proto.go
go build -o doyores -tags static doyores.go proto.go
