set GOARCH=amd64
set GOPROXY_BIN=bin\x64
set GOPROXY_GOROOT=C:\Go
set GOPROXY_PATH=C:\Go\bin
set GOPROXY_GOPATH=%UserProfile%\go
set CGO_LDFLAGS=
call build-internal.bat

set GOARCH=386
set GOPROXY_BIN=bin\x86
set GOPROXY_GOROOT=C:\go32
set GOPROXY_PATH=C:\go32\bin
set GOPROXY_GOPATH=%UserProfile%\go32
set CGO_LDFLAGS=-LC:\MinGW\lib -LC:\MinGW\lib\gcc\mingw32\4.8.1
call build-internal.bat
