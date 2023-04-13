#!/bin/bash -eux

cd `dirname $0`

go mod download

# build tools
rm -rf build-cmd/
mkdir build-cmd

export GOBIN=`pwd -P`/build-cmd
go install golang.org/x/tools/cmd/goimports
go install golang.org/x/lint/golint
