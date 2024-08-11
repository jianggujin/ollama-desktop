#!/bin/bash

set -e

echo "[ BUILD RELEASE ]"
BuildVersion="v0.0.1"
BuildTime=$(date +"%Y-%m-%d %H:%M:%S")
AllPlatform="darwin/amd64,darwin/arm64,windows/amd64,linux/amd64,linux/arm64"

# -ldflag 参数
GOLDFLAGS="-s -w -X 'ollama-desktop/internal/config.BuildTime=$BuildTime'"
GOLDFLAGS+=" -X 'ollama-desktop/internal/config.BuildVersion=$BuildVersion'"

# build the current platform
export GOOS=$(go env get GOOS | sed ':a;N;$!ba;s/^\n*//;s/\n*$//')
export GOARCH=$(go env get GOARCH | sed ':a;N;$!ba;s/^\n*//;s/\n*$//')
echo "[ DIST CURRENT PLATFORM GOOS=$GOOS GOARCH=$GOARCH ]"
if [ "$1" == "nsis" ]; then
  wails build -clean -ldflags "$GOLDFLAGS" -m -nsis -trimpath
else
  wails build -clean -ldflags "$GOLDFLAGS" -m -trimpath
fi
echo "[ BUILD SUCCESS GOOS=$GOOS GOARCH=$GOARCH ]"

