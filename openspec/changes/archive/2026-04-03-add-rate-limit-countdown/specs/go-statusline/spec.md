## MODIFIED Requirements

### Requirement: Rate limit display

The program SHALL display 5-hour and 7-day rate limits when present, colored red when >= 80%. When `resets_at` is present, the program SHALL always append a countdown to reset after the percentage, using the format `(Xd Yh)`, `(Xh Ym)`, `(Ym)`, or `(now)` depending on remaining time.

#### Scenario: Rate limit >= 80%

- **WHEN** a rate limit used_percentage is >= 80
- **THEN** it SHALL be displayed in red

#### Scenario: Rate limit with resets_at (any percentage)

- **WHEN** `resets_at` is non-zero for a rate limit entry
- **THEN** the display SHALL include the countdown regardless of used_percentage, e.g. `5h:15% (1h 23m)` or `7d:82% (2d 3h)`

#### Scenario: Rate limit < 80%

- **WHEN** a rate limit used_percentage is >= 0 and < 80
- **THEN** it SHALL be displayed in gray

#### Scenario: Rate limit absent

- **WHEN** a rate limit field is absent in the JSON
- **THEN** it SHALL be omitted from the output
