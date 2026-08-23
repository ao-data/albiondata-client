#!/bin/bash
set -eo pipefail

cd client
# go:generate stringer -type=OperationType
# go:generate stringer -type=EventType
go generate
cd ..

GITHUB_REF_NAME=dev

(cd frontend && npm ci && npm run build)

# Native macOS build of the client, including the Wails dashboard UI.
# gopacket needs CGO + libpcap headers: brew install libpcap
export CGO_ENABLED=1
go build -ldflags "-X main.version=$GITHUB_REF_NAME" -o albiondata-client albiondata-client.go
