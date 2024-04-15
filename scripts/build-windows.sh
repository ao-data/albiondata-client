#!/usr/bin/env bash

set -eo pipefail

rm -f rsrc_windows_*
rm -f albiondata-client.exe
rm -f albiondata-client.*.bak
rm -f .albiondata-client.*.old

rm -f albiondata-client-amd64-installer.exe

go install github.com/tc-hib/go-winres@v0.3.1

export PATH="$PATH:/root/go/bin"

go-winres make

env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$GITHUB_REF_NAME" -o albiondata-client.exe -v -x albiondata-client.go

go-winres patch albiondata-client.exe

cd pkg/nsis
make nsis

cd ../..
ls -la albiondata-client*

cp albiondata-client.exe albiondata-client.exe.copy
gzip -9 albiondata-client.exe
mv albiondata-client.exe.gz update-windows-amd64.exe.gz
mv albiondata-client.exe.copy albiondata-client.exe
