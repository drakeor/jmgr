#!/usr/bin/env bash
set -euo pipefail

BIN="bin/monitord"
DATA_ROOT="bin"              # throw the data-dirs next to the binary
INTERVAL="3s"                # sample interval
#GODEBUG_FLAGS="schedtrace=1000,scheddetail=1"
GODEBUG_FLAGS=""

NODES=(
  "node1 127.0.0.1:4000 $DATA_ROOT/node1"
  "node2 127.0.0.1:4001 $DATA_ROOT/node2 127.0.0.1:4000"
  "node3 127.0.0.1:4002 $DATA_ROOT/node3 127.0.0.1:4000"
  "node4 127.0.0.1:4003 $DATA_ROOT/node4 127.0.0.1:4000"
)

pids=()
cleanup() {
  echo; echo "⇢ stopping cluster..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null || true    # polite SIGTERM
  done

  # wait up to 2 s, then SIGKILL stubborn ones
  deadline=$((SECONDS+2))
  for pid in "${pids[@]}"; do
    while kill -0 "$pid" 2>/dev/null && [ $SECONDS -lt $deadline ]; do
      sleep 0.1
    done
    echo "  - force stopping $pid"
    kill -9 "$pid" 2>/dev/null || true
    echo "  - stopped $pid"
  done

  wait
}
trap cleanup INT TERM EXIT

for spec in "${NODES[@]}"; do
  read -r ID ADDR DIR JOIN <<<"$spec"
  mkdir -p "$DIR"
  GODEBUG=$GODEBUG_FLAGS "$BIN" -id "$ID" -addr "$ADDR" -data-dir "$DIR" -join "$JOIN" -v 2>&1 \
      | sed -E "s/^/[${ID}] /" &
  pids+=("$!")
  sleep 1
done

echo "✓ cluster running (Ctrl-C to stop)…"
wait
