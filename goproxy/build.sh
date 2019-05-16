#!/bin/bash

export GOPROXY_BIN=$PWD/bin/x64

export PROXY_OUTPUT_FILE=libproxy.dylib

export GOOS=darwin
export GOARCH=amd64
export CC=gcc
export CGO_ENABLED=1
bash build-internal.sh

export PROXY_OUTPUT_FILE=proxy.dll
export GOOS=windows
export CC=x86_64-w64-mingw32-gcc
bash build-internal.sh

export GOPROXY_BIN=$PWD/bin/x86
export GOARCH=386
export CC=i686-w64-mingw32-gcc
bash build-internal.sh

