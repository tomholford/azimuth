#!/usr/bin/env sh

set -e

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$SCRIPT_DIR/.."

echo "Go version: $(go version)"
echo "Running tests for $(go list ./... | wc -w | tr -d ' ') packages"

go test -count=1 ./...
