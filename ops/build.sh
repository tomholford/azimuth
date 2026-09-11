#!/usr/bin/env sh

set -e

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$SCRIPT_DIR/.."

echo "Building $(go list ./... | wc -w | tr -d ' ') packages"
go build ./...
echo "Done"
