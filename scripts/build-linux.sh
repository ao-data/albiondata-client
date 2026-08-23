#!/usr/bin/env bash

set -eo pipefail

sudo apt-get update && sudo apt-get install -y libpcap-dev patchelf libgtk-4-dev libwebkitgtk-6.0-dev

(cd frontend && npm ci && npm run build)
go build -ldflags "-s -w -X main.version=$GITHUB_REF_NAME" albiondata-client.go
patchelf --replace-needed libpcap.so.0.8 libpcap.so albiondata-client

./albiondata-client -version

cp albiondata-client albiondata-client.old
gzip -9 albiondata-client
mv albiondata-client.gz update-linux-amd64.gz
mv albiondata-client.old albiondata-client
