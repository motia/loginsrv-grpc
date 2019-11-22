#!/usr/bin/env bash

set -e
set -x

protoc -I ./ ./loginsrv.proto --go_out=plugins=grpc:.

echo "Items generrated"