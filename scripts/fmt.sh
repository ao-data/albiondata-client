#!/usr/bin/env bash

echo "Running goimports to format go files..."
goimports -w $(find . -name '*.go' -not -path './vendor/*')
