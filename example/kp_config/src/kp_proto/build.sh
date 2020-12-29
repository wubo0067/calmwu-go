#!/bin/bash

protoc --proto_path=./ --proto_path=/usr/local/include --python_out=./ --go_out=./ kp_proto.proto