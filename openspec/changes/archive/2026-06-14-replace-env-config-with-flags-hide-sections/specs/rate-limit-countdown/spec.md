## MODIFIED Requirements

### Requirement: Seven-day usage pace indicator

The program SHALL append a pace indicator to the `seven_day` rate limit display, positioned after the percentage with a single space separator and before the countdown. The indicator SHALL reflect whether actual usage deviates from the expected per-day usage derived from the number of elapsed days within the seven-day window. The pace indicator SHALL NOT be computed or displayed for the `five_hour` rate limit. The pace indicator SHALL be shown whenever `seven_day.resets_at` is non-zero and the window has not yet elapsed, regardless of how much window time remains.

The expected usage is computed using day-level granularity, where each day represents one seventh of total quota:

```
day_seconds           = 86400
window_length_seconds = 604800
elapsed               = window_length_seconds - (resets_at - now)
elapsed_days          = min(7, ceil(elapsed / day_seconds))
expected_pct          = elapsed_days * (100 / 7)
deviation             = used_percentage - expected_pct
magnitude             = max(1, round(abs(deviation)))   // floor at 1 when deviation != 0
```

The `expected_pct` SHALL step at integer multiples of `day_seconds` measured from window start, where window start equals `resets_at - 604800`. The step boundaries SHALL align with the `resets_at` clock time, NOT with calendar midnight. The `expected_pct` SHALL equal `0` only at the exact instant `elapsed = 0`; for any `elapsed >= 1` second the program SHALL treat the user as being inside day 1 and use `expected_pct = 100/7 ≈ 14.2857`.

The pace indicator SHALL use a zero tolerance threshold: any non-zero deviation SHALL produce a directional arrow (`▲` or `▼`). The neutral `≈` symbol SHALL be reserved for the exact `deviation == 0` case, which is theoretically reachable only when `used_percentage` happens to equal `expected_pct` exactly.

When a magnitude is shown, it SHALL be formatted as `<N>%` where `<N>` is `magnitude` as an integer, with no decimal point and no padding.

#### Scenario: Seven-day over-pace

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is greater than zero AND `deviation > 0`
- **THEN** the program SHALL append ` ▲<N>%` after the `seven_day` percentage and before the countdown

##### Example: over-pace display

- **GIVEN** `elapsed_days = 4`, `expected_pct ≈ 57.14`, `used_percentage = 60`, and `deviation ≈ +2.86`
- **WHEN** rendering the seven-day rate limit
- **THEN** the program SHALL append ` ▲3%`, producing a display such as `7d:60% ▲3% (3d 23h)`

#### Scenario: Seven-day under-pace

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is greater than zero AND `deviation < 0`
- **THEN** the program SHALL append ` ▼<N>%` after the `seven_day` percentage and before the countdown

##### Example: under-pace display

- **GIVEN** `elapsed_days = 1`, `expected_pct ≈ 14.29`, `used_percentage = 10`, and `deviation ≈ -4.29`
- **WHEN** rendering the seven-day rate limit
- **THEN** the program SHALL append ` ▼4%`, producing a display such as `7d:10% ▼4% (6d 16h)`

#### Scenario: Seven-day exact match

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is greater than zero AND `deviation == 0`
- **THEN** the program SHALL append ` ≈` after the `seven_day` percentage and before the countdown

#### Scenario: Seven-day day-1 under-pace example

- **WHEN** `seven_day.resets_at` is non-zero AND `elapsed` is between 1 and 86400 seconds inclusive AND `used_percentage` is `6`
- **THEN** `elapsed_days` SHALL equal `1`
- **THEN** `expected_pct` SHALL be approximately `14.29`
- **THEN** `deviation` SHALL be approximately `-8.29`
- **THEN** the program SHALL append ` ▼8%` after the `seven_day` percentage

#### Scenario: Seven-day magnitude floor at 1

- **WHEN** `(resets_at - now)` is greater than zero AND `deviation` is non-zero AND `round(abs(deviation))` would equal `0`
- **THEN** `magnitude` SHALL be set to `1`
- **THEN** the program SHALL append ` ▲1%` when `deviation > 0` or ` ▼1%` when `deviation < 0`
- **THEN** the program SHALL NOT display ` ▲0%` or ` ▼0%`

#### Scenario: Seven-day day boundary step-up

- **WHEN** `elapsed` crosses an integer multiple of 86400 seconds
- **THEN** `elapsed_days` SHALL increase by exactly 1 at the crossing
- **THEN** `expected_pct` SHALL increase by approximately `14.29` percentage points

#### Scenario: Seven-day near-reset still shown

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is less than 60480 seconds
- **THEN** the program SHALL still compute and append the pace indicator using the same formulas and zero-tolerance branching as the over-pace and under-pace scenarios

##### Example: near-reset under-pace display

- **GIVEN** `elapsed_days = 7`, `expected_pct = 100`, and `used_percentage = 12`
- **WHEN** rendering the seven-day rate limit
- **THEN** the program SHALL append ` ▼88%`, producing a display such as `7d:12% ▼88% (1h 15m)`

#### Scenario: Seven-day elapsed_days capped at 7

- **WHEN** `elapsed` is greater than `604800`
- **THEN** `elapsed_days` SHALL be clamped to `7`
- **THEN** `expected_pct` SHALL be `100`

#### Scenario: Seven-day at window start

- **WHEN** `elapsed` equals exactly `0`
- **THEN** `elapsed_days` SHALL equal `0`
- **THEN** `expected_pct` SHALL equal `0`

#### Scenario: Seven-day resets_at absent

- **WHEN** `seven_day.resets_at` is zero
- **THEN** the program SHALL NOT append any pace indicator

#### Scenario: Seven-day window already elapsed

- **WHEN** `seven_day.resets_at` is non-zero AND `(resets_at - now)` is less than or equal to zero
- **THEN** the program SHALL NOT append any pace indicator

#### Scenario: Seven-day pace indicator in ASCII mode

- **WHEN** `--ascii` is set
- **THEN** the program SHALL substitute `^` for `▲`, `v` for `▼`, and `~` for `≈` in the appended segment
- **THEN** the ASCII pace indicator SHALL contain no ANSI color codes
- **THEN** the ASCII pace indicator SHALL preserve the `<N>%` magnitude suffix for over-pace and under-pace cases

#### Scenario: Five-hour never shows pace indicator

- **WHEN** rendering the `five_hour` rate limit under any deviation or timing condition
- **THEN** the program SHALL NOT append any pace indicator to the `five_hour` display
