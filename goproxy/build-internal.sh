
# this file is for setting up individual builds for x64 vs x86.
# specify make directories in build.bat
# Needs:
#   $GOPROXY_BIN
#   $GOPROXY_GOROOT
#   $GOPROXY_PATH

export BUILDMODE=c-shared

if [ "$PROXY_OUTPUT_FILE" == "" ]; then
	export PROXY_OUTPUT_FILE=libproxy.so
fi;

if [ "$PROXY_OUTPUT_FILE" == "libproxy.so" ] || [ "$PROXY_OUTPUT_FILE" == "libproxy.dylib" ] || [ "$PROXY_OUTPUT_FILE" == "proxy.dll" ]; then
	export BUILDMODE=c-shared
else
	export BUILDMODE=exe
fi;

OLD_GOROOT=$GOROOT
OLD_GOPATH=$GOPATH

go get -d ./...

mkdir -p $GOPROXY_BIN

go build -ldflags "-s -w" --buildmode=$BUILDMODE -o $GOPROXY_BIN/$PROXY_OUTPUT_FILE
