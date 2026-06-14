## ADDED Requirements

### Requirement: Command-line configuration

The program SHALL parse command-line configuration before reading stdin. Project-specific display configuration SHALL be controlled by command-line flags, not by `CLAUDE_STATUSLINE_*` environment variables. CLI configuration problems SHALL NOT blank the statusline: except for `--version`, the program SHALL continue to read stdin, render a valid statusline to stdout, write any configuration warnings to stderr, and exit with code 0.

The program SHALL support these flags:

| Flag | Effect |
| ---- | ------ |
| `--ascii` | Enable pure ASCII rendering. |
| `--nerdfont` | Enable Nerd Font symbols and imply Powerline separators. |
| `--powerline` | Enable Powerline separators. |
| `--hide <keys>` | Hide comma-separated renderer sections. |
| `--version` | Print the binary version and exit. |

The program SHALL continue to use `COLORTERM=truecolor` and `COLORTERM=24bit` as terminal capability signals for true color rendering.

#### Scenario: Version flag

- **WHEN** the program is invoked with `--version`
- **THEN** the program SHALL print the build version to stdout
- **THEN** the program SHALL exit with code 0
- **THEN** the program SHALL NOT read stdin or render a statusline

#### Scenario: ASCII flag

- **WHEN** the program is invoked with `--ascii`
- **THEN** the renderer options SHALL enable ASCII mode
- **THEN** ASCII rendering SHALL use the same ASCII symbols and separators as the previous ASCII mode

#### Scenario: Nerd Font flag

- **WHEN** the program is invoked with `--nerdfont`
- **THEN** the renderer options SHALL enable Nerd Font mode
- **THEN** the renderer options SHALL enable Powerline separators

#### Scenario: Powerline flag

- **WHEN** the program is invoked with `--powerline` and without `--nerdfont`
- **THEN** the renderer options SHALL enable Powerline separators
- **THEN** the renderer options SHALL NOT enable Nerd Font mode

#### Scenario: Removed project-specific environment variables

- **WHEN** `CLAUDE_STATUSLINE_ASCII=1`, `CLAUDE_STATUSLINE_NERDFONT=1`, or `CLAUDE_STATUSLINE_POWERLINE=1` is present in the environment
- **THEN** those environment variables SHALL NOT enable ASCII mode, Nerd Font mode, or Powerline separators
- **THEN** only command-line flags SHALL control those project-specific display modes

#### Scenario: True color detection remains environment-based

- **WHEN** `COLORTERM` is `truecolor` or `24bit`
- **THEN** the renderer options SHALL enable true color rendering

#### Scenario: Invalid display mode combination

- **WHEN** the program is invoked with `--ascii` together with `--nerdfont` or `--powerline`
- **THEN** the program SHALL enable ASCII mode and ignore Nerd Font or Powerline rendering for that invocation
- **THEN** the program SHALL write a concise warning to stderr naming the conflict
- **THEN** the program SHALL render a statusline to stdout and exit with code 0

#### Scenario: Invalid CLI syntax

- **WHEN** the program is invoked with an unknown flag, a missing `--hide` value, or a non-empty positional argument
- **THEN** the program SHALL write a concise warning to stderr
- **THEN** the program SHALL continue rendering with default options or any successfully parsed safe options
- **THEN** the program SHALL render a statusline to stdout and exit with code 0

### Requirement: Configurable section visibility

The program SHALL support `--hide <keys>` for explicit renderer section hiding. The default hidden set SHALL be empty. Existing zero-value and unavailable-data suppression SHALL continue to apply.

The only valid hide keys SHALL be:

| Key | Section |
| --- | ------- |
| `model` | Line 1 brand/model segment, including the brand symbol and model display name. |
| `effort` | Line 1 execution-mode segment. |
| `bar` | Line 1 context progress bar, percentage, and high-usage warning symbol. |
| `size` | Line 1 context size label such as `200k` or `1M`. |
| `cost` | Line 1 total cost display. |
| `cache` | Line 1 prompt cache hit-rate display. |
| `duration` | Line 1 duration display. |
| `rate` | Line 1 five-hour and seven-day rate-limit display, including countdown and pace indicator. |
| `branch` | Line 2 branch display and dirty marker. |
| `lines` | Line 2 `+added/-removed` display. |
| `dir` | Line 2 directory display. |
| `agent` | Line 2 agent or worktree indicator. |

Hide keys SHALL be lowercase and canonical. The program SHALL trim whitespace around comma-separated keys, ignore empty tokens, and de-duplicate repeated keys. Multiple `--hide` flags SHALL merge into a single hidden set. Unknown non-empty hide keys SHALL be ignored with a stderr warning; known keys from the same `--hide` value SHALL still apply.

#### Scenario: Default visibility

- **WHEN** the program is invoked without `--hide`
- **THEN** every section with available data SHALL render according to its existing requirement

#### Scenario: Hide line 1 sections

- **WHEN** the program is invoked with `--hide model,effort,bar,size,cost,cache,duration,rate`
- **THEN** each listed line 1 section SHALL be omitted even when the payload contains data that would normally render it
- **THEN** unlisted line 2 sections SHALL remain eligible for rendering

#### Scenario: Hide line 2 sections

- **WHEN** the program is invoked with `--hide branch,lines,dir,agent`
- **THEN** each listed line 2 section SHALL be omitted even when the payload contains data that would normally render it
- **THEN** unlisted line 1 sections SHALL remain eligible for rendering

#### Scenario: Cost and cache hide independently

- **WHEN** cost and prompt cache hit-rate data are both available
- **THEN** `--hide cost` SHALL omit the cost text while leaving the cache hit-rate segment eligible for rendering
- **THEN** `--hide cache` SHALL omit the cache hit-rate text while leaving the cost segment eligible for rendering
- **THEN** `--hide cost,cache` SHALL omit both cost and cache hit-rate text

#### Scenario: Hidden sections do not leave separators

- **WHEN** adjacent sections are hidden or unavailable
- **THEN** the rendered line SHALL NOT contain leading separators, trailing separators, or duplicate separators created by those hidden sections

##### Example: adjacent line 1 sections hidden

- **GIVEN** line 1 would normally contain `model`, `effort`, `bar`, `size`, `cost`, `cache`, `duration`, and `rate`
- **WHEN** the program is invoked with `--hide cost,cache,duration`
- **THEN** the rendered line 1 SHALL contain `model`, `effort`, `bar`, `size`, and `rate` in order with single separators between visible sections only

#### Scenario: All sections hidden on a line

- **WHEN** every section on a line is hidden or unavailable
- **THEN** that line SHALL render as an empty string
- **THEN** the program SHALL still emit the two-line statusline protocol

##### Example: all line 2 sections hidden

- **GIVEN** line 2 would normally contain `branch`, `lines`, `dir`, and `agent`
- **WHEN** the program is invoked with `--hide branch,lines,dir,agent`
- **THEN** line 2 SHALL be an empty string

#### Scenario: Unknown hide key

- **WHEN** `--hide` contains a non-empty key outside the canonical key set, such as `rate_limits`, `directory`, or `worktree`
- **THEN** the program SHALL ignore the unknown key
- **THEN** the program SHALL apply any known keys from the same `--hide` value
- **THEN** the program SHALL write a concise warning naming the invalid key to stderr
- **THEN** the program SHALL render a statusline to stdout and exit with code 0

##### Example: mixed known and unknown hide keys

- **GIVEN** line 1 would normally contain `effort`, `duration`, and `rate`
- **WHEN** the program is invoked with `--hide effort,rate_limits,duration`
- **THEN** the `effort` and `duration` sections SHALL be omitted
- **THEN** the unknown `rate_limits` key SHALL be ignored and SHALL NOT suppress the `rate` section

## MODIFIED Requirements

### Requirement: Render two-line ANSI output

The program SHALL output exactly two lines of ANSI-colored text. Visible sections SHALL retain the existing order: line 1 uses model, execution mode, context bar, context size label, cost/cache, duration, and rate limits; line 2 uses branch, lines added/removed, directory, and agent/worktree.

Sections hidden by `--hide` SHALL be omitted after their normal data-availability rules are evaluated. Hidden or unavailable sections SHALL NOT emit separators.

#### Scenario: Line 1 structure

- **WHEN** the program renders output without hidden line 1 sections
- **THEN** line 1 SHALL follow the existing visible-section order: `<model> <execution_mode> <progress_bar> <pct>% <context_size> <cost/cache> <duration> <rate_limits>`

#### Scenario: Line 1 structure with cache hit rate

- **WHEN** the program renders output with available cache hit-rate data, a non-zero hit-rate denominator, and the `cache` section is not hidden
- **THEN** line 1 SHALL include the cache hit-rate display using the existing cache display design
- **THEN** all other visible line 1 sections SHALL retain their existing order relative to each other

#### Scenario: Line 1 structure with execution mode

- **WHEN** the program renders output with usable execution-mode data and the `effort` section is not hidden
- **THEN** line 1 SHALL include the execution-mode segment using the existing execution-mode display design
- **THEN** all other visible line 1 sections SHALL retain their existing order relative to each other

#### Scenario: Line 2 structure

- **WHEN** the program renders output without hidden line 2 sections
- **THEN** line 2 SHALL follow the existing visible-section order: `<branch>* <+added/-removed> <dirname> <agent_or_worktree>`

#### Scenario: Zero-value hiding

- **WHEN** a field value is zero, empty, or unavailable by its field contract
- **THEN** that section SHALL be omitted from the output entirely

##### Example: unavailable duration and rate limits

- **GIVEN** `total_duration_ms = 0` and rate-limit fields are absent
- **WHEN** the program renders output without `--hide`
- **THEN** the duration and rate sections SHALL be omitted

#### Scenario: Duration sub-minute suppression

- **WHEN** `total_duration_ms` is greater than 0 but the computed result is `0m0s`
- **THEN** the duration section SHALL be omitted

### Requirement: Three-tier color rendering

The program SHALL support three rendering tiers based on `--ascii` and terminal true color detection.

#### Scenario: True color mode

- **WHEN** `COLORTERM` environment variable is `truecolor` or `24bit`
- **THEN** the progress bar SHALL render with per-cell RGB gradient

#### Scenario: ANSI fallback mode

- **WHEN** `COLORTERM` is not set to `truecolor` or `24bit`
- **THEN** the progress bar SHALL render with solid ANSI color based on overall percentage

#### Scenario: ASCII mode

- **WHEN** `--ascii` is set
- **THEN** the progress bar SHALL use `#` for filled cells and `-` for empty cells, with no Unicode or ANSI color codes

### Requirement: Nerd Font and Powerline support

The program SHALL support optional Nerd Font icons and Powerline separators through command-line flags.

#### Scenario: Nerd Font mode

- **WHEN** `--nerdfont` is set
- **THEN** the program SHALL use Nerd Font icons for model, time, and cost symbols
- **THEN** the program SHALL use Powerline separators

#### Scenario: Powerline separators

- **WHEN** `--powerline` is set
- **THEN** the program SHALL use Powerline arrow separators instead of `│`

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

- **WHEN** current usage is available, the denominator is non-zero, ASCII mode is disabled, and the `cache` section is not hidden
- **THEN** the cache hit-rate segment SHALL use the existing non-ASCII cache indicator followed by the rounded whole-number percentage, for example `⚡99%`

#### Scenario: ASCII cache hit text

- **WHEN** current usage is available, the denominator is non-zero, `--ascii` is set, and the `cache` section is not hidden
- **THEN** the cache hit-rate segment SHALL use the ASCII fallback format `cache:<pct>%`
- **THEN** the cache hit-rate segment SHALL contain no ANSI color escapes
- **THEN** the cache hit-rate segment SHALL contain no Unicode glyphs

### Requirement: Execution mode display

The program SHALL render an execution-mode segment on line 1 when usable execution-mode data is available and the `effort` section is not hidden. Effort is the primary execution-mode signal. The renderer SHALL keep execution-mode formatting isolated so parsing availability, effort validation, text formatting, color selection, and ASCII/Nerd Font behavior remain independently testable.

The execution-mode formatter SHALL return an empty string when effort is unavailable, when the effort level is empty, or when the effort level is not one of `low`, `medium`, `high`, `xhigh`, or `max`. The formatter MUST NOT panic when thinking or fast-mode data is absent, null, false, or malformed.

#### Scenario: Effort available

- **WHEN** the renderer receives parsed execution-mode data with `effort.level = "max"` and the `effort` section is not hidden
- **THEN** the execution-mode formatter SHALL return a non-empty segment containing the level `max` using the existing execution-mode display format

#### Scenario: Effort absent

- **WHEN** the renderer receives parsed execution-mode data where effort is unavailable
- **THEN** the execution-mode formatter SHALL return an empty string
- **THEN** line 1 SHALL omit the execution-mode segment
- **THEN** all existing line 1 and line 2 segments SHALL retain their current behavior

#### Scenario: Effort hidden

- **WHEN** the renderer receives usable execution-mode data and the `effort` section is hidden
- **THEN** line 1 SHALL omit the execution-mode segment
- **THEN** all unhidden line 1 and line 2 segments SHALL retain their current behavior

#### Scenario: Unknown effort level

- **WHEN** the renderer receives parsed execution-mode data with an effort level outside `low`, `medium`, `high`, `xhigh`, or `max`
- **THEN** the execution-mode formatter SHALL return an empty string
- **THEN** line 1 SHALL omit the execution-mode segment

#### Scenario: ASCII mode execution mode

- **WHEN** execution-mode data is usable, `--ascii` is set, and the `effort` section is not hidden
- **THEN** the execution-mode segment SHALL use the existing ASCII fallback text
- **THEN** the execution-mode segment SHALL contain no ANSI color escapes
- **THEN** the execution-mode segment SHALL contain no Unicode glyphs

#### Scenario: Default mode execution mode

- **WHEN** execution-mode data is usable, ASCII mode is disabled, and the `effort` section is not hidden
- **THEN** the execution-mode segment SHALL use the existing default text or glyph format
- **THEN** the effort level SHALL be colored according to the existing low-to-max intensity scale

#### Scenario: Nerd Font mode execution mode

- **WHEN** execution-mode data is usable, `--nerdfont` is set, and the `effort` section is not hidden
- **THEN** the execution-mode segment SHALL use the existing Nerd Font compatible text or glyph format
- **THEN** the effort level SHALL be colored according to the same low-to-max intensity scale as default mode

#### Scenario: Thinking and fast mode signals

- **WHEN** parsed `thinking.enabled` or `fast_mode` data is available and the `effort` section is not hidden
- **THEN** the renderer SHALL include or suppress those signals according to the existing execution-mode display design
- **THEN** absent thinking or fast-mode data SHALL NOT produce placeholder text
