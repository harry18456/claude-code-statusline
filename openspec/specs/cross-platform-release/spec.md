## ADDED Requirements

### Requirement: Cross-compile for all target platforms

The release workflow SHALL produce binary artifacts for all supported platforms using Go cross-compilation.

#### Scenario: Supported build targets

- **WHEN** a release is triggered
- **THEN** the workflow SHALL build binaries for: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `windows/amd64`

#### Scenario: Binary naming convention

- **WHEN** producing build artifacts
- **THEN** each binary SHALL be named `statusline-<os>-<arch>` (e.g., `statusline-darwin-arm64`), except Windows which SHALL use `.exe` suffix (`statusline-windows-amd64.exe`)


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

### Requirement: Triggered by git tag push

The release workflow SHALL be triggered automatically when a version tag is pushed.

#### Scenario: Tag format

- **WHEN** a tag matching `v*` (e.g., `v1.0.0`) is pushed to the repository
- **THEN** the GitHub Actions workflow SHALL start building and releasing


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

### Requirement: Attach binaries to GitHub Release

The workflow SHALL create a GitHub Release and attach all compiled binaries as release assets.

#### Scenario: Release asset upload

- **WHEN** all builds succeed
- **THEN** the workflow SHALL create a GitHub Release for the tag and upload all four platform binaries as assets


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

### Requirement: Install script detects platform and downloads correct binary

The `install.sh` script SHALL auto-detect the current OS and architecture and download the matching binary from the latest GitHub Release.

#### Scenario: macOS ARM64 (Apple Silicon)

- **WHEN** `uname -s` returns `Darwin` and `uname -m` returns `arm64`
- **THEN** `install.sh` SHALL download `statusline-darwin-arm64` and install it to `~/.claude/statusline`

#### Scenario: macOS x86_64

- **WHEN** `uname -s` returns `Darwin` and `uname -m` returns `x86_64`
- **THEN** `install.sh` SHALL download `statusline-darwin-amd64` and install it to `~/.claude/statusline`

#### Scenario: Linux x86_64

- **WHEN** `uname -s` returns `Linux` and `uname -m` returns `x86_64`
- **THEN** `install.sh` SHALL download `statusline-linux-amd64` and install it to `~/.claude/statusline`

#### Scenario: Unsupported platform

- **WHEN** `uname -s` returns a value other than `Darwin` or `Linux`
- **THEN** `install.sh` SHALL print an error message instructing the user to manually download the appropriate binary from GitHub Releases

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

### Requirement: Cross-compile for all target platforms

The release workflow SHALL produce binary artifacts for all supported platforms using Go cross-compilation.

#### Scenario: Supported build targets

- **WHEN** a release is triggered
- **THEN** the workflow SHALL build binaries for: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `windows/amd64`

#### Scenario: Binary naming convention

- **WHEN** producing build artifacts
- **THEN** each binary SHALL be named `statusline-<os>-<arch>` (e.g., `statusline-darwin-arm64`), except Windows which SHALL use `.exe` suffix (`statusline-windows-amd64.exe`)

---
### Requirement: Triggered by git tag push

The release workflow SHALL be triggered automatically when a version tag is pushed.

#### Scenario: Tag format

- **WHEN** a tag matching `v*` (e.g., `v1.0.0`) is pushed to the repository
- **THEN** the GitHub Actions workflow SHALL start building and releasing

---
### Requirement: Attach binaries to GitHub Release

The workflow SHALL create a GitHub Release and attach all compiled binaries as release assets.

#### Scenario: Release asset upload

- **WHEN** all builds succeed
- **THEN** the workflow SHALL create a GitHub Release for the tag and upload all four platform binaries as assets

---
### Requirement: Install script detects platform and downloads correct binary

The `install.sh` script SHALL auto-detect the current OS and architecture and download the matching binary from the latest GitHub Release.

#### Scenario: macOS ARM64 (Apple Silicon)

- **WHEN** `uname -s` returns `Darwin` and `uname -m` returns `arm64`
- **THEN** `install.sh` SHALL download `statusline-darwin-arm64` and install it to `~/.claude/statusline`

#### Scenario: macOS x86_64

- **WHEN** `uname -s` returns `Darwin` and `uname -m` returns `x86_64`
- **THEN** `install.sh` SHALL download `statusline-darwin-amd64` and install it to `~/.claude/statusline`

#### Scenario: Linux x86_64

- **WHEN** `uname -s` returns `Linux` and `uname -m` returns `x86_64`
- **THEN** `install.sh` SHALL download `statusline-linux-amd64` and install it to `~/.claude/statusline`

#### Scenario: Unsupported platform

- **WHEN** `uname -s` returns a value other than `Darwin` or `Linux`
- **THEN** `install.sh` SHALL print an error message instructing the user to manually download the appropriate binary from GitHub Releases