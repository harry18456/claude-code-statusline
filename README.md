# ◆ claude-code-statusline

A real-time status line for [Claude Code](https://docs.anthropic.com/en/docs/claude-code). Displays model, context usage, cost, duration, git branch, and rate limits after every response.

[English](README.md) | [繁體中文](README.zh-TW.md)

---

## What You See

```
◆ Sonnet 4.6 ⚙ high T │ ████████░░ 78% 1M │ $1.23 ⚡96% │ 14m32s │ 5h:85% (1h 23m) 7d:55% ▲7% (3d 9h)
⎇ main* │ +84/-12 │ my-project/internal/renderer │ ⚙ code-reviewer
```

### Line 1

| Segment | Example | Description |
|---------|---------|-------------|
| `◆` | `◆` | Anthropic brand diamond (purple). ASCII mode: `<>` |
| Model | `Sonnet 4.6` | Current Claude model name |
| Execution effort | `⚙ max` | Reasoning effort from the payload's `effort.level` (`low`/`medium`/`high`/`xhigh`/`max`), shown right after the model name. Color ramps with cost risk: low gray, medium cyan, high yellow, xhigh red, max bold red. A gray `T`/`F` suffix appears when extended thinking / fast mode are on (`⚙ max T`, `⚙ high TF`); off or absent signals are not shown. Hidden entirely when the field is absent (older Claude Code, or a model without effort support) or the level is unknown. ASCII: `effort:max think fast` |
| Progress bar | `████████░░` | 10-cell context window usage bar |
| Percentage | `78%` | Context used. Green < 70%, yellow 70–89%, red ≥ 90% |
| ⚠ warning | `⚠` | Appears only when context ≥ 90% |
| Context size | `200k` / `1M` | Driven purely by the payload's `context_window_size` field. `1M` turns **red** when `exceeds_200k_tokens=true`, signalling the session has crossed the 200k-token premium-pricing threshold (input 2×, output 1.5×) |
| Cost | `$1.23` | Cumulative token cost this session (estimate). Yellow > $0, red ≥ $10, gray at $0.00 |
| Cache hit rate | `⚡99%` | Prompt cache hit rate for the **latest request** (not a session total): `cache_read / (input + cache_creation + cache_read)`. Gray ≥ 80%, yellow 50–79%, red < 50%. Hidden when `current_usage` is absent or null (session start, just after `/compact`) or the denominator is zero. ASCII: `cache:99%` |
| Duration | `14m32s` | Total session time. Hidden if under 1 second |
| Rate limits | `5h:85% (1h 23m) 7d:55% ▲7% (3d 9h)` | 5-hour and 7-day quota usage (Claude Pro/Max only). Red when ≥ 80%. Countdown to reset appended when available: `(Xd Yh)` / `(Xh Ym)` / `(Ym)` / `(now)` |
| 7d pace | `▲7%` / `▼3%` / `≈` | Deviation from daily linear expected usage on the `seven_day` bucket: `expected = ceil(elapsed / 1 day) × (100/7)`, so day 1 expects 14.29%, day 2 expects 28.57%, … day 7 expects 100%. Step boundaries align with the `resets_at` clock time, not calendar midnight. Any non-zero deviation produces a directional indicator: red `▲<N>%` when over-pace, gray `▼<N>%` when under-pace; `<N>` is `round(abs(deviation))` floored at `1`. Gray `≈` only appears when deviation is exactly zero (rare, since `100/7` is not a finite decimal). Suppressed only when `resets_at` is missing or the window has already elapsed. Never shown for the 5-hour bucket. ASCII fallbacks: `^<N>%` / `v<N>%` / `~` |

### Line 2

| Segment | Example | Description |
|---------|---------|-------------|
| Branch | `⎇ main*` | Current git branch. `*` means uncommitted changes |
| Lines | `+84/-12` | Lines added/removed by Claude this session |
| Directory | `my-project/internal/renderer` | Project root + relative path (forward slashes). Root resolves via three-step fallback: (1) `workspace.project_dir` if it is a strict ancestor of the current dir, (2) walk upward for a `.git` entry (file or directory), (3) fall back to the current directory's base name. Shows only the base name when the current dir equals the root |
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

To check the installed version at any time:

```bash
~/.claude/statusline --version   # macOS / Linux
~/.claude/statusline.exe --version  # Windows
```

---

## Command-line flags

Configure statusline behavior directly in the Claude Code `settings.json` command.

```json
{
  "statusLine": {
    "type": "command",
    "command": "C:/Users/YOUR_USERNAME/.claude/statusline.exe --nerdfont --hide effort,duration"
  }
}
```

Breaking change: `CLAUDE_STATUSLINE_ASCII`, `CLAUDE_STATUSLINE_NERDFONT`, and `CLAUDE_STATUSLINE_POWERLINE` are no longer read. Use command-line flags instead.

| Flag | Effect |
|------|--------|
| `--ascii` | Pure ASCII output (`#---`). ASCII takes priority if combined with `--nerdfont` or `--powerline` |
| `--nerdfont` | Enable Nerd Font icons and Powerline separators. Requires a [Nerd Font](https://www.nerdfonts.com/) in your terminal |
| `--powerline` | Enable Powerline arrow separators without enabling Nerd Font icons |
| `--hide <keys>` | Hide comma-separated sections. May be repeated; known keys are merged |
| `--version` | Print the binary version and exit without rendering a status line |

Hide keys: `model`, `effort`, `bar`, `size`, `cost`, `cache`, `duration`, `rate`, `branch`, `lines`, `dir`, `agent`.

Invalid configuration is tolerant: unknown hide keys, conflicting rendering flags, unknown flags, missing `--hide` values, and positional arguments write warnings to stderr for `claude --debug`, then render a status line and exit 0.

`COLORTERM=truecolor|24bit` is still honored as terminal capability detection for the RGB gradient progress bar. It is not project configuration.

### Rendering tiers

The binary selects rendering from flags and terminal capabilities:

| Tier | Condition | Progress bar style |
|------|-----------|--------------------|
| True color | `COLORTERM=truecolor` or `24bit` | Per-cell RGB gradient, green → yellow → red |
| ANSI | default | Solid color based on overall percentage |
| ASCII | `--ascii` | `#` filled, `-` empty |

### Example: Nerd Font + hidden sections

```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/YOUR_USERNAME/.claude/statusline --nerdfont --hide effort,duration,rate"
  }
}
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
