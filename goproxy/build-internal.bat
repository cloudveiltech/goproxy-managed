
@echo off

REM this file is for setting up individual builds for x64 vs x86.
REM specify make directories in build.bat
REM Needs:
REM   %GOPROXY_BIN%
REM   %GOPROXY_GOROOT%
REM   %GOPROXY_PATH%

%GOPROXY_PATH%\go get gopkg.in/elazarl/goproxy.v1

mkdir %GOPROXY_BIN%

%GOPROXY_GOROOT%\go build -ldflags "-s -w" --buildmode=c-shared -o bin\x64\proxy.dll

set OLD_GOROOT=%GOROOT%
set OLD_GOPATH=%GOPATH%