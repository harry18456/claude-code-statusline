#!/usr/bin/env bash
# install.sh — One-line installer for claude-code-statusline (binary release)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/harry18456/claude-code-statusline/main/install.sh | bash
#   — or —
#   git clone https://github.com/harry18456/claude-code-statusline.git && cd claude-code-statusline && ./install.sh

set -euo pipefail

REPO="harry18456/claude-code-statusline"
TARGET="$HOME/.claude/statusline"
SETTINGS="$HOME/.claude/settings.json"

echo "◆ claude-code-statusline installer"
echo ""

# ─── Platform detection ───────────────────────────────────────────────────────

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin)
    case "$ARCH" in
      arm64)   BINARY="statusline-darwin-arm64" ;;
      x86_64)  BINARY="statusline-darwin-amd64" ;;
      *)
        echo "✗ Unsupported macOS architecture: $ARCH"
        echo ""
        echo "  Please download the binary manually from GitHub Releases:"
        echo "  https://github.com/$REPO/releases/latest"
        exit 1
        ;;
    esac
    ;;
  Linux)
    case "$ARCH" in
      x86_64)  BINARY="statusline-linux-amd64" ;;
      *)
        echo "✗ Unsupported Linux architecture: $ARCH"
        echo ""
        echo "  Please download the binary manually from GitHub Releases:"
        echo "  https://github.com/$REPO/releases/latest"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "✗ Unsupported platform: $OS ($ARCH)"
    echo ""
    echo "  This installer supports macOS (arm64, x86_64) and Linux (x86_64)."
    echo "  For other platforms, download the binary manually from GitHub Releases:"
    echo "  https://github.com/$REPO/releases/latest"
    exit 1
    ;;
esac

echo "  Platform : $OS / $ARCH"
echo "  Binary   : $BINARY"
echo ""

# ─── Downloader detection ─────────────────────────────────────────────────────

if command -v curl &>/dev/null; then
  DOWNLOADER="curl"
elif command -v wget &>/dev/null; then
  DOWNLOADER="wget"
else
  echo "✗ Neither curl nor wget found. Please install one and re-run this script."
  exit 1
fi

# ─── Resolve latest release URL ───────────────────────────────────────────────

DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$BINARY"

echo "  Downloading $BINARY ..."
echo "  URL: $DOWNLOAD_URL"
echo ""

mkdir -p "$(dirname "$TARGET")"

if [[ "$DOWNLOADER" == "curl" ]]; then
  curl -fL --progress-bar "$DOWNLOAD_URL" -o "$TARGET"
else
  wget --show-progress -q "$DOWNLOAD_URL" -O "$TARGET"
fi

chmod +x "$TARGET"

echo ""
echo "✓ Installed to $TARGET"
echo ""

# ─── settings.json snippet ────────────────────────────────────────────────────

if [[ -f "$SETTINGS" ]]; then
  if grep -q '"statusLine"' "$SETTINGS" 2>/dev/null; then
    echo "⚠  Your $SETTINGS already contains a statusLine entry."
    echo "   Update it to the following if needed:"
  else
    echo "Add the following to your $SETTINGS:"
  fi
else
  echo "~/.claude/settings.json not found."
  echo "Create it with the following content:"
  echo ""
  echo '{'
fi

echo ""
echo '  "statusLine": {'
echo '    "type": "command",'
echo "    \"command\": \"$TARGET\","
echo '    "timeout": 10'
echo '  }'

if [[ ! -f "$SETTINGS" ]]; then
  echo '}'
fi

echo ""
echo "✓ Done! Restart Claude Code to see the status line."
