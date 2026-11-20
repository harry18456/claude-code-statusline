# ci-tests Specification

## Purpose

TBD - created by archiving change 'add-ci-race-tests'. Update Purpose after archive.

## Requirements

### Requirement: Run vet and race tests in CI

The repository SHALL run `go vet ./...` and `go test -race -count=1 -v ./...` in GitHub Actions on `ubuntu-latest` and `windows-latest` for every push to `main`, every pull request, and every call from another workflow. Each operating system SHALL run as a separate job, and a failure on one operating system SHALL NOT cancel the job on the other. The Go version SHALL come from `go.mod`.

#### Scenario: Push to main

- **WHEN** a commit is pushed to `main`
- **THEN** the test workflow SHALL run vet and race tests on `ubuntu-latest` and on `windows-latest`
- **THEN** the workflow run SHALL fail when either command fails on either operating system

#### Scenario: Pull request

- **WHEN** a pull request is opened or updated
- **THEN** the test workflow SHALL run the same jobs as for a push to `main`

#### Scenario: Independent operating system jobs

- **WHEN** a test fails on `windows-latest`
- **THEN** the `ubuntu-latest` job SHALL still run to completion and report its own result

#### Scenario: Read-only token

- **WHEN** the test workflow runs
- **THEN** its `GITHUB_TOKEN` SHALL have only the `contents: read` permission


<!-- @trace
source: add-ci-race-tests
updated: 2026-11-20
code:
  - .github/workflows/release.yml
  - .github/workflows/test.yml
  - debug-sub.json
  - debug-entra.json
  - CLAUDE.md
  - debug.json.bak
-->

---
### Requirement: Gate release builds on tests

The release workflow SHALL run the test workflow first. It SHALL build and publish binaries only after every test job succeeds.

#### Scenario: Tests pass

- **WHEN** a `v*` tag is pushed and every test job succeeds
- **THEN** the release workflow SHALL build the binaries and publish the GitHub Release

#### Scenario: Tests fail

- **WHEN** a `v*` tag is pushed and any test job fails
- **THEN** the release workflow SHALL NOT build binaries
- **THEN** the release workflow SHALL NOT create a GitHub Release

<!-- @trace
source: add-ci-race-tests
updated: 2026-11-20
code:
  - .github/workflows/release.yml
  - .github/workflows/test.yml
  - debug-sub.json
  - debug-entra.json
  - CLAUDE.md
  - debug.json.bak
-->