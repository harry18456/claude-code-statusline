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


<!-- @trace
source: replace-env-config-with-flags-hide-sections
updated: 2026-06-14
code:
  - README.zh-TW.md
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
  - README.md
tests:
  - cmd/statusline/main_test.go
  - internal/renderer/renderer_test.go
-->


<!-- @trace
source: show-session-peer-name
updated: 2026-11-20
code:
  - CLAUDE.md
  - README.md
  - internal/model/payload.go
  - debug-sub.json
  - internal/peername/open_other.go
  - debug-entra.json
  - debug.json.bak
  - cmd/statusline/main.go
  - README.zh-TW.md
  - internal/peername/peername.go
  - internal/renderer/renderer.go
  - internal/peername/open_windows.go
tests:
  - internal/peername/peername_test.go
  - cmd/statusline/session_test.go
  - internal/model/payload_test.go
  - internal/peername/peername_windows_test.go
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


<!-- @trace
source: replace-env-config-with-flags-hide-sections
updated: 2026-06-14
code:
  - README.zh-TW.md
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
  - README.md
tests:
  - cmd/statusline/main_test.go
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


<!-- @trace
source: replace-env-config-with-flags-hide-sections
updated: 2026-06-14
code:
  - README.zh-TW.md
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
  - README.md
tests:
  - cmd/statusline/main_test.go
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

The program SHALL output exactly two lines of ANSI-colored text. Visible sections SHALL retain the existing order: line 1 uses model, execution mode, context bar, context size label, cost/cache, duration, and rate limits; line 2 uses branch, lines added/removed, directory, agent/worktree, and session peer name.

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
- **THEN** line 2 SHALL follow the existing visible-section order: `<branch>* <+added/-removed> <dirname> <agent_or_worktree> <session_peer_name>`

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

---
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

---
### Requirement: Nerd Font and Powerline support

The program SHALL support optional Nerd Font icons and Powerline separators through command-line flags.

#### Scenario: Nerd Font mode

- **WHEN** `--nerdfont` is set
- **THEN** the program SHALL use Nerd Font icons for model, time, and cost symbols
- **THEN** the program SHALL use Powerline separators

#### Scenario: Powerline separators

- **WHEN** `--powerline` is set
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

- **WHEN** current usage is available, the denominator is non-zero, ASCII mode is disabled, and the `cache` section is not hidden
- **THEN** the cache hit-rate segment SHALL use the existing non-ASCII cache indicator followed by the rounded whole-number percentage, for example `⚡99%`

#### Scenario: ASCII cache hit text

- **WHEN** current usage is available, the denominator is non-zero, `--ascii` is set, and the `cache` section is not hidden
- **THEN** the cache hit-rate segment SHALL use the ASCII fallback format `cache:<pct>%`
- **THEN** the cache hit-rate segment SHALL contain no ANSI color escapes
- **THEN** the cache hit-rate segment SHALL contain no Unicode glyphs


<!-- @trace
source: replace-env-config-with-flags-hide-sections
updated: 2026-06-14
code:
  - README.zh-TW.md
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
  - README.md
tests:
  - cmd/statusline/main_test.go
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


<!-- @trace
source: replace-env-config-with-flags-hide-sections
updated: 2026-06-14
code:
  - README.zh-TW.md
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
  - README.md
tests:
  - cmd/statusline/main_test.go
  - internal/renderer/renderer_test.go
-->

---
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


<!-- @trace
source: replace-env-config-with-flags-hide-sections
updated: 2026-06-14
code:
  - README.zh-TW.md
  - cmd/statusline/main.go
  - internal/renderer/renderer.go
  - README.md
tests:
  - cmd/statusline/main_test.go
  - internal/renderer/renderer_test.go
-->

---
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
| `session` | Line 2 session peer name display, including the bridge address. |
| `bridge` | Line 2 Remote Control bridge address inside the session section; the peer name stays visible. |

Hide keys SHALL be lowercase and canonical. The program SHALL trim whitespace around comma-separated keys, ignore empty tokens, and de-duplicate repeated keys. Multiple `--hide` flags SHALL merge into a single hidden set. Unknown non-empty hide keys SHALL be ignored with a stderr warning; known keys from the same `--hide` value SHALL still apply.

#### Scenario: Default visibility

- **WHEN** the program is invoked without `--hide`
- **THEN** every section with available data SHALL render according to its existing requirement

#### Scenario: Hide line 1 sections

- **WHEN** the program is invoked with `--hide model,effort,bar,size,cost,cache,duration,rate`
- **THEN** each listed line 1 section SHALL be omitted even when the payload contains data that would normally render it
- **THEN** unlisted line 2 sections SHALL remain eligible for rendering

#### Scenario: Hide line 2 sections

- **WHEN** the program is invoked with `--hide branch,lines,dir,agent,session,bridge`
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

- **GIVEN** line 2 would normally contain `branch`, `lines`, `dir`, `agent`, and `session`
- **WHEN** the program is invoked with `--hide branch,lines,dir,agent,session`
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


<!-- @trace
source: show-rc-bridge-address
updated: 2026-11-20
code:
  - debug-entra.json
  - CLAUDE.md
  - debug.json.bak
  - README.zh-TW.md
  - cmd/statusline/main.go
  - README.md
  - internal/renderer/renderer.go
  - debug-sub.json
  - internal/peername/peername.go
tests:
  - internal/renderer/renderer_test.go
  - cmd/statusline/session_test.go
  - internal/peername/peername_test.go
-->

---
### Requirement: Parse session id field

The program SHALL parse the top-level `session_id` string field from the payload. When `session_id` is absent, null, or not a JSON string, the program SHALL treat the session id as unavailable and SHALL continue to parse every other payload field unchanged.

#### Scenario: Session id present

- **WHEN** the payload contains `"session_id": "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"`
- **THEN** the parsed session id SHALL equal `aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee`

#### Scenario: Session id unavailable

- **WHEN** the payload omits `session_id`, sets it to null, or sets it to a non-string value
- **THEN** the parsed session id SHALL be empty
- **THEN** the payload SHALL parse without error and every other field SHALL keep its parsed value

##### Example: unavailable session id values

| `session_id` value | Parsed session id | Parse error |
| ------------------ | ----------------- | ----------- |
| (absent) | empty | none |
| `null` | empty | none |
| `123` | empty | none |
| `{"id":"x"}` | empty | none |


<!-- @trace
source: show-session-peer-name
updated: 2026-11-20
code:
  - CLAUDE.md
  - README.md
  - internal/model/payload.go
  - debug-sub.json
  - internal/peername/open_other.go
  - debug-entra.json
  - debug.json.bak
  - cmd/statusline/main.go
  - README.zh-TW.md
  - internal/peername/peername.go
  - internal/renderer/renderer.go
  - internal/peername/open_windows.go
tests:
  - internal/peername/peername_test.go
  - cmd/statusline/session_test.go
  - internal/model/payload_test.go
  - internal/peername/peername_windows_test.go
  - internal/renderer/renderer_test.go
-->

---
### Requirement: Resolve session peer name from the session registry

The program SHALL resolve the session peer name, which is the name that other Claude Code sessions use as the `SendMessage` address, by reading the Claude Code session registry. The program SHALL NOT use the payload `session_name` field as the peer name.

The lookup SHALL follow these rules:

- The candidate config directories SHALL be, in order: the value of `CLAUDE_CONFIG_DIR` when it is non-empty after trimming whitespace, then `<home>/.claude`. A directory that appears twice after path cleaning SHALL be searched once.
- In each candidate directory, the program SHALL read only regular files located directly inside `<config>/sessions/` whose names end in `.json`. Entries with any other suffix, such as `.key`, and entries that are not regular files SHALL NOT be read.
- After opening a file, the program SHALL confirm on the open handle that it is a regular file before reading any content. A file that is a pipe or a device once opened SHALL be skipped without reading.
- Files larger than 65536 bytes SHALL be skipped.
- A file SHALL match when it parses as a JSON object whose `sessionId` string equals the payload session id exactly and whose `name` string is non-empty after sanitization.
- Sanitization SHALL remove every character in the Unicode general categories Cc (control), Cf (format, which includes bidirectional controls and zero-width characters), Zl (line separator), and Zp (paragraph separator), and SHALL then trim surrounding whitespace.
- When several files in one directory match, the match with the greatest numeric `updatedAt` SHALL win. Ties keep the earliest match in file-name order. A missing or non-numeric `updatedAt` SHALL count as 0 and SHALL NOT prevent a match.
- The first candidate directory that yields a match SHALL determine the result. Later candidate directories SHALL NOT be searched.
- An empty session id SHALL resolve to no peer name without reading any file.
- Every failure, including a missing directory, an unreadable file, an empty or partially written file, invalid JSON, and a `sessionId` or `name` that is not a JSON string, SHALL result in no peer name for that file or directory. Lookup failures SHALL NOT write to stderr and SHALL NOT change the exit code.

#### Scenario: Matching registry entry

- **WHEN** `<config>/sessions/42184.json` contains `{"sessionId":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee","name":"claude-code-statusline-56","nameSource":"derived"}`
- **WHEN** the payload session id is `aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee`
- **THEN** the resolved peer name SHALL be `claude-code-statusline-56`

#### Scenario: Payload session name is not used

- **WHEN** the payload contains `"session_name": "Refactor config loader"` and the registry entry for the same session id has `"name": "my-project-75"`
- **THEN** the resolved peer name SHALL be `my-project-75`

#### Scenario: Config directory precedence

- **WHEN** `CLAUDE_CONFIG_DIR` is set and both `CLAUDE_CONFIG_DIR/sessions/` and `<home>/.claude/sessions/` contain a matching entry with different names
- **THEN** the resolved peer name SHALL be the name from `CLAUDE_CONFIG_DIR/sessions/`

#### Scenario: Fallback to the home config directory

- **WHEN** `CLAUDE_CONFIG_DIR` is set but `CLAUDE_CONFIG_DIR/sessions/` contains no match
- **WHEN** `<home>/.claude/sessions/` contains a matching entry
- **THEN** the resolved peer name SHALL be the name from `<home>/.claude/sessions/`

#### Scenario: Most recently updated match wins

- **WHEN** several files in the same sessions directory match the payload session id
- **THEN** the resolved peer name SHALL come from the match with the greatest `updatedAt`

##### Example: two matching files

- **GIVEN** `a.json` contains `{"sessionId":"s-1","name":"old-name","updatedAt":100}`
- **GIVEN** `b.json` contains `{"sessionId":"s-1","name":"new-name","updatedAt":200}`
- **WHEN** the payload session id is `s-1`
- **THEN** the resolved peer name SHALL be `new-name`

#### Scenario: Control characters are removed from the name

- **WHEN** the matching entry name contains characters from the categories Cc, Cf, Zl, or Zp
- **THEN** the resolved peer name SHALL contain none of those characters

##### Example: newline and ESC in the name

- **GIVEN** the matching entry name is `"peer\nname\u001b[31m"`
- **WHEN** the program resolves the peer name
- **THEN** the resolved peer name SHALL be `peername[31m`

##### Example: invisible characters in the name

| Registry name | Resolved peer name | Removed characters |
| ------------- | ------------------ | ------------------ |
| `peer` + U+202E RIGHT-TO-LEFT OVERRIDE + `name` + U+202C POP DIRECTIONAL FORMATTING | `peername` | bidirectional controls (Cf) |
| U+2066 LEFT-TO-RIGHT ISOLATE + `peer` + U+2069 POP DIRECTIONAL ISOLATE | `peer` | bidirectional isolates (Cf) |
| `a` + U+200B ZERO WIDTH SPACE + `b` + U+200D ZERO WIDTH JOINER + `c` + U+FEFF ZERO WIDTH NO-BREAK SPACE + `d` | `abcd` | zero-width characters (Cf) |
| `x` + U+2028 LINE SEPARATOR + `y` + U+2029 PARAGRAPH SEPARATOR + `z` | `xyz` | separators (Zl, Zp) |
| ESC + ` spaced` | `spaced` | ESC is removed before trimming |

#### Scenario: Non-numeric updatedAt does not block a match

- **WHEN** a matching entry has an `updatedAt` value that is a string, null, or an object
- **THEN** the entry SHALL still match, with `updatedAt` counted as 0

##### Example: string and numeric updatedAt

- **GIVEN** `a.json` contains `{"sessionId":"s-1","name":"string-ts","updatedAt":"9999"}`
- **GIVEN** `b.json` contains `{"sessionId":"s-1","name":"numeric-ts","updatedAt":5}`
- **WHEN** the payload session id is `s-1`
- **THEN** the resolved peer name SHALL be `numeric-ts`

#### Scenario: Pipe or device swapped in after the directory scan

- **WHEN** a `.json` path in the sessions directory resolves to a named pipe or a device at open time
- **THEN** the program SHALL skip it without reading from it and SHALL NOT block

#### Scenario: Registry unavailable or unusable

- **WHEN** no candidate directory contains a usable matching entry
- **THEN** the resolved peer name SHALL be empty
- **THEN** the lookup SHALL write nothing to stderr
- **THEN** the program SHALL render both lines and exit with code 0

##### Example: unusable registry content

| Registry state | Resolved peer name | Notes |
| -------------- | ------------------ | ----- |
| `sessions/` directory missing | empty | directory skipped |
| `40404.json` contains only whitespace | empty | file skipped |
| `40404.json` contains `{"sessionId":` | empty | partially written file skipped |
| matching `sessionId` with `"name": ""` | empty | empty name is not a match |
| matching `sessionId` with `"name": 42` | empty | wrong type is not a match |
| matching data stored in `40404.abc.key` | empty | `.key` files are never read |
| matching data in a 65537-byte `.json` file | empty | oversized file skipped |
| `sessionId` differs only by letter case | empty | exact match required |


<!-- @trace
source: show-session-peer-name
updated: 2026-11-20
code:
  - CLAUDE.md
  - README.md
  - internal/model/payload.go
  - debug-sub.json
  - internal/peername/open_other.go
  - debug-entra.json
  - debug.json.bak
  - cmd/statusline/main.go
  - README.zh-TW.md
  - internal/peername/peername.go
  - internal/renderer/renderer.go
  - internal/peername/open_windows.go
tests:
  - internal/peername/peername_test.go
  - cmd/statusline/session_test.go
  - internal/model/payload_test.go
  - internal/peername/peername_windows_test.go
  - internal/renderer/renderer_test.go
-->

---
### Requirement: Session peer name display

The program SHALL display the resolved session peer name as the last visible section of line 2, after the agent or worktree indicator. When a bridge address is available and the `bridge` section is not hidden, the same section SHALL also show the bridge address after the name.

- In default mode and Nerd Font mode, the section SHALL render as `⇄ <name>` in gray (`\033[90m`), followed by an ANSI reset. With a bridge address, the section SHALL render as `⇄ <name> · bridge:<id>` inside the same gray span.
- In ASCII mode, the section SHALL render as `peer:<name>` with no ANSI color codes and no renderer-added Unicode glyph. With a bridge address, the section SHALL render as `peer:<name> bridge:<id>`.
- When the resolved peer name is empty, the section SHALL be omitted and SHALL NOT emit a separator, even when a bridge address is available.
- When the `session` section is hidden by `--hide`, the section SHALL be omitted and the program SHALL NOT read the session registry.
- When only the `bridge` section is hidden by `--hide`, the section SHALL render the peer name without the bridge address.

#### Scenario: Default mode display

- **WHEN** the resolved peer name is `claude-code-statusline-56` and the program renders in default mode
- **THEN** line 2 SHALL end with `\033[90m⇄ claude-code-statusline-56\033[0m`

#### Scenario: ASCII mode display

- **WHEN** the resolved peer name is `claude-code-statusline-56` and `--ascii` is set
- **THEN** line 2 SHALL end with `peer:claude-code-statusline-56`
- **THEN** the session section SHALL contain no ANSI escape sequence

#### Scenario: Default mode display with bridge address

- **WHEN** the resolved peer name is `claude-code-statusline-56`, the bridge session id is `session_01EXAMPLEb2kQ7mZp4Lx9VdR`, and the program renders in default mode
- **THEN** line 2 SHALL end with `\033[90m⇄ claude-code-statusline-56 · bridge:session_01EXAMPLEb2kQ7mZp4Lx9VdR\033[0m`

#### Scenario: ASCII mode display with bridge address

- **WHEN** the resolved peer name is `claude-code-statusline-56`, the bridge session id is `session_01EXAMPLEb2kQ7mZp4Lx9VdR`, and `--ascii` is set
- **THEN** line 2 SHALL end with `peer:claude-code-statusline-56 bridge:session_01EXAMPLEb2kQ7mZp4Lx9VdR`
- **THEN** the session section SHALL contain no ANSI escape sequence

#### Scenario: Hide only the bridge address

- **WHEN** the program is invoked with `--hide bridge` and both a peer name and a bridge address are available
- **THEN** line 2 SHALL end with the name-only session section
- **THEN** line 2 SHALL NOT contain `bridge:`

#### Scenario: Bridge address without a peer name

- **WHEN** the resolved peer name is empty and a bridge session id is set
- **THEN** line 2 SHALL NOT contain `⇄`, `peer:`, or `bridge:`

#### Scenario: Session section after the agent indicator

- **WHEN** the payload has `agent.name = "code-reviewer"` and the resolved peer name is `claude-code-statusline-56`
- **THEN** the session section SHALL appear after the `⚙ code-reviewer` section, separated by exactly one separator

#### Scenario: No peer name

- **WHEN** the resolved peer name is empty
- **THEN** line 2 SHALL NOT contain `⇄` or `peer:`
- **THEN** line 2 SHALL NOT end with a separator

#### Scenario: Hidden session section skips registry access

- **WHEN** the program is invoked with `--hide session` and the payload contains a session id that has a matching registry entry
- **THEN** line 2 SHALL NOT contain the peer name
- **THEN** the program SHALL NOT read any file under `<config>/sessions/`


<!-- @trace
source: show-rc-bridge-address
updated: 2026-11-20
code:
  - debug-entra.json
  - CLAUDE.md
  - debug.json.bak
  - README.zh-TW.md
  - cmd/statusline/main.go
  - README.md
  - internal/renderer/renderer.go
  - debug-sub.json
  - internal/peername/peername.go
tests:
  - internal/renderer/renderer_test.go
  - cmd/statusline/session_test.go
  - internal/peername/peername_test.go
-->

---
### Requirement: Resolve Remote Control bridge address

The program SHALL read the `bridgeSessionId` field from the same registry entry that supplies the session peer name. The bridge address SHALL be `bridge:` followed by that value. Claude Code sets `bridgeSessionId` while the session is connected to Remote Control, sets it to null after Remote Control disconnects, and assigns a new value after each reconnect.

The program SHALL follow these rules:

- The value SHALL be used only when it is a JSON string of 1 to 128 characters, and every character is an ASCII letter, an ASCII digit, `_`, or `-`.
- A missing, null, non-string, empty, too long, or otherwise invalid value SHALL mean that there is no bridge address.
- An unusable `bridgeSessionId` SHALL NOT prevent the entry from matching and SHALL NOT change which matching entry wins. The file-level rules of the registry lookup still apply to the whole file: a file larger than 65536 bytes or with invalid JSON is skipped, whatever its `bridgeSessionId` holds.
- The bridge address SHALL come from the winning entry only. The program SHALL NOT take a bridge value from any other matching entry.

#### Scenario: Connected to Remote Control

- **WHEN** the matching entry contains `"bridgeSessionId":"session_01EXAMPLEb2kQ7mZp4Lx9VdR"`
- **THEN** the resolved bridge session id SHALL be `session_01EXAMPLEb2kQ7mZp4Lx9VdR`

#### Scenario: Remote Control disconnected

- **WHEN** the matching entry contains `"bridgeSessionId":null`
- **THEN** there SHALL be no bridge address
- **THEN** the peer name SHALL still resolve from that entry

##### Example: bridge value handling

| `bridgeSessionId` value | Bridge address | Peer name still resolves |
| ----------------------- | -------------- | ------------------------ |
| `"session_01EXAMPLEb2kQ7mZp4Lx9VdR"` | `bridge:session_01EXAMPLEb2kQ7mZp4Lx9VdR` | yes |
| `"session_01-_Ab9"` | `bridge:session_01-_Ab9` | yes |
| `null` | none | yes |
| (absent) | none | yes |
| `42` | none | yes |
| `""` | none | yes |
| `"session_01 bad"` (contains a space) | none | yes |
| a string that contains ESC | none | yes |
| a string of 129 `a` characters | none | yes |

#### Scenario: Bridge comes from the winning entry

- **WHEN** `a.json` matches with `updatedAt` 100 and a valid `bridgeSessionId`
- **WHEN** `b.json` matches with `updatedAt` 200 and `"bridgeSessionId":null`
- **THEN** the peer name SHALL come from `b.json`
- **THEN** there SHALL be no bridge address

<!-- @trace
source: show-rc-bridge-address
updated: 2026-11-20
code:
  - debug-entra.json
  - CLAUDE.md
  - debug.json.bak
  - README.zh-TW.md
  - cmd/statusline/main.go
  - README.md
  - internal/renderer/renderer.go
  - debug-sub.json
  - internal/peername/peername.go
tests:
  - internal/renderer/renderer_test.go
  - cmd/statusline/session_test.go
  - internal/peername/peername_test.go
-->