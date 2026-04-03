#!/usr/bin/env bash
set -euo pipefail

INSTALL_PATH="${STATUSLINE_INSTALL:-$HOME/.claude/statusline.exe}"
LOG_FILE="${STATUSLINE_LOG:-${TEMP:-${TMPDIR:-/tmp}}/cc-statusline-debug.jsonl}"

usage() {
  echo "Usage: $0 <command>"
  echo ""
  echo "Commands:"
  echo "  build      Compile statusline and replace $INSTALL_PATH"
  echo "  last-json  Extract latest JSON payload from debug log to ./debug.json"
  exit 1
}

cmd_build() {
  local ver
  ver=$(git describe --tags --dirty 2>/dev/null || echo "dev")
  echo "Building $ver..."
  if ! go build -ldflags="-X main.version=${ver}" ./cmd/statusline/; then
    echo "Build failed." >&2
    exit 1
  fi
  cp statusline.exe "$INSTALL_PATH"
  echo "Installed: $INSTALL_PATH"
}

cmd_last_json() {
  if [[ ! -f "$LOG_FILE" ]]; then
    echo "Error: log file not found: $LOG_FILE" >&2
    echo "Enable debug-tee first by pointing statusLine command to debug-tee.exe." >&2
    exit 1
  fi
  grep -v '^$' "$LOG_FILE" | tail -n 1 > ./debug.json
  echo "Written: ./debug.json"
}

case "${1:-}" in
  build)     cmd_build ;;
  last-json) cmd_last_json ;;
  *)         usage ;;
esac
