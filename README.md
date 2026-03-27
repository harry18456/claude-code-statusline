# ◆ claude-code-statusline

A real-time status line for [Claude Code](https://docs.anthropic.com/en/docs/claude-code). Displays model, context usage, cost, duration, git branch, and rate limits after every response.

[English](README.md) | [繁體中文](README.zh-TW.md)

---

## What You See

```
◆ Sonnet 4.6 │ ████████░░ 78% │ $1.23 │ 14m32s │ 5h:42% 7d:15%
⎇ main* │ +84/-12 │ my-project │ ⚙ code-reviewer
```

### Line 1

| Segment | Example | Description |
|---------|---------|-------------|
| `◆` | `◆` | Anthropic brand diamond (purple). ASCII mode: `<>` |
| Model | `Sonnet 4.6` | Current Claude model name |
| Progress bar | `████████░░` | 10-cell context window usage bar |
| Percentage | `78%` | Context used. Green < 70%, yellow 70–89%, red ≥ 90% |
| ⚠ warning | `⚠` | Appears only when context ≥ 90% |
| Context size | `200k` / `1M` | Shown only when not already in the model name |
| Cost | `$1.23` | Cumulative token cost this session (estimate). Yellow > $0, red ≥ $10, gray at $0.00 |
| Duration | `14m32s` | Total session time. Hidden if under 1 second |
| Rate limits | `5h:42%` `7d:15%` | 5-hour and 7-day quota usage (Claude Pro/Max only). Red when ≥ 80% |

### Line 2

| Segment | Example | Description |
|---------|---------|-------------|
| Branch | `⎇ main*` | Current git branch. `*` means uncommitted changes |
| Lines | `+84/-12` | Lines added/removed by Claude this session |
| Directory | `my-project` | Current working directory name |
| Indicator | `⚙ code-reviewer` | Active subagent name, or `⚙ worktree:name` if in a git worktree. Worktree takes priority |

Zero-value segments are hidden entirely (`+0/-0`, `0m0s`, missing rate limits).

---

## Installation

### Step 1 — Download the binary

Go to [Releases](https://github.com/harry18456/claude-code-statusline/releases/latest) and download the file for your platform:

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `statusline-darwin-arm64` |
| macOS (Intel) | `statusline-darwin-amd64` |
| Linux (x86_64) | `statusline-linux-amd64` |
| Windows (x86_64) | `statusline-windows-amd64.exe` |

### Step 2 — Place the binary

**macOS / Linux**

```bash
# Replace the filename with the one you downloaded
mv statusline-darwin-arm64 ~/.claude/statusline
chmod +x ~/.claude/statusline
```

**Windows** (PowerShell)

```powershell
Move-Item statusline-windows-amd64.exe "$env:USERPROFILE\.claude\statusline.exe"
```

### Step 3 — Configure Claude Code

Edit `~/.claude/settings.json` (create it if it does not exist).

**macOS / Linux**

```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/YOUR_USERNAME/.claude/statusline"
  }
}
```

**Windows**

```json
{
  "statusLine": {
    "type": "command",
    "command": "C:/Users/YOUR_USERNAME/.claude/statusline.exe"
  }
}
```

Replace `YOUR_USERNAME` with your actual username. Use forward slashes even on Windows.

If `settings.json` already has content, add the `"statusLine"` key inside the existing object:

```json
{
  "someOtherSetting": true,
  "statusLine": {
    "type": "command",
    "command": "/Users/YOUR_USERNAME/.claude/statusline"
  }
}
```

### Step 4 — Verify

Start or restart Claude Code. The status line should appear at the bottom of the terminal after the first response.

---

## Environment Variables

Set these in your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) or in the Claude Code `env` settings.

| Variable | Default | Effect |
|----------|---------|--------|
| `CLAUDE_STATUSLINE_ASCII` | `0` | Set to `1` for pure ASCII output (`#---`). Use when Unicode is unavailable |
| `CLAUDE_STATUSLINE_NERDFONT` | `0` | Set to `1` to enable Nerd Font icons (, 󰔟, ). Requires a [Nerd Font](https://www.nerdfonts.com/) in your terminal |
| `CLAUDE_STATUSLINE_POWERLINE` | follows `NERDFONT` | Set to `1` to use Powerline arrow separators (``) instead of `│`. Enabled automatically when `NERDFONT=1` |
| `COLORTERM` | system | Set to `truecolor` or `24bit` to enable the RGB gradient progress bar. Most modern terminals set this automatically |

### Rendering tiers

The binary auto-selects the best rendering based on environment:

| Tier | Condition | Progress bar style |
|------|-----------|--------------------|
| True color | `COLORTERM=truecolor` or `24bit` | Per-cell RGB gradient, green → yellow → red |
| ANSI | default | Solid color based on overall percentage |
| ASCII | `CLAUDE_STATUSLINE_ASCII=1` | `#` filled, `-` empty |

### Example: Nerd Font + true color

```bash
# Add to ~/.zshrc or ~/.bashrc
export CLAUDE_STATUSLINE_NERDFONT=1
export COLORTERM=truecolor
```

---

## Building from Source

Requires [Go](https://go.dev/) 1.21 or later.

```bash
git clone https://github.com/harry18456/claude-code-statusline.git
cd claude-code-statusline
go build -o ~/.claude/statusline ./cmd/statusline/
chmod +x ~/.claude/statusline
```

---

## Notes on cost display

The `$cost` value is an **estimate** of token usage for the current session, calculated from the Claude API token rates. If you use a Claude Pro or Max subscription, you are not billed per token — the number is informational only and will not match your billing page.

---

## License

[MIT](LICENSE)
