// Package gitcache resolves the current git branch and dirty state,
// caching the result in a temp file to avoid frequent git subprocess calls.
package gitcache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DefaultCacheFile returns the default cache file path using os.TempDir().
func DefaultCacheFile() string {
	return filepath.Join(os.TempDir(), "claude-statusline-git-cache")
}

// Get returns (branch, dirty) for the given directory.
// It reads from cacheFile if it's fresher than maxAge, otherwise runs git.
// On error or when dir is not a git repo, returns ("", false).
func Get(dir, cacheFile string, maxAge time.Duration) (branch string, dirty bool) {
	if !isCacheStale(cacheFile, maxAge) {
		b, d, err := readCache(cacheFile)
		if err == nil {
			return b, d
		}
	}

	// Cache miss — run git
	b, d, isGit := fetchFromGit(dir)
	if !isGit {
		// Write empty cache so we don't re-run git on every call
		_ = writeCache(cacheFile, "", false)
		return "", false
	}

	_ = writeCache(cacheFile, b, d)
	return b, d
}

// ─── Cache I/O ────────────────────────────────────────────────────────────────

// writeCache writes "branch|dirty" to cacheFile.
// dirty is encoded as "1" (true) or "0" (false).
func writeCache(cacheFile, branch string, dirty bool) error {
	dirtyStr := "0"
	if dirty {
		dirtyStr = "1"
	}
	content := fmt.Sprintf("%s|%s", branch, dirtyStr)
	return os.WriteFile(cacheFile, []byte(content), 0o600)
}

// readCache reads the cache file and returns (branch, dirty, err).
func readCache(cacheFile string) (branch string, dirty bool, err error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", false, err
	}
	parts := strings.SplitN(string(data), "|", 2)
	if len(parts) != 2 {
		return "", false, fmt.Errorf("invalid cache format")
	}
	branch = parts[0]
	dirty = parts[1] == "1"
	return branch, dirty, nil
}

// isCacheStale returns true when cacheFile is absent or older than maxAge.
func isCacheStale(cacheFile string, maxAge time.Duration) bool {
	info, err := os.Stat(cacheFile)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > maxAge
}

// ─── Git subprocess ───────────────────────────────────────────────────────────

// fetchFromGit runs git commands to determine branch and dirty state.
// Returns (branch, dirty, isGitRepo).
func fetchFromGit(dir string) (branch string, dirty bool, isGit bool) {
	// Confirm it's a git repo
	if err := gitCmd(dir, "rev-parse", "--git-dir").Run(); err != nil {
		return "", false, false
	}

	// Branch name
	out, err := gitCmd(dir, "branch", "--show-current").Output()
	if err == nil {
		branch = strings.TrimSpace(string(out))
	}
	if branch == "" {
		// Detached HEAD fallback to short SHA
		out, err = gitCmd(dir, "rev-parse", "--short", "HEAD").Output()
		if err == nil {
			branch = strings.TrimSpace(string(out))
		}
	}

	// Dirty check: staged or unstaged changes
	errU := gitCmd(dir, "diff", "--quiet").Run()
	errS := gitCmd(dir, "diff", "--cached", "--quiet").Run()
	dirty = errU != nil || errS != nil

	return branch, dirty, true
}

// gitCmd builds an exec.Cmd for a git subcommand in the given directory.
// Adds -c core.useBuiltinFSMonitor=false to suppress fsmonitor noise.
func gitCmd(dir string, args ...string) *exec.Cmd {
	allArgs := append([]string{"-C", dir, "-c", "core.useBuiltinFSMonitor=false"}, args...)
	cmd := exec.Command("git", allArgs...)
	cmd.Env = os.Environ()
	return cmd
}
