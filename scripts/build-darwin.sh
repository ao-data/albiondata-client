#!/usr/bin/env bash

set -eo pipefail

rm -f albiondata-client
rm -f albiondata-client.gz
rm -f update-darwin-amd64.gz
rm -f albiondata-client-amd64-mac.zip

(cd frontend && npm ci && npm run build)

# Native macOS build. GitHub's macos-latest runners are arm64 hardware, so
# amd64 output requires explicit CC/CGO_LDFLAGS arch flags. CGO stays on
# since gopacket links against libpcap.
export CGO_ENABLED=1
export GOARCH=amd64
export CC="clang -arch x86_64"
export CGO_LDFLAGS="-arch x86_64"
go build -ldflags "-s -w -X main.version=$GITHUB_REF_NAME" -o albiondata-client albiondata-client.go

gzip -k9 albiondata-client
mv albiondata-client.gz update-darwin-amd64.gz

# Zipped folder with a run.command file that runs the client under sudo
TEMP="albiondata-client"
ZIPNAME="albiondata-client-amd64-mac.zip"
rm -rfv ./scripts/$TEMP
rm -rfv ./$ZIPNAME
mkdir -v ./scripts/$TEMP
cp -v albiondata-client ./scripts/$TEMP/albiondata-client-executable
cp -v ./scripts/run.command ./scripts/$TEMP/run.command
chmod -v 777 ./scripts/$TEMP/*
(cd scripts && zip -v ../$ZIPNAME -r ./"$TEMP")
