## ADDED Requirements

### Requirement: Parse execution mode fields

The program SHALL parse Claude Code execution-mode fields from the top-level JSON payload when they are present. The parsed payload SHALL expose `effort.level`, `thinking.enabled`, and `fast_mode` with field availability information so renderers can distinguish unavailable data from explicit boolean false values.

Absent and `null` execution-mode fields SHALL be represented as unavailable. Field-level malformed values in execution-mode fields MUST NOT make the entire payload fail; the affected execution-mode value SHALL be treated as unavailable or invalid while all other successfully parsed payload fields remain available.

#### Scenario: Execution mode fields present

- **WHEN** stdin contains a valid JSON payload with `effort.level = "max"`, `thinking.enabled = true`, and `fast_mode = false`
- **THEN** the parsed payload SHALL expose effort as available with level `max`
- **THEN** the parsed payload SHALL expose thinking as available with enabled value `true`
- **THEN** the parsed payload SHALL expose fast mode as available with value `false`

#### Scenario: Execution mode fields absent

- **WHEN** stdin contains a valid JSON payload that omits `effort`, `thinking`, and `fast_mode`
- **THEN** the parsed payload SHALL expose effort as unavailable
- **THEN** the parsed payload SHALL expose thinking as unavailable
- **THEN** the parsed payload SHALL expose fast mode as unavailable
- **THEN** all other successfully parsed fields SHALL remain available for rendering

#### Scenario: Execution mode fields null

- **WHEN** stdin contains a valid JSON payload with `effort = null`, `thinking = null`, and `fast_mode = null`
- **THEN** the parsed payload SHALL expose effort as unavailable
- **THEN** the parsed payload SHALL expose thinking as unavailable
- **THEN** the parsed payload SHALL expose fast mode as unavailable
- **THEN** parsing SHALL succeed

#### Scenario: Malformed execution mode fields

- **WHEN** stdin contains syntactically valid JSON where one execution-mode field has the wrong scalar or object shape
- **THEN** parsing SHALL succeed
- **THEN** the affected execution-mode value SHALL NOT be used for rendering
- **THEN** the parsed model, context, cost, workspace, worktree, agent, and rate-limit fields SHALL retain their successfully parsed values

#### Scenario: Known effort levels

- **WHEN** stdin contains a valid JSON payload with `effort.level` equal to `low`, `medium`, `high`, `xhigh`, or `max`
- **THEN** the parsed payload SHALL expose the exact effort level string for renderer use

##### Example: supported effort level values

| JSON `effort.level` | Parsed effort level |
| ------------------- | ------------------- |
| `low`               | `low`               |
| `medium`            | `medium`            |
| `high`              | `high`              |
| `xhigh`             | `xhigh`             |
| `max`               | `max`               |

### Requirement: Execution mode display

The program SHALL render an optional execution-mode segment on line 1 when usable execution-mode data is available. Effort is the primary execution-mode signal and SHALL be the default displayed value after the review-approved display design is selected. The renderer SHALL provide an isolated formatter, such as `formatEffort` or `formatExecutionMode`, so parsing availability, effort validation, text formatting, color selection, and ASCII/Nerd Font behavior can be tested independently.

The execution-mode formatter SHALL return an empty string when effort is unavailable, when the effort level is empty, or when the effort level is not one of `low`, `medium`, `high`, `xhigh`, or `max`. The formatter MUST NOT panic when thinking or fast-mode data is absent, null, false, or malformed.

#### Scenario: Effort available

- **WHEN** the renderer receives parsed execution-mode data with `effort.level = "max"`
- **THEN** the execution-mode formatter SHALL return a non-empty segment containing the level `max` using the review-approved display format

#### Scenario: Effort absent

- **WHEN** the renderer receives parsed execution-mode data where effort is unavailable
- **THEN** the execution-mode formatter SHALL return an empty string
- **THEN** line 1 SHALL omit the execution-mode segment
- **THEN** all existing line 1 and line 2 segments SHALL retain their current behavior

#### Scenario: Unknown effort level

- **WHEN** the renderer receives parsed execution-mode data with an effort level outside `low`, `medium`, `high`, `xhigh`, or `max`
- **THEN** the execution-mode formatter SHALL return an empty string
- **THEN** line 1 SHALL omit the execution-mode segment

#### Scenario: ASCII mode execution mode

- **WHEN** execution-mode data is usable and `CLAUDE_STATUSLINE_ASCII=1` is active
- **THEN** the execution-mode segment SHALL use the review-approved ASCII fallback text
- **THEN** the execution-mode segment SHALL contain no ANSI color escapes
- **THEN** the execution-mode segment SHALL contain no Unicode glyphs

#### Scenario: Default mode execution mode

- **WHEN** execution-mode data is usable and ASCII mode is disabled
- **THEN** the execution-mode segment SHALL use the review-approved default text or glyph format
- **THEN** the effort level SHALL be colored according to the review-approved low-to-max intensity scale

#### Scenario: Nerd Font mode execution mode

- **WHEN** execution-mode data is usable and `CLAUDE_STATUSLINE_NERDFONT=1` is active
- **THEN** the execution-mode segment SHALL use the review-approved Nerd Font compatible text or glyph format
- **THEN** the effort level SHALL be colored according to the same low-to-max intensity scale as default mode

#### Scenario: Thinking and fast mode signals

- **WHEN** parsed `thinking.enabled` or `fast_mode` data is available
- **THEN** the renderer SHALL include or suppress those signals according to the review-approved display design
- **THEN** absent thinking or fast-mode data SHALL NOT produce placeholder text

## MODIFIED Requirements

### Requirement: Render two-line ANSI output

The program SHALL output exactly two lines of ANSI-colored text, identical in structure to the existing statusline except for optional prompt cache hit-rate and execution-mode segments on line 1.

#### Scenario: Line 1 structure

- **WHEN** the program renders output without available cache hit-rate data and without available execution-mode data
- **THEN** line 1 SHALL follow the existing format: `<model> <progress_bar> <pct>% <cost> <duration> <rate_limits>`

#### Scenario: Line 1 structure with cache hit rate

- **WHEN** the program renders output with available cache hit-rate data and a non-zero hit-rate denominator
- **THEN** line 1 SHALL include one additional cache hit-rate segment using the reviewed cache display design
- **THEN** all existing line 1 segments SHALL retain their existing order relative to each other

#### Scenario: Line 1 structure with execution mode

- **WHEN** the program renders output with usable execution-mode data
- **THEN** line 1 SHALL include one additional execution-mode segment using the review-approved execution-mode display design
- **THEN** all existing line 1 segments SHALL retain their existing order relative to each other except for the approved insertion point of the execution-mode segment

#### Scenario: Line 2 structure

- **WHEN** the program renders output
- **THEN** line 2 SHALL follow the format: `<branch>* <+added/-removed> <dirname> <agent_or_worktree>`

#### Scenario: Zero-value hiding

- **WHEN** a field value is zero, empty, or unavailable by its field contract
- **THEN** that section SHALL be omitted from the output entirely

#### Scenario: Duration sub-minute suppression

- **WHEN** `total_duration_ms` is greater than 0 but the computed result is `0m0s`
- **THEN** the duration section SHALL be omitted
