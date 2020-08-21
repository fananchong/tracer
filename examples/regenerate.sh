#!/bin/bash

set -eu -o pipefail

WORKDIR=$PWD
export GOBIN=${WORKDIR}/bin
export PATH=${GOBIN}:${PATH}
mkdir -p ${GOBIN}

go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
wget -P ${WORKDIR}/bin https://github.com/protocolbuffers/protobuf/releases/download/v3.12.4/protoc-3.12.4-linux-x86_64.zip
unzip -q -o ${WORKDIR}/bin/protoc-3.12.4-linux-x86_64.zip -d ${WORKDIR}/bin/
cp -f ${WORKDIR}/bin/bin/protoc ${WORKDIR}/bin/
rm -rf ${WORKDIR}/bin/bin/ ${WORKDIR}/bin/include ${WORKDIR}/bin/protoc-3.12.4-linux-x86_64.zip ${WORKDIR}/bin/readme.txt

protoc --go_out=plugins=grpc:./proto -I./proto test.proto

go mod tidy

echo "done"
