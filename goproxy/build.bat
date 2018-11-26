go get gopkg.in/elazarl/goproxy.v1

mkdir bin
mkdir bin/x86
mkdir bin/x64

set GOARCH=amd64
go build -ldflags "-s -w" --buildmode=c-shared -o bin/x64/proxy.dll

set GOARCH=386
go build -ldflags "-s -w" --buildmode=c-shared -o bin/x86/proxy.dll

