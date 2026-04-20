## ADDED Requirements

### Requirement: Seven-day usage pace indicator

The program SHALL append a pace indicator to the `seven_day` rate limit display, positioned after the percentage with a single space separator and before the countdown. The indicator SHALL reflect whether actual usage deviates from the linear expected usage derived from elapsed window time. The pace indicator SHALL NOT be computed or displayed for the `five_hour` rate limit.

The expected usage is computed as:

```
window_length_seconds = 604800
elapsed               = window_length_seconds - (resets_at - now)
expected_pct          = elapsed / window_length_seconds * 100
deviation             = used_percentage - expected_pct
magnitude             = round(abs(deviation))   // nearest integer
```

When a magnitude is shown, it SHALL be formatted as `<N>%` where `<N>` is `magnitude` as an integer (no decimal point, no padding).

#### Scenario: Seven-day over-pace

- **WHEN** `seven_day.resets_at` is non-zero AND remaining window time `(resets_at - now)` is at least 10% of 604800 seconds AND `deviation > 5`
- **THEN** the program SHALL append ` ▲<N>%` (single space + red solid-up triangle + integer magnitude + percent sign) after the `seven_day` percentage and before the countdown (e.g., `7d:55% ▲7% (4d 2h)`)

#### Scenario: Seven-day under-pace

- **WHEN** `seven_day.resets_at` is non-zero AND remaining window time `(resets_at - now)` is at least 10% of 604800 seconds AND `deviation < -5`
- **THEN** the program SHALL append ` ▼<N>%` (single space + gray solid-down triangle + integer magnitude + percent sign) after the `seven_day` percentage and before the countdown (e.g., `7d:36% ▼7% (3d 9h)`)

#### Scenario: Seven-day within tolerance

- **WHEN** `seven_day.resets_at` is non-zero AND remaining window time `(resets_at - now)` is at least 10% of 604800 seconds AND the absolute value of `deviation` is at most 5
- **THEN** the program SHALL append ` ≈` (single space + gray approximately-equal sign, no magnitude) after the `seven_day` percentage and before the countdown (e.g., `7d:43% ≈ (3d 12h)`)

#### Scenario: Seven-day near-reset suppression

- **WHEN** `seven_day.resets_at` is non-zero AND remaining window time `(resets_at - now)` is less than 10% of 604800 seconds (less than 60480 seconds)
- **THEN** the program SHALL NOT append any pace indicator regardless of deviation

#### Scenario: Seven-day resets_at absent

- **WHEN** `seven_day.resets_at` is zero (absent from payload)
- **THEN** the program SHALL NOT append any pace indicator

#### Scenario: Seven-day pace indicator in ASCII mode

- **WHEN** `CLAUDE_STATUSLINE_ASCII=1` is set
- **THEN** the program SHALL substitute `^` for `▲`, `v` for `▼`, and `~` for `≈` in the appended segment (no ANSI color codes), preserving the `<N>%` magnitude suffix for over/under-pace cases

#### Scenario: Five-hour never shows pace indicator

- **WHEN** rendering the `five_hour` rate limit under any deviation or timing condition
- **THEN** the program SHALL NOT append any pace indicator to the `five_hour` display
