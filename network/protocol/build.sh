#!/bin/bash

protoc --proto_path=./ --proto_path=/usr/local/include --go_out=./ --cpp_out=./ proto_hs.proto