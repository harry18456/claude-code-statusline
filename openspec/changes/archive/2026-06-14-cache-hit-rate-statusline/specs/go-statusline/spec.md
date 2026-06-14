## MODIFIED Requirements

### Requirement: Read JSON from stdin

The program SHALL read a JSON payload from stdin, parse it using the standard library, and extract all fields required for statusline rendering.

Recoverable field-level failures in known integer-target numeric fields SHALL NOT make the entire payload fail. These fields include `context_window.context_window_size`, `context_window.current_usage.input_tokens`, `context_window.current_usage.cache_creation_input_tokens`, `context_window.current_usage.cache_read_input_tokens`, `cost.total_duration_ms`, and rate-limit `resets_at` fields. When one of these fields is syntactically present but cannot be converted to an integer value, the parsed payload SHALL retain all other successfully parsed fields and SHALL expose the affected field as its zero value.

The program SHALL parse `context_window.current_usage` as optional latest-request usage data. When `current_usage` is `null` or absent, the parsed payload SHALL expose cache hit-rate usage as unavailable rather than as a zero-token request.

#### Scenario: Valid JSON input

- **WHEN** Claude Code pipes a valid JSON payload to the program via stdin
- **THEN** the program SHALL parse the model, context percentage, context size, latest-request current usage when present, cost, directory, branch, rate limits, agent name, lines added/removed, duration, and worktree fields

#### Scenario: Current usage present

- **WHEN** stdin contains a valid JSON payload where `context_window.current_usage` contains `input_tokens`, `cache_creation_input_tokens`, and `cache_read_input_tokens`
- **THEN** the parsed payload SHALL expose current usage as available
- **THEN** the parsed payload SHALL expose each current usage token field as an integer value

##### Example: latest request token breakdown

- **GIVEN** `context_window.current_usage.input_tokens = 1`, `cache_creation_input_tokens = 1302`, and `cache_read_input_tokens = 144198`
- **WHEN** the program parses the payload
- **THEN** current usage SHALL be available with values `1`, `1302`, and `144198`

#### Scenario: Current usage null

- **WHEN** stdin contains a valid JSON payload where `context_window.current_usage` is `null`
- **THEN** the program SHALL parse the payload successfully
- **THEN** current usage SHALL be unavailable
- **THEN** all other successfully parsed fields SHALL remain available for rendering

#### Scenario: Current usage absent

- **WHEN** stdin contains a valid JSON payload without `context_window.current_usage`
- **THEN** the program SHALL parse the payload successfully
- **THEN** current usage SHALL be unavailable
- **THEN** all other successfully parsed fields SHALL remain available for rendering

#### Scenario: Integer-like decimal numeric fields

- **WHEN** stdin contains a valid JSON payload where `context_window.context_window_size`, current usage token fields, `cost.total_duration_ms`, or rate-limit `resets_at` is encoded as an integer-like decimal JSON number such as `1000000.0`, `144198.0`, or `1700000000.0`
- **THEN** the program SHALL parse the payload successfully
- **THEN** the program SHALL expose the corresponding integer field using the numeric value represented by the JSON number

#### Scenario: Scientific notation numeric fields

- **WHEN** stdin contains a valid JSON payload where `context_window.context_window_size`, current usage token fields, `cost.total_duration_ms`, or rate-limit `resets_at` is encoded as an integer-like scientific-notation JSON number such as `1e6`, `1.44198e5`, or `1.7e9`
- **THEN** the program SHALL parse the payload successfully
- **THEN** the program SHALL expose the corresponding integer field using the numeric value represented by the JSON number

#### Scenario: Unconvertible numeric field

- **WHEN** stdin contains syntactically valid JSON and one known integer-target numeric field cannot be converted to an integer value
- **THEN** the program SHALL parse the payload successfully
- **THEN** the program SHALL leave the affected field at zero value
- **THEN** the program SHALL preserve all other successfully parsed fields for rendering

##### Example: current usage token wrong scalar type

- **GIVEN** a payload with `model.display_name = "Claude Opus 4.6"`, `cost.total_cost_usd = 0.85`, `context_window.current_usage.input_tokens = "wide"`, `cache_creation_input_tokens = 1302`, and `cache_read_input_tokens = 144198`
- **WHEN** the program parses the payload
- **THEN** parsing SHALL succeed
- **THEN** the parsed model and cost fields SHALL retain their payload values
- **THEN** current usage SHALL remain available
- **THEN** `input_tokens` SHALL equal `0`, `cache_creation_input_tokens` SHALL equal `1302`, and `cache_read_input_tokens` SHALL equal `144198`

#### Scenario: Invalid or empty JSON input

- **WHEN** stdin contains invalid JSON or is empty
- **THEN** the program SHALL output a single fallback line `? ??parse error` with gray ANSI color and exit with code 0

### Requirement: Render two-line ANSI output

The program SHALL output exactly two lines of ANSI-colored text, identical in structure to the existing statusline except for an optional prompt cache hit-rate segment on line 1.

#### Scenario: Line 1 structure

- **WHEN** the program renders output without available cache hit-rate data
- **THEN** line 1 SHALL follow the existing format: `<model> <progress_bar> <pct>% <cost> <duration> <rate_limits>`

#### Scenario: Line 1 structure with cache hit rate

- **WHEN** the program renders output with available cache hit-rate data and a non-zero hit-rate denominator
- **THEN** line 1 SHALL include one additional cache hit-rate segment using the display design selected during proposal review
- **THEN** all existing line 1 segments SHALL retain their existing order relative to each other

#### Scenario: Line 2 structure

- **WHEN** the program renders output
- **THEN** line 2 SHALL follow the format: `<branch>* <+added/-removed> <dirname> <agent_or_worktree>`

#### Scenario: Zero-value hiding

- **WHEN** a field value is zero, empty, or unavailable by its field contract
- **THEN** that section SHALL be omitted from the output entirely

#### Scenario: Duration sub-minute suppression

- **WHEN** `total_duration_ms` is greater than 0 but the computed result is `0m0s`
- **THEN** the duration section SHALL be omitted

## ADDED Requirements

### Requirement: Prompt cache hit-rate display

The program SHALL compute and optionally render the latest-request prompt cache hit rate from `context_window.current_usage`.

The input-side denominator SHALL be `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`. The numerator SHALL be `cache_read_input_tokens`. `output_tokens` SHALL NOT affect the denominator or numerator. The line display SHALL use the rounded whole-number percentage for compact statusline rendering.

#### Scenario: Cache hit rate calculated from input-side current usage

- **WHEN** current usage is available with `input_tokens = 1`, `cache_creation_input_tokens = 1302`, and `cache_read_input_tokens = 144198`
- **THEN** the cache hit-rate numerator SHALL be `144198`
- **THEN** the cache hit-rate denominator SHALL be `145501`
- **THEN** the cache hit rate SHALL be approximately `99.1%`
- **THEN** the compact line display SHALL render `99%`

#### Scenario: Output tokens ignored

- **WHEN** current usage is available with `input_tokens = 10`, `cache_creation_input_tokens = 10`, `cache_read_input_tokens = 80`, and `output_tokens = 900`
- **THEN** the cache hit-rate denominator SHALL be `100`
- **THEN** the cache hit-rate numerator SHALL be `80`
- **THEN** the compact line display SHALL render `80%`

#### Scenario: Current usage unavailable

- **WHEN** current usage is unavailable because `context_window.current_usage` is `null` or absent
- **THEN** the cache hit-rate segment SHALL be omitted from line 1
- **THEN** rendering SHALL continue for all other line 1 and line 2 segments

#### Scenario: Zero denominator

- **WHEN** current usage is available and `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` equals `0`
- **THEN** the cache hit-rate segment SHALL be omitted from line 1
- **THEN** rendering SHALL NOT divide by zero

#### Scenario: Default or Nerd Font cache hit text

- **WHEN** current usage is available, the denominator is non-zero, and ASCII mode is disabled
- **THEN** the cache hit-rate segment SHALL use the reviewed non-ASCII cache indicator followed by the rounded whole-number percentage, for example `⚡99%`

#### Scenario: ASCII cache hit text

- **WHEN** current usage is available, the denominator is non-zero, and `CLAUDE_STATUSLINE_ASCII=1` is set
- **THEN** the cache hit-rate segment SHALL use the ASCII fallback format `cache:<pct>%`
- **THEN** the cache hit-rate segment SHALL contain no ANSI color escapes
- **THEN** the cache hit-rate segment SHALL contain no Unicode glyphs
