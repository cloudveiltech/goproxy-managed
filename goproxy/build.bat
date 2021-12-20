set PROXY_OUTPUT_FILE=proxy.dll
set OLD_PATH=%PATH%

set BASE_PATH=%SystemRoot%;%SystemRoot%\System32

set PATH=%OLD_PATH%;%BASE_PATH%;C:\Go\bin;C:\Program Files\Git\bin

set GOARCH=amd64
set GOPROXY_BIN=bin\x64
set GOPROXY_GOROOT=C:\Go
set GOPROXY_PATH=C:\Go\bin
set GOPROXY_GOPATH=%UserProfile%\go
set CGO_LDFLAGS=
set CGO_ENABLED=1
call build-internal.bat

REM TODO Add Darwin building

set PATH=%OLD_PATH%;%BASE_PATH%;C:\Go\bin;C:\Program Files\Git\bin
set GOARCH=386
set GOPROXY_BIN=bin\x86
set GOPROXY_GOPATH=%UserProfile%\go32
call build-internal.bat

set PATH=%OLD_PATH%

