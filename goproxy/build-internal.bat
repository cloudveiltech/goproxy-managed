@echo off

REM this file is for setting up individual builds for x64 vs x86.
REM specify make directories in build.bat
REM Needs:
REM   %GOPROXY_BIN%
REM   %GOPROXY_GOROOT%
REM   %GOPROXY_PATH%

set OLD_GOROOT=%GOROOT%
set OLD_GOPATH=%GOPATH%

set GOROOT=%GOPROXY_GOROOT%
set GOPATH=%GOPROXY_GOPATH%

%GOPROXY_PATH%\go get -d .\...

mkdir %GOPROXY_BIN%

%GOPROXY_PATH%\go build -x -ldflags "-s -w" --buildmode=c-shared -o %GOPROXY_BIN%\proxy.dll

set GOROOT=%OLD_GOROOT%
set GOPATH=%OLD_GOPATH%
