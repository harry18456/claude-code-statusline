## ADDED Requirements

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

## MODIFIED Requirements

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
