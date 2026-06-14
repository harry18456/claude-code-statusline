## Why

Claude Code 2.1.177+ payload exposes `effort.level`, `thinking.enabled`, and `fast_mode`; these fields describe execution mode and directly affect thinking-token spend. The statusline already exposes cost and cache efficiency, so execution mode belongs on line 1 as another visible cost driver.

## What Changes

- Parse top-level `effort`, `thinking`, and `fast_mode` fields without breaking older Claude Code payloads where those fields are absent.
- Represent absent/null fields with pointer or presence-aware data so renderers can distinguish unavailable data from explicit false/off values.
- Add a renderer helper such as `formatEffort` or `formatExecutionMode` that returns an empty segment when mode data is absent, invalid, or intentionally suppressed by the reviewed display design.
- Add one optional line 1 execution-mode segment after product review selects the final display design.
- Preserve all existing cache, cost, duration, rate-limit, progress-bar, line 2, and environment-variable behavior.

### Display Design Options for Review

#### Option A: compact model-adjacent effort badge

Default/Nerd Font mockup:

```text
◆ Claude Opus 4.6 ⚙max │ ███████--- 73% 200k │ $0.85 ⚡99% │ 3m42s │ 5h:15% 7d:8%
```

ASCII mockup:

```text
<> Claude Opus 4.6 effort:max | #######--- 73% 200k | $0.85 cache:99% | 3m42s | 5h:15% 7d:8%
```

Review points:

- Position: directly after model name, because effort is a model invocation attribute rather than elapsed telemetry.
- Format: `⚙<level>` in default/Nerd Font mode; `effort:<level>` in ASCII mode.
- Color: level intensity maps to cost risk. Proposed scale: `low` gray, `medium` cyan, `high` yellow, `xhigh` red, `max` bold red or magenta+red.
- `thinking.enabled` and `fast_mode`: do not display by default. They are weaker signals than effort and would widen line 1. Add small suffixes only when product decides they are actionable, for example `⚙max T` for thinking on and `⚙max F` for fast mode on.

#### Option B: combined mode segment after cost/cache

Default/Nerd Font mockup:

```text
◆ Claude Opus 4.6 │ ███████--- 73% 200k │ $0.85 ⚡99% │ mode:max think fast:off │ 3m42s │ 5h:15% 7d:8%
```

ASCII mockup:

```text
<> Claude Opus 4.6 | #######--- 73% 200k | $0.85 cache:99% | mode:max think fast:off | 3m42s | 5h:15% 7d:8%
```

Review points:

- Position: after cost/cache, because execution mode explains token spend and sits near other cost signals.
- Format: explicit `mode:<effort>` text. This is wider but avoids ambiguous icon-only semantics.
- Color: apply the effort intensity color only to the effort value; keep `think` and `fast:<state>` gray.
- `thinking.enabled` and `fast_mode`: display only when their payload fields are present. `thinking.enabled=true` renders `think`; `fast_mode=true` renders `fast`; `fast_mode=false` renders `fast:off` only if product wants off-state visibility.

## Non-Goals

- Do not implement before the display placement, text/icon format, and color behavior are approved.
- Do not add config flags, hide lists, CLI flags, or new `CLAUDE_STATUSLINE_*` environment variables.
- Do not change cache hit-rate calculation or placement.
- Do not change cost, duration, rate-limit, context label, progress-bar, branch, directory, agent, or worktree behavior.
- Do not infer effort from model name or other fields when `effort.level` is absent.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `go-statusline`: parse Claude Code execution-mode payload fields and render an optional execution-mode segment on line 1.

## Impact

- Affected specs:
  - `openspec/specs/go-statusline/spec.md`
- Affected code for apply phase:
  - `internal/model/payload.go`
  - `internal/model/payload_test.go`
  - `internal/renderer/renderer.go`
  - `internal/renderer/renderer_test.go`
- No changes to `cmd/statusline/main.go` environment-variable parsing.
- No new runtime dependencies.
- Apply-phase verification target: `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test -race ./...`, and `gofmt -l .`.
