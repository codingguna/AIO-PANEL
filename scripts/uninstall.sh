#!/usr/bin/env sh
# ==============================================================================
# AIO-PANEL Uninstaller Wrapper
# Executes built-in Go CLI 'aio uninstall'
# ==============================================================================
set -e

if command -v aio >/dev/null 2>&1; then
  exec aio uninstall "$@"
elif [ -f "/usr/local/bin/aio" ]; then
  exec /usr/local/bin/aio uninstall "$@"
elif [ -f "./aio" ]; then
  exec ./aio uninstall "$@"
else
  echo "❌ Error: 'aio' binary not found to execute uninstallation."
  exit 1
fi
