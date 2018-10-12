#!/bin/bash

# installation/update of mage tool: https://github.com/magefile/mage
# required for quickly generation of protobuf files dynamically

go get -u -d github.com/magefile/mage
cd $GOPATH/src/github.com/magefile/mage
go run bootstrap.go