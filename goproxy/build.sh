#!/bin/bash

export GOPROXY_BIN=$PWD/bin/x64

export GOOS=darwin
export GOARCH=amd64
export CC=gcc
export CGO_ENABLED=1

echo "Building x64..."
go build -ldflags "-s -w" --buildmode=c-archive -o proxy-x64.a

echo "Building arm..."
export GOARCH=arm64
go build -ldflags "-s -w" --buildmode=c-archive -o proxy-arm64.a

lipo -create proxy-x64.a proxy-arm64.a -output proxy.a
lipo -info proxy.a

echo "done"