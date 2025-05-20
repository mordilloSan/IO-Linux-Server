#!/usr/bin/env bash

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/.."

# Load .env and secret.env if present
if [[ -f "$ROOT_DIR/secret.env" ]]; then
    set -a
    source "$ROOT_DIR/secret.env"
    set +a
fi

# Prefer $HOME/linuxio-bridge if it exists, else use built location
if [[ -x "$HOME/linuxio-bridge" ]]; then
    BRIDGE_BIN="$HOME/linuxio-bridge"
elif [[ -x "$ROOT_DIR/go-backend/cmd/bridge/linuxio-bridge" ]]; then
    BRIDGE_BIN="$ROOT_DIR/go-backend/cmd/bridge/linuxio-bridge"
else
    echo "❌ Bridge binary not found!"
    exit 1
fi

# Kill old bridge
echo "🚦 Stopping any old linuxio-bridge processes..."
pid_list=$(ps -eo pid,user,comm,args | awk '/linuxio-bridge/ && $2=="root" {print $1}')
if [ -n "$pid_list" ]; then
    echo "$SUDO_PASSWORD" | sudo -S kill $pid_list
    echo "✅ linuxio-bridge stopped (PID(s): $pid_list)."
else
    echo "No running linuxio-bridge found."
fi

# Start bridge (run in background)
echo "🚦 Starting linuxio-bridge (privileged helper)..."
echo "$SUDO_PASSWORD" | sudo -S setsid $BRIDGE_BIN > /dev/null 2>&1 < /dev/null &

sleep 1