#!/bin/bash -e

protoc --go_opt=paths=source_relative \
        --go_out=. \
       --go-grpc_out=. \
       --go-grpc_opt=paths=source_relative \
       helloworld.proto
