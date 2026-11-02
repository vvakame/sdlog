#!/bin/sh -eux

cd `dirname $0`

go tool goimports -w .
go vet ./...
# fails when a modernize fix is available
go fix -diff ./...

go test ./... -count 1 -p 1 -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt $@
