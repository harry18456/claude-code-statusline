## MODIFIED Requirements

### Requirement: Parse resets_at timestamp

The program SHALL parse the `resets_at` Unix timestamp from each rate limit entry in the JSON payload and store it in the `RateLimit` struct.

The `resets_at` field SHALL accept JSON numbers that represent integer Unix timestamps, including integer literals, integer-like decimal literals, and integer-like scientific-notation literals. When a rate limit entry is present but `resets_at` cannot be converted to an integer timestamp, the program SHALL keep the rate limit entry present, SHALL set `ResetsAt` to zero, and SHALL continue parsing the rest of the payload.

#### Scenario: resets_at present

- **WHEN** the JSON payload contains `rate_limits.five_hour.resets_at` or `rate_limits.seven_day.resets_at`
- **THEN** the parsed value SHALL be stored as `ResetsAt int64` in the corresponding `RateLimit` struct

#### Scenario: resets_at decimal number

- **WHEN** the JSON payload contains `rate_limits.five_hour.resets_at` or `rate_limits.seven_day.resets_at` as an integer-like decimal JSON number such as `1700000000.0`
- **THEN** the parsed value SHALL be stored as `ResetsAt int64` in the corresponding `RateLimit` struct

#### Scenario: resets_at scientific notation

- **WHEN** the JSON payload contains `rate_limits.five_hour.resets_at` or `rate_limits.seven_day.resets_at` as an integer-like scientific-notation JSON number such as `1.7e9`
- **THEN** the parsed value SHALL be stored as `ResetsAt int64` in the corresponding `RateLimit` struct

#### Scenario: resets_at unconvertible

- **WHEN** a rate limit entry is present and its `resets_at` field cannot be converted to an integer timestamp
- **THEN** `ResetsAt` SHALL be zero, the rate limit entry SHALL remain present, and no countdown SHALL be displayed for that entry

#### Scenario: resets_at absent

- **WHEN** the JSON payload does not include `resets_at` for a rate limit entry
- **THEN** `ResetsAt` SHALL be zero and no countdown SHALL be displayed
