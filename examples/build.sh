#!/bin/bash

export GOBIN=$PWD/bin
echo "GOBIN:"$GOBIN
go install -gcflags="all=-N -l"  ./...
