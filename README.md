# Setup

Download, unzip and add binary to path:
https://github.com/protocolbuffers/protobuf/releases

go:generate protoc -I ../auth --go_out=plugins=grpc:../auth ../auth/auth.proto


protoc -I auth/ auth/auth.proto --go_out=plugins=grpc:auth
