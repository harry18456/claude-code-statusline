# rate-limit-countdown Specification

## Purpose

TBD - created by archiving change 'add-rate-limit-countdown'. Update Purpose after archive.

## Requirements

### Requirement: Parse resets_at timestamp

The program SHALL parse the `resets_at` Unix timestamp from each rate limit entry in the JSON payload and store it in the `RateLimit` struct.

#### Scenario: resets_at present

- **WHEN** the JSON payload contains `rate_limits.five_hour.resets_at` or `rate_limits.seven_day.resets_at`
- **THEN** the parsed value SHALL be stored as `ResetsAt int64` in the corresponding `RateLimit` struct

#### Scenario: resets_at absent

- **WHEN** the JSON payload does not include `resets_at` for a rate limit entry
- **THEN** `ResetsAt` SHALL be zero and no countdown SHALL be displayed


<!-- @trace
source: add-rate-limit-countdown
updated: 2026-04-03
code:
  - internal/renderer/renderer.go
  - debug.json
  - cmd/debug-tee/main.go
  - .spectra.yaml
  - statusline.sh
  - cc-statusline-debug.jsonl
  - internal/model/payload.go
tests:
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->

---
### Requirement: Display rate limit countdown

The program SHALL append a countdown to reset after the percentage value whenever `resets_at` is present, regardless of the used percentage.

#### Scenario: Countdown >= 24 hours

- **WHEN** `resets_at` is present and the remaining time is 24 hours or more
- **THEN** the countdown SHALL be displayed as `(Xd Yh)` appended to the rate limit value

#### Scenario: Countdown >= 60 minutes

- **WHEN** `resets_at` is present and the remaining time is between 60 minutes and 23 hours 59 minutes
- **THEN** the countdown SHALL be displayed as `(Xh Ym)` appended to the rate limit value

#### Scenario: Countdown < 60 minutes

- **WHEN** `resets_at` is present and the remaining time is between 1 and 59 minutes
- **THEN** the countdown SHALL be displayed as `(Ym)` appended to the rate limit value

#### Scenario: Countdown expired

- **WHEN** `resets_at` is present and `resets_at <= time.Now().Unix()`
- **THEN** the countdown SHALL be displayed as `(now)`

#### Scenario: Countdown always shown

- **WHEN** `resets_at` is non-zero
- **THEN** the countdown SHALL be appended regardless of used_percentage

#### Scenario: resets_at is zero

- **WHEN** `ResetsAt` is 0 (absent from payload)
- **THEN** no countdown SHALL be appended

<!-- @trace
source: add-rate-limit-countdown
updated: 2026-04-03
code:
  - internal/renderer/renderer.go
  - debug.json
  - cmd/debug-tee/main.go
  - .spectra.yaml
  - statusline.sh
  - cc-statusline-debug.jsonl
  - internal/model/payload.go
tests:
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->