package gitcache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ─── Cache hit / miss ─────────────────────────────────────────────────────────

func TestCacheHit(t *testing.T) {
	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "test-git-cache")

	// Write a fresh cache entry
	writeCache(cacheFile, "main", false)

	info, err := os.Stat(cacheFile)
	if err != nil {
		t.Fatalf("cache file not created: %v", err)
	}
	age := time.Since(info.ModTime())
	if age > 5*time.Second {
		t.Fatalf("cache is stale, age = %v", age)
	}

	branch, dirty, err := readCache(cacheFile)
	if err != nil {
		t.Fatalf("readCache: %v", err)
	}
	if branch != "main" {
		t.Errorf("branch: got %q, want %q", branch, "main")
	}
	if dirty {
		t.Error("dirty should be false")
	}
}

func TestCacheHitDirty(t *testing.T) {
	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "test-git-cache")
	writeCache(cacheFile, "feat/foo", true)

	branch, dirty, err := readCache(cacheFile)
	if err != nil {
		t.Fatalf("readCache: %v", err)
	}
	if branch != "feat/foo" {
		t.Errorf("branch: got %q, want %q", branch, "feat/foo")
	}
	if !dirty {
		t.Error("dirty should be true")
	}
}

func TestCacheMissWhenFileAbsent(t *testing.T) {
	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "nonexistent-cache")

	stale := isCacheStale(cacheFile, 5*time.Second)
	if !stale {
		t.Error("absent cache file should be considered stale")
	}
}

func TestCacheMissWhenOld(t *testing.T) {
	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "old-cache")

	// Write cache file then backdate its mtime
	writeCache(cacheFile, "main", false)
	oldTime := time.Now().Add(-10 * time.Second)
	if err := os.Chtimes(cacheFile, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	stale := isCacheStale(cacheFile, 5*time.Second)
	if !stale {
		t.Error("old cache file should be considered stale")
	}
}

func TestCacheFreshIsNotStale(t *testing.T) {
	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "fresh-cache")
	writeCache(cacheFile, "main", false)

	stale := isCacheStale(cacheFile, 5*time.Second)
	if stale {
		t.Error("fresh cache should NOT be stale")
	}
}

// ─── Non-git directory ────────────────────────────────────────────────────────

func TestNonGitDir(t *testing.T) {
	tmpDir := t.TempDir() // not a git repo
	cacheFile := filepath.Join(tmpDir, "test-cache")

	branch, dirty := Get(tmpDir, cacheFile, 5*time.Second)
	// Non-git dir should return empty branch and not dirty
	if branch != "" {
		t.Errorf("non-git dir: expected empty branch, got %q", branch)
	}
	if dirty {
		t.Error("non-git dir: expected dirty=false")
	}
}

// ─── Integration: real git repo ───────────────────────────────────────────────
// These tests run only when the test binary is executed inside a git repo.

func TestGetFromRealRepo(t *testing.T) {
	// Find a parent directory that is a git repo
	repoDir := findGitRepo(t)
	if repoDir == "" {
		t.Skip("no git repo found in parent directories")
	}

	cacheFile := filepath.Join(t.TempDir(), "real-cache")
	branch, _ := Get(repoDir, cacheFile, 5*time.Second)
	if branch == "" {
		t.Error("expected non-empty branch from real git repo")
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func findGitRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
