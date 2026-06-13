## MODIFIED Requirements

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
