## Why

Enterprise token billing makes prompt cache effectiveness a direct cost lever: cache reads are much cheaper than normal input tokens. The statusline should expose the latest request's prompt cache hit rate so engineers can see whether their current usage pattern is cache-efficient without opening logs or cost dashboards.

## What Changes

- Parse `context_window.current_usage` from Claude Code statusLine payloads. The field is nullable and represents the most recent API request, not a session total.
- Add token fields under current usage:
  - `input_tokens`
  - `cache_creation_input_tokens`
  - `cache_read_input_tokens`
- Reuse the existing tolerant integer parsing behavior for those token fields, so integer-like decimals, scientific notation, strings, wrong scalar types, and `null` do not drop the whole payload.
- Compute prompt cache hit rate from input-side tokens only:
  - denominator = `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`
  - numerator = `cache_read_input_tokens`
  - hit rate = numerator / denominator
- Suppress the cache hit segment when `current_usage` is `null`, absent, malformed at field level, or when the denominator is zero.
- Add a renderer helper such as `formatCacheHit` so cache hit calculation, rounding, coloring, and ASCII/NerdFont rendering can be tested independently.
- Add the cache hit segment to line 1 only after the display design below is approved.

### Display Design Options for Review

#### Option A: append after rate limits

Default/Nerd Font mockup:

```text
◆ Claude Opus 4.6 │ ███████--- 73% 200k │ $0.85 │ 3m42s │ 5h:15% 7d:8% │ ⚡99%
```

ASCII mockup:

```text
<> Claude Opus 4.6 | #######--- 73% 200k | $0.85 | 3m42s | 5h:15% 7d:8% | cache:99%
```

Review points:
- Position: after rate limits, because it is request telemetry rather than accumulated session cost.
- Icon/text: use a compact high-signal symbol in default/Nerd Font mode; use `cache:<pct>%` in ASCII mode.
- Color: gray for `>= 80%`, yellow for `50%..79%`, red for `< 50%`. High cache hit is good, so only degraded cache efficiency draws attention.

#### Option B: place near cost

Default/Nerd Font mockup:

```text
◆ Claude Opus 4.6 │ ███████--- 73% 200k │ $0.85 ⚡99% │ 3m42s │ 5h:15% 7d:8%
```

ASCII mockup:

```text
<> Claude Opus 4.6 | #######--- 73% 200k | $0.85 cache:99% | 3m42s | 5h:15% 7d:8%
```

Review points:
- Position: next to cost, because prompt cache hit rate explains input-token cost efficiency.
- Icon/text: same as Option A.
- Color: same thresholds as Option A, or gray for all values except red below `50%` to reduce line noise.

## Non-Goals

- Do not implement before display placement, icon/text, and color thresholds are approved.
- Do not calculate cumulative session cache hit rate.
- Do not include `output_tokens` in the hit-rate denominator.
- Do not estimate dollar savings or rewrite cost rendering.
- Do not change rate-limit semantics, context percentage, progress bar, or line 2.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `go-statusline`: parse latest request token usage from `context_window.current_usage` and render an optional prompt cache hit-rate segment on line 1.

## Impact

- Affected specs:
  - `openspec/specs/go-statusline/spec.md`
- Affected code for apply phase:
  - `internal/model/payload.go`
  - `internal/model/payload_test.go`
  - `internal/renderer/renderer.go`
  - `internal/renderer/renderer_test.go`
- No new runtime dependencies.
- Apply-phase verification target: `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test -race ./...`, and `gofmt -l .`.
