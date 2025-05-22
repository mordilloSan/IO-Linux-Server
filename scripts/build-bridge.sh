#!/usr/bin/env bash
set -euo pipefail

GO_BIN="$(which go || echo /usr/local/go/bin/go)"
BRIDGE_SRC="${BRIDGE_SRC:-$(dirname "$0")/../go-backend/cmd/bridge}"
BRIDGE_BIN="${BRIDGE_BIN:-/usr/lib/linuxio/linuxio-bridge}"

if [[ ! -x "$GO_BIN" ]]; then
  echo "❌ 'go' binary not found at $GO_BIN!"
  exit 1
fi

# Require SUDO_PASSWORD if not already running as root
if [[ $EUID -ne 0 ]] && [[ -z "${SUDO_PASSWORD:-}" ]]; then
  echo "❌ SUDO_PASSWORD not set. Please run via 'make' or pass SUDO_PASSWORD env."
  exit 1
fi

cd "$BRIDGE_SRC"

"$GO_BIN" build -buildvcs=false -o "$BRIDGE_BIN"
chmod 755 "$BRIDGE_BIN"
echo "✅ Built bridge at $BRIDGE_BIN"
