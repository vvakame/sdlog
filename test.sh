#!/bin/sh -eux

cd `dirname $0`

targets=`find . -type f \( -name '*.go' -and -not -iwholename '*vendor*' -and -not -iwholename '*testdata*' \)`
packages=`go list ./...`

# Apply tools
export PATH=$(pwd)/build-cmd:$PATH
which goimports golint
goimports -w $targets
for package in $packages
do
    go vet $package
done
golint -set_exit_status -min_confidence 0.6 $packages

go test $packages -count 1 -p 1 -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt $@
