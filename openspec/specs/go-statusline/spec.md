## ADDED Requirements

### Requirement: Read JSON from stdin

The program SHALL read a JSON payload from stdin, parse it using the standard library, and extract all fields equivalent to the current `jq` call in `statusline.sh`.

#### Scenario: Valid JSON input

- **WHEN** Claude Code pipes a valid JSON payload to the program via stdin
- **THEN** the program SHALL parse all 14 fields (model, context percentage, cost, directory, branch, rate limits, agent name, lines added/removed, duration, context size, worktree name)

#### Scenario: Invalid or empty JSON input

- **WHEN** stdin contains invalid JSON or is empty
- **THEN** the program SHALL output a single fallback line `─ │ parse error` with gray ANSI color and exit with code 0


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->


<!-- @trace
source: tolerant-payload-parsing
updated: 2026-06-13
code:
  - docs/improvement-plan.md
  - internal/model/payload.go
tests:
  - internal/model/payload_test.go
-->


<!-- @trace
source: cache-hit-rate-statusline
updated: 2026-06-14
code:
  - docs/improvement-plan.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
tests:
  - internal/gitcache/gitcache_test.go
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Render two-line ANSI output

The program SHALL output exactly two lines of ANSI-colored text, identical in structure to the bash version.

#### Scenario: Line 1 structure

- **WHEN** the program renders output
- **THEN** line 1 SHALL follow the format: `◆ <model> │ <progress_bar> <pct>% │ $<cost> │ <duration> │ <rate_limits>`

#### Scenario: Line 2 structure

- **WHEN** the program renders output
- **THEN** line 2 SHALL follow the format: `⎇ <branch>* │ +<added>/-<removed> │ <dirname> │ ⚙ <agent_or_worktree>`

#### Scenario: Zero-value hiding

- **WHEN** a field value is zero or empty (duration = 0, lines added/removed = 0, rate limits absent)
- **THEN** that section SHALL be omitted from the output entirely

#### Scenario: Duration sub-minute suppression

- **WHEN** `total_duration_ms` is greater than 0 but the computed result is `0m0s` (i.e., less than 1 second after conversion)
- **THEN** the duration section SHALL still be omitted (only display when at least 1 second has elapsed)


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->


<!-- @trace
source: cache-hit-rate-statusline
updated: 2026-06-14
code:
  - docs/improvement-plan.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
tests:
  - internal/gitcache/gitcache_test.go
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->


<!-- @trace
source: show-execution-mode-statusline
updated: 2026-06-14
code:
  - internal/renderer/renderer.go
  - internal/model/payload.go
tests:
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Three-tier color rendering

The program SHALL support three rendering tiers based on environment variables.

#### Scenario: True color mode

- **WHEN** `COLORTERM` environment variable is `truecolor` or `24bit`
- **THEN** the progress bar SHALL render with per-cell RGB gradient (green → yellow → red)

#### Scenario: ANSI fallback mode

- **WHEN** `COLORTERM` is not set to truecolor values
- **THEN** the progress bar SHALL render with solid ANSI color (green/yellow/red) based on overall percentage

#### Scenario: ASCII mode

- **WHEN** `CLAUDE_STATUSLINE_ASCII=1` is set
- **THEN** the progress bar SHALL use `#` for filled cells and `-` for empty cells, with no Unicode or ANSI color codes


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Nerd Font and Powerline support

The program SHALL support optional Nerd Font icons and Powerline separators.

#### Scenario: Nerd Font mode

- **WHEN** `CLAUDE_STATUSLINE_NERDFONT=1` is set
- **THEN** the program SHALL use Nerd Font icons for model, time, and cost symbols

#### Scenario: Powerline separators

- **WHEN** `CLAUDE_STATUSLINE_POWERLINE=1` is set (or implied by NERDFONT=1)
- **THEN** the program SHALL use Powerline arrow separators instead of `│`


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Git branch with dirty-check caching

The program SHALL display the current git branch with a dirty marker, using a 5-second file-based cache.

#### Scenario: Cache hit

- **WHEN** the cache file exists and is less than 5 seconds old
- **THEN** the program SHALL read branch and dirty state from cache without running git

#### Scenario: Cache miss

- **WHEN** the cache file is absent or older than 5 seconds
- **THEN** the program SHALL run `git` to determine branch and dirty state, then write the result to cache

#### Scenario: Cache file location

- **WHEN** writing or reading the cache
- **THEN** the program SHALL use `os.TempDir()` to determine the temp directory path (cross-platform)

#### Scenario: Branch name fallback to short SHA

- **WHEN** `git branch --show-current` returns empty (e.g., detached HEAD state)
- **THEN** the program SHALL fall back to `git rev-parse --short HEAD` as the branch display value

#### Scenario: Git subprocess noise suppression

- **WHEN** running any git command for dirty-check or branch detection
- **THEN** the program SHALL pass `-c core.useBuiltinFSMonitor=false` to suppress fsmonitor noise


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->


<!-- @trace
source: isolate-gitcache-atomic-writes
updated: 2026-06-14
code:
  - cmd/statusline/main.go
  - internal/gitcache/gitcache.go
  - docs/improvement-plan.md
tests:
  - internal/gitcache/gitcache_test.go
-->

### Requirement: Context window size label

The program SHALL display a context window size label based solely on `context_window_size`. When `context_window_size >= 1,000,000` and the payload indicates usage has crossed the 200k token pricing threshold, the `1M` label SHALL be colored red to warn of elevated per-token pricing. The model name SHALL NOT affect whether the label is displayed; the program MUST NOT suppress the label based on substrings of the model name.

#### Scenario: 1M context window within base pricing tier

- **WHEN** context_window_size >= 1,000,000 AND `exceeds_200k_tokens` is false
- **THEN** the program SHALL display ` 1M` in gray after the percentage

#### Scenario: 1M context window in premium pricing tier

- **WHEN** context_window_size >= 1,000,000 AND `exceeds_200k_tokens` is true
- **THEN** the program SHALL display ` 1M` in red after the percentage

#### Scenario: 200k context window

- **WHEN** context_window_size >= 200,000 AND context_window_size < 1,000,000
- **THEN** the program SHALL display ` 200k` in gray after the percentage

#### Scenario: Context window size below 200k

- **WHEN** context_window_size < 200,000 (including zero when the field is absent)
- **THEN** the program SHALL NOT display any context window size label

#### Scenario: Model name containing the substring "context" does not suppress the label

- **WHEN** context_window_size >= 1,000,000 AND the model `display_name` contains the substring "context" (case-insensitive), for example `Opus 4.7 (1M context)`
- **THEN** the program SHALL still display ` 1M` after the percentage, even though the label text duplicates information already present in the model name


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->


<!-- @trace
source: additional-statusline-indicators
updated: 2026-04-20
code:
  - internal/renderer/renderer.go
  - internal/model/payload.go
tests:
  - internal/renderer/renderer_test.go
  - internal/model/payload_test.go
-->


<!-- @trace
source: context-label-drop-name-rule
updated: 2026-04-22
code:
  - README.zh-TW.md
  - internal/renderer/renderer.go
  - README.md
tests:
  - internal/renderer/renderer_test.go
-->

### Requirement: Cost color thresholds

The program SHALL color the cost display based on value thresholds.

#### Scenario: Cost >= $10

- **WHEN** total cost >= $10.00
- **THEN** the cost SHALL be displayed in red

#### Scenario: Cost >= $5 and < $10

- **WHEN** total cost >= $5.00 and < $10.00
- **THEN** the cost SHALL be displayed in yellow

#### Scenario: Cost > $0.00 and < $5.00

- **WHEN** total cost is greater than $0.00 and less than $5.00
- **THEN** the cost SHALL be displayed in yellow

#### Scenario: Cost = $0.00

- **WHEN** total cost is exactly $0.00
- **THEN** the cost SHALL be displayed in gray (dimmed) and SHALL still be shown (not hidden)


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Rate limit display

The program SHALL display 5-hour and 7-day rate limits when present, colored red when >= 80%.

#### Scenario: Rate limit >= 80%

- **WHEN** a rate limit used_percentage is >= 80
- **THEN** it SHALL be displayed in red

#### Scenario: Rate limit < 80%

- **WHEN** a rate limit used_percentage is >= 0 and < 80
- **THEN** it SHALL be displayed in gray

#### Scenario: Rate limit absent

- **WHEN** a rate limit field is -1 or absent in the JSON
- **THEN** it SHALL be omitted from the output


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->


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

### Requirement: Agent and Worktree indicator

The program SHALL display an agent or worktree indicator when active.

#### Scenario: Worktree active

- **WHEN** `worktree.name` is non-empty in the JSON
- **THEN** line 2 SHALL append `⚙ worktree:<name>` in yellow

#### Scenario: Subagent active

- **WHEN** `agent.name` is non-empty and `worktree.name` is empty
- **THEN** line 2 SHALL append `⚙ <agent_name>` in yellow


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Warning symbol at high context usage

The program SHALL display a warning symbol when context usage reaches a critical level.

#### Scenario: Warning symbol shown

- **WHEN** context usage percentage is >= 90
- **THEN** the warning symbol (`⚠` in default/nerdfont mode, `!` in ASCII mode) SHALL be displayed in red immediately after the percentage value on line 1

#### Scenario: Warning symbol absent

- **WHEN** context usage percentage is < 90
- **THEN** no warning symbol SHALL be displayed


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Progress bar dimensions and clamping

The program SHALL render a fixed-width progress bar with bounded input.

#### Scenario: Progress bar width

- **WHEN** rendering the progress bar in any mode
- **THEN** the bar SHALL always be exactly 10 characters wide (10 cells filled with `█`/`░` or `#`/`-`)

#### Scenario: Percentage clamping

- **WHEN** `context_window.used_percentage` is less than 0 or greater than 100
- **THEN** the value SHALL be clamped to the range [0, 100] before rendering


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Section text colors

The program SHALL apply consistent ANSI colors to each section of the output.

#### Scenario: Brand diamond color

- **WHEN** rendering line 1 in non-ASCII mode with true color support
- **THEN** the `◆` symbol SHALL be rendered in Anthropic brand purple (RGB 114, 102, 234)

#### Scenario: Model name color

- **WHEN** rendering the model name on line 1
- **THEN** the model name SHALL be displayed in cyan

#### Scenario: Git branch color

- **WHEN** rendering the git branch on line 2
- **THEN** the branch name (including dirty marker `*`) SHALL be displayed in gray

#### Scenario: Directory color

- **WHEN** rendering the current directory name on line 2
- **THEN** the directory name SHALL be displayed in blue


<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

### Requirement: Lines added and removed colors

The program SHALL display code change counts with directional colors.

#### Scenario: Lines added in green, removed in red

- **WHEN** `total_lines_added` or `total_lines_removed` is greater than 0
- **THEN** the added count SHALL be prefixed with `+` in green, and the removed count SHALL be prefixed with `-` in red, formatted as `+N/-N`

## Requirements

<!-- @trace
source: rewrite-to-go
updated: 2026-03-26
code:
  - cmd/statusline/main.go
  - go.mod
  - README.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - README.zh-TW.md
  - install.sh
  - CLAUDE.md
  - .github/workflows/release.yml
  - internal/renderer/renderer.go
  - statusline.sh
  - examples/test-mock.sh
tests:
  - internal/model/payload_test.go
  - internal/gitcache/gitcache_test.go
  - internal/renderer/renderer_test.go
-->

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

---
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

---
### Requirement: Three-tier color rendering

The program SHALL support three rendering tiers based on environment variables.

#### Scenario: True color mode

- **WHEN** `COLORTERM` environment variable is `truecolor` or `24bit`
- **THEN** the progress bar SHALL render with per-cell RGB gradient (green → yellow → red)

#### Scenario: ANSI fallback mode

- **WHEN** `COLORTERM` is not set to truecolor values
- **THEN** the progress bar SHALL render with solid ANSI color (green/yellow/red) based on overall percentage

#### Scenario: ASCII mode

- **WHEN** `CLAUDE_STATUSLINE_ASCII=1` is set
- **THEN** the progress bar SHALL use `#` for filled cells and `-` for empty cells, with no Unicode or ANSI color codes

---
### Requirement: Nerd Font and Powerline support

The program SHALL support optional Nerd Font icons and Powerline separators.

#### Scenario: Nerd Font mode

- **WHEN** `CLAUDE_STATUSLINE_NERDFONT=1` is set
- **THEN** the program SHALL use Nerd Font icons for model, time, and cost symbols

#### Scenario: Powerline separators

- **WHEN** `CLAUDE_STATUSLINE_POWERLINE=1` is set (or implied by NERDFONT=1)
- **THEN** the program SHALL use Powerline arrow separators instead of `│`

---
### Requirement: Git branch with dirty-check caching

The program SHALL display the current git branch with a dirty marker, using a 5-second file-based cache. The cache file for a working directory SHALL be derived from that working directory, and cache writes SHALL publish complete cache payloads atomically.

#### Scenario: Cache hit

- **WHEN** the cache file derived from the requested working directory exists and is less than 5 seconds old
- **THEN** the program SHALL read branch and dirty state from that working directory's cache without running git

#### Scenario: Cache miss

- **WHEN** the cache file derived from the requested working directory is absent or older than 5 seconds
- **THEN** the program SHALL run `git` in the requested working directory to determine branch and dirty state, then write the result to that working directory's cache

#### Scenario: Cache file location

- **WHEN** deriving the cache file path for a working directory
- **THEN** the program SHALL use `os.TempDir()` to determine the temp directory path
- **THEN** the cache filename SHALL contain a deterministic SHA-256 hash prefix derived from a normalized form of the working directory path

#### Scenario: Cache isolation across working directories

- **WHEN** two different working directories have fresh git cache entries with different branch or dirty values
- **THEN** reading git status for one working directory SHALL NOT return the cached branch or dirty value from the other working directory

##### Example: two repositories within the TTL

- **GIVEN** repository `A` has branch `main` and repository `B` has branch `feature/cache`
- **WHEN** both repositories refresh their git caches within 5 seconds and repository `A` reads from cache again
- **THEN** repository `A` returns branch `main`, not `feature/cache`

#### Scenario: Atomic cache write

- **WHEN** writing branch and dirty state to a cache file
- **THEN** the program MUST write the complete `branch|dirty_flag` payload to a temp file created in the same directory as the final cache file
- **THEN** the program MUST publish the complete temp file with `os.Rename`
- **THEN** the program MUST remove the temp file if write, close, or rename fails before a successful rename

#### Scenario: Concurrent cache refresh

- **WHEN** multiple statusline executions refresh the cache for the same working directory concurrently
- **THEN** every successful cache read SHALL parse a complete `branch|dirty_flag` payload
- **THEN** the branch returned for the working directory SHALL match the branch detected for that working directory

#### Scenario: Branch name fallback to short SHA

- **WHEN** `git branch --show-current` returns empty (e.g., detached HEAD state)
- **THEN** the program SHALL fall back to `git rev-parse --short HEAD` as the branch display value

#### Scenario: Git subprocess noise suppression

- **WHEN** running any git command for dirty-check or branch detection
- **THEN** the program SHALL pass `-c core.useBuiltinFSMonitor=false` to suppress fsmonitor noise

---
### Requirement: Cost color thresholds

The program SHALL color the cost display based on value thresholds.

#### Scenario: Cost >= $10

- **WHEN** total cost >= $10.00
- **THEN** the cost SHALL be displayed in red

#### Scenario: Cost >= $5 and < $10

- **WHEN** total cost >= $5.00 and < $10.00
- **THEN** the cost SHALL be displayed in yellow

#### Scenario: Cost > $0.00 and < $5.00

- **WHEN** total cost is greater than $0.00 and less than $5.00
- **THEN** the cost SHALL be displayed in yellow

#### Scenario: Cost = $0.00

- **WHEN** total cost is exactly $0.00
- **THEN** the cost SHALL be displayed in gray (dimmed) and SHALL still be shown (not hidden)

---
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

---
### Requirement: Agent and Worktree indicator

The program SHALL display an agent or worktree indicator when active.

#### Scenario: Worktree active

- **WHEN** `worktree.name` is non-empty in the JSON
- **THEN** line 2 SHALL append `⚙ worktree:<name>` in yellow

#### Scenario: Subagent active

- **WHEN** `agent.name` is non-empty and `worktree.name` is empty
- **THEN** line 2 SHALL append `⚙ <agent_name>` in yellow

---
### Requirement: Warning symbol at high context usage

The program SHALL display a warning symbol when context usage reaches a critical level.

#### Scenario: Warning symbol shown

- **WHEN** context usage percentage is >= 90
- **THEN** the warning symbol (`⚠` in default/nerdfont mode, `!` in ASCII mode) SHALL be displayed in red immediately after the percentage value on line 1

#### Scenario: Warning symbol absent

- **WHEN** context usage percentage is < 90
- **THEN** no warning symbol SHALL be displayed

---
### Requirement: Progress bar dimensions and clamping

The program SHALL render a fixed-width progress bar with bounded input.

#### Scenario: Progress bar width

- **WHEN** rendering the progress bar in any mode
- **THEN** the bar SHALL always be exactly 10 characters wide (10 cells filled with `█`/`░` or `#`/`-`)

#### Scenario: Percentage clamping

- **WHEN** `context_window.used_percentage` is less than 0 or greater than 100
- **THEN** the value SHALL be clamped to the range [0, 100] before rendering

---
### Requirement: Section text colors

The program SHALL apply consistent ANSI colors to each section of the output.

#### Scenario: Brand diamond color

- **WHEN** rendering line 1 in non-ASCII mode with true color support
- **THEN** the `◆` symbol SHALL be rendered in Anthropic brand purple (RGB 114, 102, 234)

#### Scenario: Model name color

- **WHEN** rendering the model name on line 1
- **THEN** the model name SHALL be displayed in cyan

#### Scenario: Git branch color

- **WHEN** rendering the git branch on line 2
- **THEN** the branch name (including dirty marker `*`) SHALL be displayed in gray

#### Scenario: Directory color

- **WHEN** rendering the current directory name on line 2
- **THEN** the directory name SHALL be displayed in blue

---
### Requirement: Lines added and removed colors

The program SHALL display code change counts with directional colors.

#### Scenario: Lines added in green, removed in red

- **WHEN** `total_lines_added` or `total_lines_removed` is greater than 0
- **THEN** the added count SHALL be prefixed with `+` in green, and the removed count SHALL be prefixed with `-` in red, formatted as `+N/-N`

---
### Requirement: Directory path display on line 2

The program SHALL display the working directory on line 2 using a project root resolved by a three-step fallback: (1) use `workspace.project_dir` when it is a strict ancestor of `workspace.current_dir`; (2) otherwise walk upward from `workspace.current_dir` looking for a `.git` entry (file or directory) and use the first matching directory as root; (3) otherwise fall back to `filepath.Base(workspace.current_dir)`. The `.git` lookup SHALL use filesystem stat only and SHALL NOT invoke the `git` CLI.

When a project root is resolved, the program SHALL display `filepath.Base(root)` in blue when `workspace.current_dir` equals the root, and `filepath.Base(root) + "/" + relative_path` (forward slashes) when `workspace.current_dir` is a strict descendant of the root.

#### Scenario: Payload project_dir is a strict ancestor of current directory

- **WHEN** `workspace.project_dir` is non-empty AND `workspace.current_dir` is a strict descendant of `workspace.project_dir`
- **THEN** line 2 SHALL display `<project_base>/<relative_path>` in blue, where `project_base = filepath.Base(workspace.project_dir)` and `relative_path` uses forward slashes

#### Scenario: Payload project_dir equals current directory but inside a git repository

- **WHEN** `workspace.project_dir` equals `workspace.current_dir` (or is empty) AND an ancestor directory of `workspace.current_dir` contains a `.git` entry (file or directory)
- **THEN** line 2 SHALL display `<git_root_base>` in blue when `workspace.current_dir` equals the resolved git root, or `<git_root_base>/<relative_path>` in blue when it is a strict descendant, where `git_root_base = filepath.Base(<resolved git root>)` and `relative_path` uses forward slashes

#### Scenario: Payload project_dir unusable and no git repository detected

- **WHEN** `workspace.project_dir` cannot be used as an ancestor (empty, equals current, or not an ancestor) AND no `.git` entry is found walking upward from `workspace.current_dir` to the filesystem root
- **THEN** line 2 SHALL fall back to `filepath.Base(workspace.current_dir)` in blue

#### Scenario: Current directory is empty or "."

- **WHEN** `workspace.current_dir` is an empty string or `.`
- **THEN** line 2 SHALL display `.` in blue

#### Scenario: Git worktree with `.git` as a file

- **WHEN** an ancestor of `workspace.current_dir` contains `.git` as a regular file (git worktree layout) rather than a directory AND no usable `workspace.project_dir` is available
- **THEN** the program SHALL treat that ancestor as the project root for display purposes

#### Scenario: Git submodule treated as its own root

- **WHEN** `workspace.current_dir` is located inside a git submodule (a `.git` entry — typically a file with `gitdir:` metadata — exists at the submodule directory AND an outer `.git` directory also exists further up in the parent repository) AND no usable `workspace.project_dir` is available
- **THEN** the program SHALL resolve the submodule directory as the project root (first `.git` wins) and SHALL NOT continue walking upward to the parent repository


<!-- @trace
source: additional-statusline-indicators
updated: 2026-04-20
code:
  - internal/renderer/renderer.go
  - internal/model/payload.go
tests:
  - internal/renderer/renderer_test.go
  - internal/model/payload_test.go
-->

---
### Requirement: Parse exceeds_200k_tokens field

The program SHALL parse the top-level `exceeds_200k_tokens` boolean field from the JSON payload and store it on the parsed `Payload` struct for use by the renderer.

#### Scenario: Field present and true

- **WHEN** the JSON payload contains `"exceeds_200k_tokens": true`
- **THEN** the parsed payload SHALL expose the value `true` for this field

#### Scenario: Field present and false

- **WHEN** the JSON payload contains `"exceeds_200k_tokens": false`
- **THEN** the parsed payload SHALL expose the value `false` for this field

#### Scenario: Field absent

- **WHEN** the JSON payload omits the `exceeds_200k_tokens` field
- **THEN** the parsed payload SHALL expose the value `false` for this field (zero-value default)

<!-- @trace
source: additional-statusline-indicators
updated: 2026-04-20
code:
  - internal/renderer/renderer.go
  - internal/model/payload.go
tests:
  - internal/renderer/renderer_test.go
  - internal/model/payload_test.go
-->

---
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

<!-- @trace
source: cache-hit-rate-statusline
updated: 2026-06-14
code:
  - docs/improvement-plan.md
  - internal/gitcache/gitcache.go
  - internal/model/payload.go
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
tests:
  - internal/gitcache/gitcache_test.go
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->

---
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


<!-- @trace
source: show-execution-mode-statusline
updated: 2026-06-14
code:
  - internal/renderer/renderer.go
  - internal/model/payload.go
tests:
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->

---
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

<!-- @trace
source: show-execution-mode-statusline
updated: 2026-06-14
code:
  - internal/renderer/renderer.go
  - internal/model/payload.go
tests:
  - internal/model/payload_test.go
  - internal/renderer/renderer_test.go
-->