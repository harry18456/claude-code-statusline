# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

<!-- SPECTRA:START v1.0.1 -->

# Spectra Instructions

This project uses Spectra for Spec-Driven Development(SDD). Specs live in `openspec/specs/`, change proposals in `openspec/changes/`.

## Use `/spectra:*` skills when:

- A discussion needs structure before coding → `/spectra:discuss`
- User wants to plan, propose, or design a change → `/spectra:propose`
- Tasks are ready to implement → `/spectra:apply`
- There's an in-progress change to continue → `/spectra:ingest`
- User asks about specs or how something works → `/spectra:ask`
- Implementation is done → `/spectra:archive`

## Workflow

discuss? → propose → apply ⇄ ingest → archive

- `discuss` is optional — skip if requirements are clear
- Requirements change mid-work? Plan mode → `ingest` → resume `apply`

## Parked Changes

Changes can be parked（暫存）— temporarily moved out of `openspec/changes/`. Parked changes won't appear in `spectra list` but can be found with `spectra list --parked`. To restore: `spectra unpark <name>`. The `/spectra:apply` and `/spectra:ingest` skills handle parked changes automatically.

<!-- SPECTRA:END -->

## Commands

### Local development
```bash
./dev.sh build      # Compile with git-describe version + install to ~/.claude/statusline.exe
./dev.sh last-json  # Extract latest Claude payload from debug log → ./debug.json
```

### Testing
```bash
# Run all Go unit tests
go test ./...

# Run a specific test
go test ./internal/renderer/ -v -run TestRenderWarningSymbolAt90
```

### Checking installed version
```bash
~/.claude/statusline.exe --version
```

### Building without installing
```bash
go build ./cmd/statusline/
```

## Architecture

This project is a Go binary (`cmd/statusline/main.go`) that acts as a Claude Code `statusLine` hook. Claude Code pipes a JSON payload to the binary via stdin after every assistant response; the binary outputs 2 lines of ANSI-colored text.

### Data flow

```
Claude Code → JSON via stdin → encoding/json (Go) → typed structs → assembled output lines
```

The JSON payload is decoded into typed Go structs defined in `internal/model/`. The renderer in `internal/renderer/` assembles the final output from those structs. The renderer does no IO: `cmd/statusline` resolves git state (`internal/gitcache`) and the session peer name (`internal/peername`) first and passes them in.

### Session peer name lookup

`internal/peername` maps the payload `session_id` to the name other Claude Code sessions use as the `SendMessage` address. It reads Claude Code's undocumented session registry, `<config>/sessions/<pid>.json` (fields `sessionId`, `name`, `updatedAt`), with `<config>` = `$CLAUDE_CONFIG_DIR` then `~/.claude`. Only regular `*.json` files up to 64 KiB are read; `*.key` files in the same directory are secrets and are never opened. Every failure resolves to "no name", so the section silently disappears. The payload `session_name` looks like an AI title and can differ from this address, so the lookup never uses it. Claude Code can rename a session at any time (interactive sessions start as `<folder>-<xx>` with two hex characters, background jobs as their job id); every render re-reads the registry, so the section follows renames. On Windows the files are opened with `FILE_SHARE_DELETE` so Claude Code can still delete them during a read. `bridgeSessionId` is a string only while the session is connected to Remote Control (null after a disconnect, a new value after each reconnect); `Lookup` returns it as `Info.BridgeSessionID`, from the same winning entry as the name, after an ASCII `[A-Za-z0-9_-]{1,128}` allowlist check. Sessions on other machines can only reach `bridge:<bridgeSessionId>`, never the local name.

### Rendering tiers

The binary auto-detects terminal capabilities and renders accordingly:

| Mode | Trigger | Bar style |
|------|---------|-----------|
| True color | `COLORTERM=truecolor\|24bit` | Per-cell RGB gradient (green→yellow→red) |
| ANSI | default | Solid color based on overall percentage |
| ASCII | `--ascii` flag | `#` filled, `-` empty |

### Git dirty-check caching

Git status is cached in `os.TempDir()/claude-statusline-git-cache` for 5 seconds. The cache format is `branch|dirty_flag` (`1`/`0`). Cache freshness is checked using `os.Stat().ModTime()` (cross-platform; fixes the macOS-only `stat -f %m` bug from the original bash version).

### Output structure

- **Line 1**: `◆ model ⚙ effort │ progress_bar pct% size │ $cost ⚡cache% │ duration │ rate_limits` — the effort glyph (`⚙ low`…`max`, with an optional gray `T`/`F` suffix for thinking/fast mode) sits right after the model name; `size` is the context-window label (`200k`/`1M`, with `1M` turning red past the 200k threshold); a `⚠` follows the percentage at ≥90%; `⚡cache%` is the latest request's prompt-cache hit rate; rate limits show as `5h:85% (1h 23m) 7d:55% ▲7% (3d 9h)` when `resets_at` is present (the 7d bucket carries a `▲`/`▼` pace arrow), countdown format: `(Xd Yh)` / `(Xh Ym)` / `(Ym)` / `(now)`
- **Line 2**: `⎇ branch* │ +added/-removed │ dirname │ ⚙ agent_or_worktree │ ⇄ peer_name` — note the line-1 `⚙` is the effort glyph, while this line-2 `⚙` is the active agent or worktree; `⇄ peer_name` (ASCII `peer:<name>`) is this session's `SendMessage` address from the session registry, hidden when no registry entry matches; while Remote Control is connected it continues with ` · bridge:<bridgeSessionId>` (ASCII ` bridge:<id>`), the address sessions on other machines must use

Zero-value sections are omitted entirely. The `$0.00` cost is shown but dimmed. Cache hit rate is hidden when `current_usage` is absent. Duration is suppressed if under 1 second. Worktree indicator takes priority over agent indicator. Any section can be suppressed with `--hide <keys>`.

### CLI flags

Behavior is configured via flags appended to the `statusLine.command` in `settings.json`. The `CLAUDE_STATUSLINE_ASCII`/`NERDFONT`/`POWERLINE` env vars were removed. Use these flags instead:

| Flag | Effect |
|------|--------|
| `--ascii` | Pure ASCII, no Unicode (takes priority over `--nerdfont`/`--powerline`) |
| `--nerdfont` | Nerd Font icons + Powerline separators |
| `--powerline` | `\ue0b0` arrow separators without Nerd Font icons |
| `--hide <keys>` | Hide comma-separated sections: `model,effort,bar,size,cost,cache,duration,rate,branch,lines,dir,agent,session,bridge` (repeatable; hiding `session` also skips the registry read; `bridge` hides only the Remote Control address) |
| `--version` | Print version and exit |

Invalid flags, unknown `--hide` keys, and conflicting render flags emit a stderr warning (visible with `claude --debug`) but still render and exit 0.

### Environment variables

| Variable | Default | Effect |
|----------|---------|--------|
| `COLORTERM` | system | `truecolor`/`24bit` enables the RGB gradient bar (terminal capability, not project config) |
| `CLAUDE_CONFIG_DIR` | `~/.claude` | Claude Code config dir; the session registry is read from `<dir>/sessions/`, then `~/.claude/sessions/` (set by Claude Code, not project config) |

### Version injection

The `version` variable in `main.go` defaults to `"dev"`. Release builds inject the git tag via:
```
-ldflags="-X main.version=v1.0.0"
```
This is handled automatically by `.github/workflows/release.yml` using `${{ github.ref_name }}`. Local builds via `./dev.sh build` inject `git describe --tags --dirty`. The binary exposes this via `--version`.

### CI

`.github/workflows/test.yml` runs `go vet ./...` and `go test -race -count=1 -v ./...` on `ubuntu-latest` and `windows-latest` for every push to `main` and every pull request (`fail-fast: false`, read-only token). `release.yml` calls it as its first job, so a `v*` tag builds and publishes binaries only after both test jobs pass. The Windows job is the only place where the Windows-specific tests (named pipe, `FILE_SHARE_DELETE`, symlink skipping with admin rights) run outside a developer machine.
