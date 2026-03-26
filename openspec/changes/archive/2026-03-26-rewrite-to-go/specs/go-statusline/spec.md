## ADDED Requirements

### Requirement: Read JSON from stdin

The program SHALL read a JSON payload from stdin, parse it using the standard library, and extract all fields equivalent to the current `jq` call in `statusline.sh`.

#### Scenario: Valid JSON input

- **WHEN** Claude Code pipes a valid JSON payload to the program via stdin
- **THEN** the program SHALL parse all 14 fields (model, context percentage, cost, directory, branch, rate limits, agent name, lines added/removed, duration, context size, worktree name)

#### Scenario: Invalid or empty JSON input

- **WHEN** stdin contains invalid JSON or is empty
- **THEN** the program SHALL output a single fallback line `─ │ parse error` with gray ANSI color and exit with code 0

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

### Requirement: Nerd Font and Powerline support

The program SHALL support optional Nerd Font icons and Powerline separators.

#### Scenario: Nerd Font mode

- **WHEN** `CLAUDE_STATUSLINE_NERDFONT=1` is set
- **THEN** the program SHALL use Nerd Font icons for model, time, and cost symbols

#### Scenario: Powerline separators

- **WHEN** `CLAUDE_STATUSLINE_POWERLINE=1` is set (or implied by NERDFONT=1)
- **THEN** the program SHALL use Powerline arrow separators instead of `│`

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

### Requirement: Context window size label

The program SHALL display context window size only when not already implied by the model name.

#### Scenario: 1M context window

- **WHEN** context_window_size >= 1,000,000 AND model name does not contain "context" or "Context"
- **THEN** the program SHALL display ` 1M` in gray after the percentage

#### Scenario: 200k context window

- **WHEN** context_window_size >= 200,000 AND model name does not contain "context" or "Context"
- **THEN** the program SHALL display ` 200k` in gray after the percentage

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

### Requirement: Agent and Worktree indicator

The program SHALL display an agent or worktree indicator when active.

#### Scenario: Worktree active

- **WHEN** `worktree.name` is non-empty in the JSON
- **THEN** line 2 SHALL append `⚙ worktree:<name>` in yellow

#### Scenario: Subagent active

- **WHEN** `agent.name` is non-empty and `worktree.name` is empty
- **THEN** line 2 SHALL append `⚙ <agent_name>` in yellow

### Requirement: Warning symbol at high context usage

The program SHALL display a warning symbol when context usage reaches a critical level.

#### Scenario: Warning symbol shown

- **WHEN** context usage percentage is >= 90
- **THEN** the warning symbol (`⚠` in default/nerdfont mode, `!` in ASCII mode) SHALL be displayed in red immediately after the percentage value on line 1

#### Scenario: Warning symbol absent

- **WHEN** context usage percentage is < 90
- **THEN** no warning symbol SHALL be displayed

### Requirement: Progress bar dimensions and clamping

The program SHALL render a fixed-width progress bar with bounded input.

#### Scenario: Progress bar width

- **WHEN** rendering the progress bar in any mode
- **THEN** the bar SHALL always be exactly 10 characters wide (10 cells filled with `█`/`░` or `#`/`-`)

#### Scenario: Percentage clamping

- **WHEN** `context_window.used_percentage` is less than 0 or greater than 100
- **THEN** the value SHALL be clamped to the range [0, 100] before rendering

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

### Requirement: Lines added and removed colors

The program SHALL display code change counts with directional colors.

#### Scenario: Lines added in green, removed in red

- **WHEN** `total_lines_added` or `total_lines_removed` is greater than 0
- **THEN** the added count SHALL be prefixed with `+` in green, and the removed count SHALL be prefixed with `-` in red, formatted as `+N/-N`
