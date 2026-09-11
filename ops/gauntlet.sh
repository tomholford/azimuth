#!/usr/bin/env sh

set -e

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$SCRIPT_DIR/.."

echo "Running gauntlet: format → lint → test → build"
echo

echo "=== Step 1/4: Format ==="
./ops/fmt.sh

echo
echo "=== Step 2/4: Lint ==="
./ops/lint.sh

echo
echo "=== Step 3/4: Test ==="
./ops/test.sh

echo
echo "=== Step 4/4: Build ==="
./ops/build.sh

echo
echo "Gauntlet completed successfully!"
