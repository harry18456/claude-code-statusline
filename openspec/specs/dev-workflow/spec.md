# dev-workflow Specification

## Purpose

TBD - created by archiving change 'improve-dev-workflow'. Update Purpose after archive.

## Requirements

### Requirement: debug-tee exit code propagation

The debug-tee binary SHALL propagate the real statusline binary's exit code to its own exit code.

#### Scenario: Real binary exits with non-zero code

- **WHEN** the real statusline binary exits with a non-zero exit code
- **THEN** debug-tee SHALL exit with the same exit code

#### Scenario: Real binary exits successfully

- **WHEN** the real statusline binary exits with code 0
- **THEN** debug-tee SHALL exit with code 0


<!-- @trace
source: improve-dev-workflow
updated: 2026-04-03
code:
  - cmd/debug-tee/main.go
  - dev.sh
  - cmd/statusline/main.go
-->

---
### Requirement: dev.sh build command

The `dev.sh` script SHALL provide a `build` subcommand that compiles the statusline binary and replaces the installed binary.

#### Scenario: Successful build and install

- **WHEN** the developer runs `./dev.sh build`
- **THEN** the script SHALL run `go build ./cmd/statusline/` and copy the resulting binary to `~/.claude/statusline.exe`, replacing the previous version

#### Scenario: Build failure

- **WHEN** `go build` fails
- **THEN** the script SHALL print the error and exit with a non-zero code without replacing the installed binary


<!-- @trace
source: improve-dev-workflow
updated: 2026-04-03
code:
  - cmd/debug-tee/main.go
  - dev.sh
  - cmd/statusline/main.go
-->

---
### Requirement: dev.sh last-json command

The `dev.sh` script SHALL provide a `last-json` subcommand that extracts the most recent JSON payload from the debug log to the current working directory.

#### Scenario: Log file exists with entries

- **WHEN** the developer runs `./dev.sh last-json` and `$TEMP/cc-statusline-debug.jsonl` exists and is non-empty
- **THEN** the script SHALL write the last line of the log file to `./debug.json` in the caller's current working directory

#### Scenario: Log file does not exist

- **WHEN** the developer runs `./dev.sh last-json` and the log file does not exist
- **THEN** the script SHALL print a clear error message indicating the log file is missing and exit with a non-zero code

#### Scenario: Unknown subcommand

- **WHEN** the developer runs `./dev.sh` with an unrecognised subcommand or no subcommand
- **THEN** the script SHALL print usage information listing available subcommands and exit with a non-zero code

<!-- @trace
source: improve-dev-workflow
updated: 2026-04-03
code:
  - cmd/debug-tee/main.go
  - dev.sh
  - cmd/statusline/main.go
-->