## Why

Claude Code controls the JSON serialization shape of statusLine payload numbers. `context_window_size`, `total_duration_ms`, and rate-limit `resets_at` currently decode into `int64`, so valid JSON numbers such as `1000000.0`, `1e6`, or `1.7e9` can make the whole payload fail and collapse the statusline to the parse-error fallback.

The payload parser also maintains `Payload` and `payloadJSON` as near-duplicate structs. That duplication makes every new payload field a two-place edit and increases drift risk.

## What Changes

- Make parsing tolerant for integer-target numeric payload fields that can arrive as integer, decimal, or scientific-notation JSON numbers.
- Keep malformed or empty JSON as a hard parse error that triggers the existing fallback line.
- Treat unparseable values for the tolerant numeric fields as field-level failures: leave the affected field at its zero value and keep the rest of the payload available for rendering.
- Move rate-limit presence detection into `RateLimit` / `RateLimits` unmarshalling so `payloadJSON` can be removed or reduced to a non-mirroring adapter.
- Add table-driven tests covering decimal numbers, scientific notation, missing fields, malformed JSON, and normal payload regression.

## Non-Goals

- No renderer redesign, ANSI format change, or new statusline sections.
- No broad schema validation layer for every payload field.
- No change to the parse-error fallback for empty input or syntactically malformed JSON.
- No commit in this phase.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `go-statusline`: payload parsing distinguishes malformed JSON from recoverable numeric field shape drift.
- `rate-limit-countdown`: `resets_at` parsing accepts integer-like decimal and scientific-notation JSON numbers without dropping the entire payload.

## Impact

- Affected code: `internal/model/payload.go`, `internal/model/payload_test.go`.
- Possible supporting code: `cmd/statusline/main.go` only if fallback behavior needs a small adjustment while preserving exit code 0.
- No new runtime dependencies.
- Verification target for apply phase: `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test -race ./...`, and `gofmt -l .`.
