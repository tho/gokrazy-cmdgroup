#!/bin/sh
#
# Integration test for cmdgroup using common Linux tools.
# Build the binary, then exercise the main usage patterns.
#
set -eu

CMDGROUP="$(mktemp -d)/cmdgroup"
trap 'rm -rf "$(dirname "$CMDGROUP")"' EXIT

echo "=== Building cmdgroup ==="
go build -o "$CMDGROUP" .

echo "=== Single instance ==="
# Run echo once with arguments.
"$CMDGROUP" echo hello world

echo "=== Multiple instances with global args ==="
# Global args "-l" are prepended to each instance.
# Instance 0: ls -l /tmp
# Instance 1: ls -l /var
"$CMDGROUP" ls -l -- /tmp -- /var

echo "=== Watched instance (auto-restart) ==="
# Watch instance 0: "echo" runs, exits, restarts, until we kill cmdgroup.
# Instance 1 runs once (unwatched).
"$CMDGROUP" -watch 0 echo -- ping -- pong &
PID=$!
sleep 3
kill "$PID"
wait "$PID" 2>/dev/null || true

echo "=== Exit code propagation ==="
# A failing instance causes a non-zero exit.
if "$CMDGROUP" false 2>/dev/null; then
    echo "FAIL: expected non-zero exit"
    exit 1
else
    echo "OK: non-zero exit propagated"
fi

echo "=== All tests passed ==="
