#!/usr/bin/env sh
# ==============================================================================
# AIO-PANEL Installer Wrapper
# Executes built-in Go CLI 'aio install'
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN="${SCRIPT_DIR}/aio"

if [ ! -f "${BIN}" ]; then
  BIN="/usr/local/bin/aio"
fi

if [ ! -f "${BIN}" ]; then
  echo "❌ Error: 'aio' binary not found in ${SCRIPT_DIR} or /usr/local/bin"
  exit 1
fi

chmod +x "${BIN}"
exec "${BIN}" install "$@"
