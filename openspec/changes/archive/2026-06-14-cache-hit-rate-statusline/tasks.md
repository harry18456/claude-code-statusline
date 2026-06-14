## 1. Review Gate

- [x] 1.1 Gate final display through review before implementation: approved decision recorded as line 1 placement = tight to `$cost` inside the same cost segment with one separating space, non-ASCII indicator = `⚡<pct>%`, ASCII fallback = `cache:<pct>%`, color thresholds = `>=80%` gray / `50-79%` yellow / `<50%` red; verified by content review before apply work started.

## 2. Model Parsing

- [x] [P] 2.1 Add model tests for `Read JSON from stdin`: current usage object is available, `current_usage: null` is unavailable, absent `current_usage` is unavailable, tolerant decimal/scientific token values parse, and wrong scalar token values zero only that field; verified with `go test ./internal/model -run TestParsePayload_CurrentUsage`.
- [x] 2.2 Implement Parse current_usage as nullable data so `context_window.current_usage` exposes latest-request token usage without treating `null` or absent as a zero-token request; verified `Read JSON from stdin` with `go test ./internal/model -run TestParsePayload_CurrentUsage`.

## 3. Renderer Behavior

- [x] [P] 3.1 Add renderer tests for `Prompt cache hit-rate display`: the formula uses `cache_read_input_tokens / (input_tokens + cache_creation_input_tokens + cache_read_input_tokens)`, ignores `output_tokens`, rounds the compact display to whole percent, and verifies ASCII/default/NerdFont output; verified with `go test ./internal/renderer -run TestFormatCacheHit`.
- [x] 3.2 Implement Keep cache hit rate scoped to latest request and Keep renderer formatting isolated by adding `formatCacheHit` or an equivalent helper that uses only current usage from the latest request; verified with `go test ./internal/renderer -run TestFormatCacheHit`.
- [x] 3.3 Implement Suppress unusable cache data so unavailable current usage and denominator zero return an empty segment and never divide by zero; verified with `go test ./internal/renderer -run TestFormatCacheHit`.
- [x] 3.4 Implement `Render two-line ANSI output` insertion for the reviewed display choice while preserving existing line 1 segment order and line 2 behavior; verified with `go test ./internal/renderer -run TestRenderCacheHit`.

## 4. Final Verification

- [x] 4.1 Run formatting and static verification for the completed change; verified `gofmt -l .` printed no files, `go build ./...` passed, `go vet ./...` passed, `staticcheck ./...` passed, and `go test -race ./...` passed.
