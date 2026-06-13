## MODIFIED Requirements

### Requirement: Read JSON from stdin

The program SHALL read a JSON payload from stdin, parse it using the standard library, and extract all fields equivalent to the current `jq` call in `statusline.sh`.

Recoverable field-level failures in known integer-target numeric fields SHALL NOT make the entire payload fail. These fields include `context_window.context_window_size`, `cost.total_duration_ms`, and rate-limit `resets_at` fields. When one of these fields is syntactically present but cannot be converted to an integer value, the parsed payload SHALL retain all other successfully parsed fields and SHALL expose the affected field as its zero value.

#### Scenario: Valid JSON input

- **WHEN** Claude Code pipes a valid JSON payload to the program via stdin
- **THEN** the program SHALL parse all 14 fields (model, context percentage, cost, directory, branch, rate limits, agent name, lines added/removed, duration, context size, worktree name)

#### Scenario: Integer-like decimal numeric fields

- **WHEN** stdin contains a valid JSON payload where `context_window.context_window_size`, `cost.total_duration_ms`, or rate-limit `resets_at` is encoded as an integer-like decimal JSON number such as `1000000.0` or `1700000000.0`
- **THEN** the program SHALL parse the payload successfully and SHALL expose the corresponding integer field using the numeric value represented by the JSON number

#### Scenario: Scientific notation numeric fields

- **WHEN** stdin contains a valid JSON payload where `context_window.context_window_size`, `cost.total_duration_ms`, or rate-limit `resets_at` is encoded as an integer-like scientific-notation JSON number such as `1e6` or `1.7e9`
- **THEN** the program SHALL parse the payload successfully and SHALL expose the corresponding integer field using the numeric value represented by the JSON number

#### Scenario: Unconvertible numeric field

- **WHEN** stdin contains syntactically valid JSON and one known integer-target numeric field cannot be converted to an integer value
- **THEN** the program SHALL parse the payload successfully, SHALL leave the affected field at zero value, and SHALL preserve all other successfully parsed fields for rendering

##### Example: context window size wrong scalar type

- **GIVEN** a payload with `model.display_name = "Claude Opus 4.6"`, `cost.total_cost_usd = 0.85`, and `context_window.context_window_size = "wide"`
- **WHEN** the program parses the payload
- **THEN** parsing succeeds, the parsed model and cost fields retain their payload values, and `ContextWindow.ContextWindowSize` equals `0`

#### Scenario: Invalid or empty JSON input

- **WHEN** stdin contains invalid JSON or is empty
- **THEN** the program SHALL output a single fallback line `─ │ parse error` with gray ANSI color and exit with code 0
