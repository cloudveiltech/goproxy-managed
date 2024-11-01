set PROXY_OUTPUT_FILE=proxy.dll
set OLD_PATH=%PATH%

set BASE_PATH=%SystemRoot%;%SystemRoot%\System32

set PATH=%OLD_PATH%;%BASE_PATH%;C:\Program Files\Go;C:\Program Files\Git\bin

set GOARCH=amd64
set GOPROXY_BIN=bin\x64
set GOPROXY_GOROOT=C:\Program Files\Go
set GOPROXY_PATH=C:\Program Files\Go\bin
set GOPROXY_GOPATH=%UserProfile%\go
set CGO_LDFLAGS=
set CGO_ENABLED=1
set CC=d:\llvm-mingw-20240404-msvcrt-i686\bin\x86_64-w64-mingw32-gcc.exe
set CXX=d:\llvm-mingw-20240404-msvcrt-i686\bin\x86_64-w64-mingw32-g++.exe
call build-internal.bat

REM TODO Add Darwin building

set PATH=%OLD_PATH%;%BASE_PATH%;C:\Program Files\Go;C:\Program Files\Git\bin
set GOARCH=386
set GOPROXY_BIN=bin\x86
set GOPROXY_GOPATH=%UserProfile%\go32
call build-internal.bat

set GOARCH=arm64
set CC=d:\llvm-mingw-20240404-msvcrt-i686\bin\aarch64-w64-mingw32-gcc.exe
set CXX=d:\llvm-mingw-20240404-msvcrt-i686\bin\aarch64-w64-mingw32-g++.exe
set GOPROXY_BIN=bin\arm64
set GOPROXY_GOPATH=%UserProfile%\go
call build-internal.bat

set PATH=%OLD_PATH%

