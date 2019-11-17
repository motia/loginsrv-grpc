#!/usr/bin/env bash

set -e
set -x

protoc -I auth/ auth/auth.proto --go_out=plugins=grpc:auth

echo "Items generrated"