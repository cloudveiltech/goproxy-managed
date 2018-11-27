set GOARCH=amd64
set GOPROXY_BIN=bin\x64
set GOPROXY_GOROOT=C:\Go
set GOPROXY_PATH=C:\Go\bin
.\build-internal.bat

set GOARCH=386
set GOPROXY_BIN=bin\x86
set GOPROXY_GOROOT=C:\go32
set GOPROXY_PATH=C:\go32\bin
.\build-internal.bat
