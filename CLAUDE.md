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

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Building
```bash
go build ./cmd/statusline/
```

### Testing
```bash
# Run all Go unit tests
go test ./...

# Run tests for internal packages only
go test ./internal/...

# Run all display scenarios against the built binary
./examples/test-mock.sh

# Run a specific scenario
./examples/test-mock.sh normal     # 42%, green
./examples/test-mock.sh warning    # 75%, yellow
./examples/test-mock.sh danger     # 92%, red + ⚠
./examples/test-mock.sh startup    # zero-value state
./examples/test-mock.sh agent      # subagent indicator
./examples/test-mock.sh worktree   # worktree indicator
./examples/test-mock.sh ascii      # ASCII fallback mode
./examples/test-mock.sh nerdfont   # Nerd Font mode
```

### Installing
```bash
./install.sh   # Downloads the pre-built binary from GitHub Releases
```

### Slides generation (docs only)
```bash
npm install
node docs/slides.js   # Generates the PPTX presentation
```

## Architecture

This project is a Go binary (`cmd/statusline/main.go`) that acts as a Claude Code `statusLine` hook. Claude Code pipes a JSON payload to the binary via stdin after every assistant response; the binary outputs 2 lines of ANSI-colored text.

### Data flow

```
Claude Code → JSON via stdin → encoding/json (Go) → typed structs → assembled output lines
```

The JSON payload is decoded into typed Go structs defined in `internal/model/`. The renderer in `internal/renderer/` assembles the final output from those structs.

### Rendering tiers

The binary auto-detects terminal capabilities and renders accordingly:

| Mode | Trigger | Bar style |
|------|---------|-----------|
| True color | `COLORTERM=truecolor\|24bit` | Per-cell RGB gradient (green→yellow→red) |
| ANSI | default | Solid color based on overall percentage |
| ASCII | `CLAUDE_STATUSLINE_ASCII=1` | `#` filled, `-` empty |

### Git dirty-check caching

Git status is cached in `/tmp/claude-statusline-git-cache` for 5 seconds. The cache format is `branch|dirty_marker`. Cache freshness is checked using `os.Stat().ModTime()` (cross-platform Go standard library).

### Output structure

- **Line 1**: `◆ model │ progress_bar pct% │ $cost │ duration │ rate_limits`
- **Line 2**: `⎇ branch* │ +added/-removed │ dirname │ ⚙ agent_or_worktree`

Zero-value sections are omitted entirely. The `$0.00` cost is shown but dimmed.

### Environment variables

| Variable | Default | Effect |
|----------|---------|--------|
| `CLAUDE_STATUSLINE_ASCII` | `0` | Pure ASCII, no Unicode |
| `CLAUDE_STATUSLINE_NERDFONT` | `0` | Nerd Font icons + optional Powerline |
| `CLAUDE_STATUSLINE_POWERLINE` | follows NERDFONT | `` arrow separators |
| `COLORTERM` | system | `truecolor`/`24bit` enables RGB gradient |

### File layout

- `cmd/statusline/main.go` — entry point; reads stdin and writes output
- `internal/model/` — JSON struct definitions for the Claude Code payload
- `internal/renderer/` — ANSI output assembly and rendering tiers
- `internal/gitcache/` — git branch and dirty-state caching logic
- `install.sh` — downloads the pre-built binary from GitHub Releases
- `examples/test-mock.sh` — pipes mock JSON into the binary for local testing
- `docs/slides.js` — standalone Node.js script using `pptxgenjs` to generate the presentation; has no relation to the statusline runtime
