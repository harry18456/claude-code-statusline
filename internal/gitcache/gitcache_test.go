package gitcache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

func TestDefaultCacheFileDerivesDirectoryHash(t *testing.T) {
	tempRoot := setTestTempDir(t)
	dirA := filepath.Join(t.TempDir(), "repo-a")
	dirB := filepath.Join(t.TempDir(), "repo-b")
	if err := os.MkdirAll(dirA, 0o755); err != nil {
		t.Fatalf("mkdir dirA: %v", err)
	}
	if err := os.MkdirAll(dirB, 0o755); err != nil {
		t.Fatalf("mkdir dirB: %v", err)
	}

	cacheA := DefaultCacheFile(dirA)
	cacheAAgain := DefaultCacheFile(filepath.Join(dirA, "."))
	cacheB := DefaultCacheFile(dirB)

	if cacheA != cacheAAgain {
		t.Fatalf("same directory should derive same cache path: %q != %q", cacheA, cacheAAgain)
	}
	if cacheA == cacheB {
		t.Fatalf("different directories should derive different cache paths: %q", cacheA)
	}
	if filepath.Dir(cacheA) != tempRoot {
		t.Fatalf("cache path dir: got %q, want %q", filepath.Dir(cacheA), tempRoot)
	}

	base := filepath.Base(cacheA)
	const prefix = "claude-statusline-git-cache-"
	if !strings.HasPrefix(base, prefix) {
		t.Fatalf("cache filename %q should have prefix %q", base, prefix)
	}
	hash := strings.TrimPrefix(base, prefix)
	if len(hash) != 32 {
		t.Fatalf("hash length: got %d, want 32", len(hash))
	}
	if hash != strings.ToLower(hash) || !isLowerHex(hash) {
		t.Fatalf("hash should be lowercase hex, got %q", hash)
	}
}

// ─── Non-git directory ────────────────────────────────────────────────────────

func TestNonGitDir(t *testing.T) {
	tmpDir := t.TempDir() // not a git repo
	cacheFile := filepath.Join(t.TempDir(), "test-cache")

	branch, dirty := getCached(tmpDir, cacheFile, 5*time.Second)
	// Non-git dir should return empty branch and not dirty
	if branch != "" {
		t.Errorf("non-git dir: expected empty branch, got %q", branch)
	}
	if dirty {
		t.Error("non-git dir: expected dirty=false")
	}
}

func TestDirtyFromRunError(t *testing.T) {
	if dirtyFromRunError(nil) {
		t.Fatal("nil error should not be dirty")
	}

	diffErr := gitDiffNoIndexError(t)
	if !dirtyFromRunError(diffErr) {
		t.Fatalf("git diff exit 1 should be dirty: %v", diffErr)
	}

	invalidFlagErr := exec.Command("git", "diff", "--quiet", "--definitely-bogus-flag").Run()
	if invalidFlagErr == nil {
		t.Fatal("invalid git diff flag should fail")
	}
	if dirtyFromRunError(invalidFlagErr) {
		t.Fatalf("git diff execution error should not be dirty: %v", invalidFlagErr)
	}
}

func TestGetUsesDirectoryIsolatedCache(t *testing.T) {
	tempRoot := setTestTempDir(t)
	repoA := initGitRepo(t, "main")
	repoB := initGitRepo(t, "feature/cache")

	branchA, dirtyA := Get(repoA, 5*time.Second)
	if branchA != "main" {
		t.Fatalf("repoA branch: got %q, want %q", branchA, "main")
	}
	if dirtyA {
		t.Fatal("repoA should not be dirty")
	}

	branchB, dirtyB := Get(repoB, 5*time.Second)
	if branchB != "feature/cache" {
		t.Fatalf("repoB branch: got %q, want %q", branchB, "feature/cache")
	}
	if dirtyB {
		t.Fatal("repoB should not be dirty")
	}

	branchA, dirtyA = Get(repoA, 5*time.Second)
	if branchA != "main" {
		t.Fatalf("repoA cached branch: got %q, want %q", branchA, "main")
	}
	if dirtyA {
		t.Fatal("repoA cached result should not be dirty")
	}

	cacheA := DefaultCacheFile(repoA)
	cacheB := DefaultCacheFile(repoB)
	if cacheA == cacheB {
		t.Fatalf("repos should not share cache path: %q", cacheA)
	}
	if filepath.Dir(cacheA) != tempRoot || filepath.Dir(cacheB) != tempRoot {
		t.Fatalf("cache paths should stay in test temp dir: %q, %q; want dir %q", cacheA, cacheB, tempRoot)
	}
}

func TestGetConcurrentRefreshesUseAtomicCache(t *testing.T) {
	tempRoot := setTestTempDir(t)
	repo := initGitRepo(t, "main")

	const goroutines = 8
	const iterations = 5
	errCh := make(chan error, goroutines*iterations)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				branch, dirty := Get(repo, -time.Second)
				if branch != "main" {
					errCh <- fmt.Errorf("branch: got %q, want %q", branch, "main")
					return
				}
				if dirty {
					errCh <- fmt.Errorf("dirty: got true, want false")
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}

	cacheFile := DefaultCacheFile(repo)
	if filepath.Dir(cacheFile) != tempRoot {
		t.Fatalf("cache path should stay in test temp dir: got %q, want dir %q", cacheFile, tempRoot)
	}
	branch, dirty, err := readCache(cacheFile)
	if err != nil {
		t.Fatalf("readCache after concurrent refreshes: %v", err)
	}
	if branch != "main" {
		t.Fatalf("cached branch: got %q, want %q", branch, "main")
	}
	if dirty {
		t.Fatal("cached dirty should be false")
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
	branch, _ := getCached(repoDir, cacheFile, 5*time.Second)
	if branch == "" {
		t.Error("expected non-empty branch from real git repo")
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func gitDiffNoIndexError(t *testing.T) error {
	t.Helper()
	dir := t.TempDir()
	left := filepath.Join(dir, "left.txt")
	right := filepath.Join(dir, "right.txt")
	if err := os.WriteFile(left, []byte("left\n"), 0o644); err != nil {
		t.Fatalf("write left file: %v", err)
	}
	if err := os.WriteFile(right, []byte("right\n"), 0o644); err != nil {
		t.Fatalf("write right file: %v", err)
	}

	err := exec.Command("git", "diff", "--quiet", "--no-index", left, right).Run()
	if err == nil {
		t.Fatal("git diff --no-index should report differences")
	}
	return err
}

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

func setTestTempDir(t *testing.T) string {
	t.Helper()
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	t.Setenv("TMP", tempRoot)
	t.Setenv("TEMP", tempRoot)
	return tempRoot
}

func initGitRepo(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "symbolic-ref", "HEAD", "refs/heads/"+branch)
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	allArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", allArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func isLowerHex(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
