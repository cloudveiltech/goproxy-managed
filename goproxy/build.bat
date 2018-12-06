set GOARCH=amd64
set GOPROXY_BIN=bin\x64
set GOPROXY_GOROOT=C:\Go
set GOPROXY_PATH=C:\Go\bin
set GOPROXY_GOPATH=%UserProfile%\go
set CGO_LDFLAGS=
set CC=C:\msys64\mingw64\bin\gcc
set CGO_ENABLED=1
call build-internal.bat

set GOARCH=386
set GOPROXY_BIN=bin\x86
REM set GOPROXY_GOROOT=C:\go32
REM set GOPROXY_PATH=C:\go32\bin
REM set GOPROXY_GOPATH=%UserProfile%\go32
REM set CGO_LDFLAGS=-LC:\MinGW\lib -LC:\MinGW\lib\gcc\mingw32\4.8.1
set CC=C:\mingw\bin\gcc
call build-internal.bat
