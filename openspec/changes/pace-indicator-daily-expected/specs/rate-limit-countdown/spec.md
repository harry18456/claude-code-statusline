## MODIFIED Requirements

### Requirement: Seven-day usage pace indicator

The program SHALL append a pace indicator to the `seven_day` rate limit display, positioned after the percentage with a single space separator and before the countdown. The indicator SHALL reflect whether actual usage deviates from the expected per-day usage derived from the number of elapsed days within the seven-day window. The pace indicator SHALL NOT be computed or displayed for the `five_hour` rate limit. The pace indicator SHALL be shown whenever `seven_day.resets_at` is non-zero and the window has not yet elapsed, regardless of how much window time remains.

The expected usage is computed using day-level granularity (each day represents one seventh of total quota):

```
day_seconds           = 86400
window_length_seconds = 604800
elapsed               = window_length_seconds - (resets_at - now)
elapsed_days          = min(7, ceil(elapsed / day_seconds))
expected_pct          = elapsed_days * (100 / 7)
deviation             = used_percentage - expected_pct
magnitude             = round(abs(deviation))   // nearest integer
```

The `expected_pct` SHALL step at integer multiples of `day_seconds` measured from window start (where window start equals `resets_at - 604800`). The step boundaries SHALL align with the `resets_at` clock time, NOT with calendar midnight. The `expected_pct` SHALL equal `0` only at the exact instant `elapsed = 0`; for any `elapsed >= 1` second the program SHALL treat the user as being inside day 1 and use `expected_pct = 100/7 ≈ 14.2857`.

When a magnitude is shown, it SHALL be formatted as `<N>%` where `<N>` is `magnitude` as an integer (no decimal point, no padding).

#### Scenario: Seven-day over-pace

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is greater than zero AND `deviation > 5`
- **THEN** the program SHALL append ` ▲<N>%` (single space + red solid-up triangle + integer magnitude + percent sign) after the `seven_day` percentage and before the countdown (e.g., `7d:70% ▲13% (3d 23h)` when `elapsed_days = 4`, `expected_pct ≈ 57.14`)

#### Scenario: Seven-day under-pace

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is greater than zero AND `deviation < -5`
- **THEN** the program SHALL append ` ▼<N>%` (single space + gray solid-down triangle + integer magnitude + percent sign) after the `seven_day` percentage and before the countdown (e.g., `7d:40% ▼17% (3d 23h)` when `elapsed_days = 4`, `expected_pct ≈ 57.14`)

#### Scenario: Seven-day within tolerance

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is greater than zero AND the absolute value of `deviation` is at most 5
- **THEN** the program SHALL append ` ≈` (single space + gray approximately-equal sign, no magnitude) after the `seven_day` percentage and before the countdown (e.g., `7d:60% ≈ (3d 23h)` when `elapsed_days = 4`, `expected_pct ≈ 57.14`)

#### Scenario: Seven-day day-1 under-pace example

- **WHEN** `seven_day.resets_at` is non-zero AND `elapsed` is between 1 and 86400 seconds (inclusive at 86400) AND `used_percentage` is `6`
- **THEN** `elapsed_days` SHALL equal `1`, `expected_pct` SHALL be approximately `14.29`, `deviation` SHALL be approximately `-8.29`, and the program SHALL append ` ▼8%` after the `seven_day` percentage

#### Scenario: Seven-day day boundary step-up

- **WHEN** `elapsed` crosses an integer multiple of 86400 seconds (e.g., transitions from `86400` to `86401`)
- **THEN** `elapsed_days` SHALL increase by exactly 1 at the crossing AND `expected_pct` SHALL increase by approximately `14.29` percentage points

#### Scenario: Seven-day near-reset still shown

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is less than 60480 seconds (less than 10% of the 604800-second window)
- **THEN** the program SHALL still compute and append the pace indicator using the same formulas and thresholds as the over-pace, under-pace, and within-tolerance scenarios (e.g., `7d:12% ▼88% (1h 15m)` when `elapsed_days = 7`, `expected_pct = 100`)

#### Scenario: Seven-day elapsed_days capped at 7

- **WHEN** `elapsed` is greater than `604800` (e.g., due to clock skew)
- **THEN** `elapsed_days` SHALL be clamped to `7` AND `expected_pct` SHALL be `100`

#### Scenario: Seven-day at window start

- **WHEN** `elapsed` equals exactly `0` (the instant of reset)
- **THEN** `elapsed_days` SHALL equal `0` AND `expected_pct` SHALL equal `0`

#### Scenario: Seven-day resets_at absent

- **WHEN** `seven_day.resets_at` is zero (absent from payload)
- **THEN** the program SHALL NOT append any pace indicator

#### Scenario: Seven-day window already elapsed

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is less than or equal to zero
- **THEN** the program SHALL NOT append any pace indicator

#### Scenario: Seven-day pace indicator in ASCII mode

- **WHEN** `CLAUDE_STATUSLINE_ASCII=1` is set
- **THEN** the program SHALL substitute `^` for `▲`, `v` for `▼`, and `~` for `≈` in the appended segment (no ANSI color codes), preserving the `<N>%` magnitude suffix for over/under-pace cases

#### Scenario: Five-hour never shows pace indicator

- **WHEN** rendering the `five_hour` rate limit under any deviation or timing condition
- **THEN** the program SHALL NOT append any pace indicator to the `five_hour` display
