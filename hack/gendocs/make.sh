#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubepack/packserver/hack/gendocs
go run main.go

cd $GOPATH/src/github.com/kubepack/packserver/docs/reference
sed -i 's/######\ Auto\ generated\ by.*//g' *
popd
