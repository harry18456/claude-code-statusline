## Problem

`internal/gitcache` currently stores every statusline session in the same temp file. A fresh cache hit ignores the requested working directory, so two Claude Code sessions opened in different repositories can display each other's branch and dirty state within the 5-second TTL.

The cache writer also uses `os.WriteFile`, which truncates and rewrites the shared file in place. Concurrent sessions can observe an empty or partial cache file while another process is writing.

## Root Cause

- `DefaultCacheFile()` returns a fixed `os.TempDir()/claude-statusline-git-cache` path for every repository.
- `Get(dir, cacheFile, maxAge)` trusts the caller-provided cache path on cache hit and does not bind cached contents to `dir`.
- `writeCache` writes directly to the final cache file instead of publishing a complete file atomically.

## Proposed Solution

- Change the gitcache API so cache selection is derived from the requested directory:
  - `DefaultCacheFile(dir string)` returns a temp path whose filename includes a SHA-256 hash prefix of a normalized directory key.
  - `Get(dir string, maxAge time.Duration)` computes its own cache path via `DefaultCacheFile(dir)`.
- Use a stable filename format such as `claude-statusline-git-cache-<hash>`, where `<hash>` is the first 16 bytes of `sha256(dirKey)` encoded as 32 lowercase hex characters.
- Normalize the cache key with `filepath.Clean` and `filepath.Abs` when possible. If absolute resolution fails, fall back to the cleaned input path.
- Replace direct `os.WriteFile` writes with an atomic publish sequence:
  - create a temp file with `os.CreateTemp(filepath.Dir(cacheFile), ...)` in the same directory as the final cache file,
  - write the full `branch|dirty` payload,
  - close the temp file,
  - publish it with `os.Rename(tempName, cacheFile)`,
  - remove the temp file on any failure before a successful rename.
- Update `cmd/statusline/main.go` to use the new `gitcache.Get` signature.

## Non-Goals

- Do not change git dirty-state semantics or `git diff --quiet` exit-code handling.
- Do not change TTL freshness semantics based on `ModTime`.
- Do not implement temp-file symlink hardening, `O_NOFOLLOW`, debug-tee hardening, or cache directory migration.
- Do not change renderer output, ANSI formatting, payload parsing, or rate-limit behavior.
- Do not add new runtime dependencies.

## Success Criteria

- Distinct directories derive distinct cache files and cannot read each other's branch or dirty state through a fresh cache hit.
- Concurrent `Get` calls for the same repository do not produce torn reads, invalid cache payloads, data races, or wrong branch values under `go test -race`.
- Cache writes are implemented as temp-file write plus same-directory `os.Rename`.
- `cmd/statusline/main.go` compiles with the updated gitcache API.
- Existing gitcache tests continue to pass.
- Apply-phase verification commands pass: `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test -race ./...`, and `gofmt -l .`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `go-statusline`: git branch dirty-check caching becomes isolated by working directory and publishes cache updates atomically.

## Impact

- Affected spec: `go-statusline`
- Affected code: `internal/gitcache/gitcache.go`, `internal/gitcache/gitcache_test.go`, `cmd/statusline/main.go`
- Internal API change: `DefaultCacheFile` and `Get` signatures change inside the Go module.
- Dependencies: none.
