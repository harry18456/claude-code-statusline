#!/usr/bin/env bash
# test-mock.sh — Test the statusline binary with mock JSON data
#
# Usage: ./examples/test-mock.sh [scenario]
# Scenarios: normal, warning, danger, startup, agent, worktree, ascii, nerdfont
#
# The binary is built automatically if not found at the project root.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SCENARIO="${1:-all}"

# ── Locate or build the binary ──────────────────────────────────────────────

# On Windows (Git Bash) the binary has .exe extension
if [[ "$(uname -s)" == MINGW* || "$(uname -s)" == MSYS* || "$(uname -s)" == CYGWIN* ]]; then
  BIN="$PROJECT_ROOT/statusline.exe"
else
  BIN="$PROJECT_ROOT/statusline"
fi

if [[ ! -x "$BIN" ]]; then
  echo "Binary not found at $BIN — building..."
  (cd "$PROJECT_ROOT" && go build -o "$(basename "$BIN")" ./cmd/statusline/)
  echo "Built OK"
fi

# ── Test runner ──────────────────────────────────────────────────────────────

run_test() {
  local label="$1"
  local json="$2"
  shift 2

  echo ""
  echo "━━━ $label ━━━"
  if [[ $# -gt 0 ]]; then
    echo "$json" | env "$@" "$BIN"
  else
    echo "$json" | "$BIN"
  fi
  echo ""
}

# ── Test data ────────────────────────────────────────────────────────────────

JSON_NORMAL='{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"main"},"rate_limits":{"five_hour":{"used_percentage":15},"seven_day":{"used_percentage":8}}}'

JSON_WARNING='{"model":{"display_name":"Claude Sonnet 4.6"},"context_window":{"used_percentage":75,"context_window_size":200000},"cost":{"total_cost_usd":3.20,"total_lines_added":280,"total_lines_removed":45,"total_duration_ms":725000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"feat/auth"},"rate_limits":{"five_hour":{"used_percentage":48}}}'

JSON_DANGER='{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":92,"context_window_size":1000000},"cost":{"total_cost_usd":15.30,"total_lines_added":500,"total_lines_removed":120,"total_duration_ms":2712000},"workspace":{"current_dir":"/Users/dev/api-server"},"worktree":{"branch":"main"},"rate_limits":{"five_hour":{"used_percentage":85},"seven_day":{"used_percentage":62}}}'

JSON_STARTUP='{"model":{"display_name":"Opus 4.6 (1M context)"},"context_window":{"used_percentage":0,"context_window_size":1000000},"cost":{"total_cost_usd":0,"total_duration_ms":0},"workspace":{"current_dir":"/Users/dev/my-project"}}'

JSON_AGENT='{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"main"},"agent":{"name":"code-reviewer"}}'

JSON_WORKTREE='{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"worktree-my-feature","name":"my-feature","path":"/path/to/worktree"}}'

# ── Run tests ────────────────────────────────────────────────────────────────

case "$SCENARIO" in
  normal)   run_test "Normal (42%, green)" "$JSON_NORMAL" ;;
  warning)  run_test "Warning (75%, yellow)" "$JSON_WARNING" ;;
  danger)   run_test "Danger (92%, red + ⚠)" "$JSON_DANGER" ;;
  startup)  run_test "Session startup (zero values hidden)" "$JSON_STARTUP" ;;
  agent)    run_test "Agent mode (code-reviewer)" "$JSON_AGENT" ;;
  worktree) run_test "Worktree mode (my-feature)" "$JSON_WORKTREE" ;;
  ascii)    run_test "ASCII fallback" "$JSON_NORMAL" "CLAUDE_STATUSLINE_ASCII=1" ;;
  nerdfont) run_test "Nerd Font mode" "$JSON_NORMAL" "CLAUDE_STATUSLINE_NERDFONT=1" ;;
  all)
    run_test "Normal (42%, green)" "$JSON_NORMAL"
    run_test "Warning (75%, yellow)" "$JSON_WARNING"
    run_test "Danger (92%, red + ⚠)" "$JSON_DANGER"
    run_test "Session startup (zero values hidden)" "$JSON_STARTUP"
    run_test "Agent mode (code-reviewer)" "$JSON_AGENT"
    run_test "Worktree mode (my-feature)" "$JSON_WORKTREE"
    run_test "ASCII fallback" "$JSON_NORMAL" "CLAUDE_STATUSLINE_ASCII=1"
    run_test "Nerd Font mode" "$JSON_NORMAL" "CLAUDE_STATUSLINE_NERDFONT=1"
    ;;
  *)
    echo "Unknown scenario: $SCENARIO"
    echo "Available: normal, warning, danger, startup, agent, worktree, ascii, nerdfont, all"
    exit 1
    ;;
esac
