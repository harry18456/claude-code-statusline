## ADDED Requirements

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

### Requirement: Session peer name display

The program SHALL display the resolved session peer name as the last visible section of line 2, after the agent or worktree indicator.

- In default mode and Nerd Font mode, the section SHALL render as `⇄ <name>` in gray (`\033[90m`), followed by an ANSI reset.
- In ASCII mode, the section SHALL render as `peer:<name>` with no ANSI color codes and no renderer-added Unicode glyph.
- When the resolved peer name is empty, the section SHALL be omitted and SHALL NOT emit a separator.
- When the `session` section is hidden by `--hide`, the section SHALL be omitted and the program SHALL NOT read the session registry.

#### Scenario: Default mode display

- **WHEN** the resolved peer name is `claude-code-statusline-56` and the program renders in default mode
- **THEN** line 2 SHALL end with `\033[90m⇄ claude-code-statusline-56\033[0m`

#### Scenario: ASCII mode display

- **WHEN** the resolved peer name is `claude-code-statusline-56` and `--ascii` is set
- **THEN** line 2 SHALL end with `peer:claude-code-statusline-56`
- **THEN** the session section SHALL contain no ANSI escape sequence

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

## MODIFIED Requirements

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
| `session` | Line 2 session peer name display. |

Hide keys SHALL be lowercase and canonical. The program SHALL trim whitespace around comma-separated keys, ignore empty tokens, and de-duplicate repeated keys. Multiple `--hide` flags SHALL merge into a single hidden set. Unknown non-empty hide keys SHALL be ignored with a stderr warning; known keys from the same `--hide` value SHALL still apply.

#### Scenario: Default visibility

- **WHEN** the program is invoked without `--hide`
- **THEN** every section with available data SHALL render according to its existing requirement

#### Scenario: Hide line 1 sections

- **WHEN** the program is invoked with `--hide model,effort,bar,size,cost,cache,duration,rate`
- **THEN** each listed line 1 section SHALL be omitted even when the payload contains data that would normally render it
- **THEN** unlisted line 2 sections SHALL remain eligible for rendering

#### Scenario: Hide line 2 sections

- **WHEN** the program is invoked with `--hide branch,lines,dir,agent,session`
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
