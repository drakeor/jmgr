#!/usr/bin/env bash
set -euo pipefail

DATA_ROOT="bin"
DELETED=0

for d in "$DATA_ROOT"/node*; do
  if [ -d "$d" ]; then
    echo "Deleting $d"
    rm -rf "$d"
    DELETED=1
  fi
done

if [ "$DELETED" -eq 1 ]; then
  echo "All local BadgerDB node data under '$DATA_ROOT/node*' purged."
else
  echo "No node directories found under '$DATA_ROOT'."
fi
