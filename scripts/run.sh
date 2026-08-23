#!/usr/bin/env bash

set -e

(cd frontend && npm ci && npm run build)
go run -ldflags="-w -s -X main.version=dev" albiondata-client.go
