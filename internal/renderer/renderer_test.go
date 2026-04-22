package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"claude-code-statusline/internal/model"
)

// helper: parse a JSON string into Payload
func mustParse(t *testing.T, jsonStr string) *model.Payload {
	t.Helper()
	p, err := model.ParsePayload(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return p
}

// helper: render with a given mode
func renderWith(p *model.Payload, gitInfo GitInfo, opts Options) (string, string) {
	return Render(p, gitInfo, opts)
}

// ─── Scenario fixtures ───────────────────────────────────────────────────────

const jsonNormal = `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"main"},"rate_limits":{"five_hour":{"used_percentage":15},"seven_day":{"used_percentage":8}}}`

const jsonWarning = `{"model":{"display_name":"Claude Sonnet 4.6"},"context_window":{"used_percentage":75,"context_window_size":200000},"cost":{"total_cost_usd":3.20,"total_lines_added":280,"total_lines_removed":45,"total_duration_ms":725000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"feat/auth"},"rate_limits":{"five_hour":{"used_percentage":48}}}`

const jsonDanger = `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":92,"context_window_size":1000000},"cost":{"total_cost_usd":15.30,"total_lines_added":500,"total_lines_removed":120,"total_duration_ms":2712000},"workspace":{"current_dir":"/Users/dev/api-server"},"worktree":{"branch":"main"},"rate_limits":{"five_hour":{"used_percentage":85},"seven_day":{"used_percentage":62}}}`

const jsonStartup = `{"model":{"display_name":"Opus 4.6 (1M context)"},"context_window":{"used_percentage":0,"context_window_size":1000000},"cost":{"total_cost_usd":0,"total_duration_ms":0},"workspace":{"current_dir":"/Users/dev/my-project"}}`

const jsonAgent = `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"main"},"agent":{"name":"code-reviewer"}}`

const jsonWorktree = `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"worktree-my-feature","name":"my-feature","path":"/path/to/worktree"}}`

// ─── Line 1 structure ─────────────────────────────────────────────────────────

func TestRenderLine1ContainsModel(t *testing.T) {
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	if !strings.Contains(line1, "Claude Opus 4.6") {
		t.Errorf("line1 should contain model name, got: %q", stripANSI(line1))
	}
}

func TestRenderLine1ContainsCost(t *testing.T) {
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	if !strings.Contains(line1, "$0.85") {
		t.Errorf("line1 should contain cost, got: %q", stripANSI(line1))
	}
}

func TestRenderLine1ContainsDuration(t *testing.T) {
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	// 222000 ms = 3m42s
	if !strings.Contains(plain, "3m42s") {
		t.Errorf("line1 should contain duration 3m42s, got: %q", plain)
	}
}

func TestRenderLine1ContainsRateLimits(t *testing.T) {
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "5h:15%") {
		t.Errorf("line1 should contain 5h rate, got: %q", plain)
	}
	if !strings.Contains(plain, "7d:8%") {
		t.Errorf("line1 should contain 7d rate, got: %q", plain)
	}
}

// ─── Line 2 structure ─────────────────────────────────────────────────────────

func TestRenderLine2ContainsBranch(t *testing.T) {
	p := mustParse(t, jsonNormal)
	_, line2 := renderWith(p, GitInfo{Branch: "main", Dirty: false}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "main") {
		t.Errorf("line2 should contain branch, got: %q", plain)
	}
}

func TestRenderLine2ContainsDirname(t *testing.T) {
	p := mustParse(t, jsonNormal)
	_, line2 := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "my-project") {
		t.Errorf("line2 should contain dirname, got: %q", plain)
	}
}

func TestRenderLine2LinesAddedRemoved(t *testing.T) {
	p := mustParse(t, jsonNormal)
	_, line2 := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "+150") {
		t.Errorf("line2 should contain +150, got: %q", plain)
	}
	if !strings.Contains(plain, "-30") {
		t.Errorf("line2 should contain -30, got: %q", plain)
	}
}

// ─── Zero-value hiding ────────────────────────────────────────────────────────

func TestRenderStartupNoDuration(t *testing.T) {
	p := mustParse(t, jsonStartup)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	// 0 ms → no duration
	if strings.Contains(plain, "0m0s") || strings.Contains(plain, "m") {
		t.Errorf("startup should have no duration, got: %q", plain)
	}
}

func TestRenderStartupNoLines(t *testing.T) {
	p := mustParse(t, jsonStartup)
	_, line2 := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line2)
	if strings.Contains(plain, "+0") || strings.Contains(plain, "-0") {
		t.Errorf("startup should have no lines added/removed, got: %q", plain)
	}
}

func TestRenderStartupNoRateLimits(t *testing.T) {
	p := mustParse(t, jsonStartup)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if strings.Contains(plain, "5h:") || strings.Contains(plain, "7d:") {
		t.Errorf("startup should have no rate limits, got: %q", plain)
	}
}

// ─── Duration sub-minute suppression ─────────────────────────────────────────

func TestRenderDurationSubMinuteSuppressed(t *testing.T) {
	// 500 ms total → < 1 second → should be suppressed
	jsonShort := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":10,"context_window_size":1000000},"cost":{"total_cost_usd":0.01,"total_duration_ms":500},"workspace":{"current_dir":"/tmp/x"}}`
	p := mustParse(t, jsonShort)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if strings.Contains(plain, "0m0s") || strings.Contains(plain, "m0s") {
		t.Errorf("sub-second duration should be suppressed, got: %q", plain)
	}
}

// ─── Warning symbol ────────────────────────────────────────────────────────────

func TestRenderWarningSymbolAt90(t *testing.T) {
	p := mustParse(t, jsonDanger)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "⚠") {
		t.Errorf("danger scenario should show ⚠, got: %q", plain)
	}
}

func TestRenderNoWarningSymbolBelow90(t *testing.T) {
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if strings.Contains(plain, "⚠") {
		t.Errorf("normal scenario should NOT show ⚠, got: %q", plain)
	}
}

// ─── ASCII mode ────────────────────────────────────────────────────────────────

func TestRenderASCIIBar(t *testing.T) {
	p := mustParse(t, jsonNormal)
	opts := DefaultOptions()
	opts.ASCIIMode = true
	line1, _ := renderWith(p, GitInfo{}, opts)
	// 42% → 4 filled
	if !strings.Contains(line1, "####------") {
		t.Errorf("ASCII 42%% should be '####------', got: %q", line1)
	}
}

func TestRenderASCIIWarning(t *testing.T) {
	p := mustParse(t, jsonDanger)
	opts := DefaultOptions()
	opts.ASCIIMode = true
	line1, _ := renderWith(p, GitInfo{}, opts)
	if !strings.Contains(line1, "!") {
		t.Errorf("ASCII danger should show !, got: %q", line1)
	}
	if strings.Contains(line1, "⚠") {
		t.Errorf("ASCII mode should NOT use ⚠, got: %q", line1)
	}
}

// ─── Context window label ─────────────────────────────────────────────────────

func TestRenderContextLabel1M(t *testing.T) {
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "1M") {
		t.Errorf("1M context window label missing, got: %q", plain)
	}
}

func TestRenderContextLabelNotSuppressedWhenModelNameContainsContext(t *testing.T) {
	p := mustParse(t, jsonStartup) // "Opus 4.6 (1M context)"
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	// Model name already contains "context", but label SHALL still be shown.
	// Strip the model name substring so the remaining text must still contain "1M".
	withoutModel := strings.Replace(plain, "Opus 4.6 (1M context)", "", 1)
	if !strings.Contains(withoutModel, "1M") {
		t.Errorf("context label must not be suppressed by model name substring; remaining: %q", withoutModel)
	}
}

func TestCtxLabel_ModelNameContainsContextStillShows1M(t *testing.T) {
	got := ctxLabel(1_000_000, "Opus 4.7 (1M context)", false)
	if !strings.Contains(got, "1M") {
		t.Errorf("ctxLabel should emit 1M even when modelName contains 'context', got: %q", got)
	}
	if !strings.Contains(got, ansiGray) {
		t.Errorf("1M label should be gray when exceeds200k=false, got: %q", got)
	}
}

func TestCtxLabel_ModelNameContainsContextRedWhenExceeds(t *testing.T) {
	got := ctxLabel(1_000_000, "Opus 4.6 (1M context)", true)
	if !strings.Contains(got, "1M") {
		t.Errorf("ctxLabel should emit 1M even when modelName contains 'context', got: %q", got)
	}
	if !strings.Contains(got, ansiRed) {
		t.Errorf("1M label should be red when exceeds200k=true, got: %q", got)
	}
}

// ─── Agent and Worktree indicator ─────────────────────────────────────────────

func TestRenderAgentIndicator(t *testing.T) {
	p := mustParse(t, jsonAgent)
	_, line2 := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "code-reviewer") {
		t.Errorf("line2 should show agent name, got: %q", plain)
	}
}

func TestRenderWorktreeIndicator(t *testing.T) {
	p := mustParse(t, jsonWorktree)
	_, line2 := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "worktree:my-feature") {
		t.Errorf("line2 should show worktree, got: %q", plain)
	}
}

func TestRenderWorktreeTakesPriorityOverAgent(t *testing.T) {
	// Payload has both agent and worktree — worktree wins
	jsonBoth := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/x"},"worktree":{"branch":"main","name":"my-wt"},"agent":{"name":"some-agent"}}`
	p := mustParse(t, jsonBoth)
	_, line2 := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "worktree:my-wt") {
		t.Errorf("worktree should take priority, got: %q", plain)
	}
	if strings.Contains(plain, "some-agent") {
		t.Errorf("agent should NOT show when worktree present, got: %q", plain)
	}
}

// ─── Dirty marker ─────────────────────────────────────────────────────────────

func TestRenderDirtyMarker(t *testing.T) {
	p := mustParse(t, jsonNormal)
	_, line2 := renderWith(p, GitInfo{Branch: "main", Dirty: true}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "main*") {
		t.Errorf("dirty branch should show *, got: %q", plain)
	}
}

// ─── Rate limit coloring ──────────────────────────────────────────────────────

func TestRenderRateLimitAbove80IsRed(t *testing.T) {
	p := mustParse(t, jsonDanger) // 5h: 85%
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	// Check that the 5h rate appears with red ANSI code before it
	// Red ANSI = \033[31m
	if !strings.Contains(line1, "\033[31m") {
		t.Errorf("rate limit >= 80%% should be in red, line1: %q", line1)
	}
}

// ─── Cost color thresholds ────────────────────────────────────────────────────

func TestRenderCostZeroIsGray(t *testing.T) {
	p := mustParse(t, jsonStartup)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "$0.00") {
		t.Errorf("zero cost should still show, got: %q", plain)
	}
}

func TestRenderCostAbove10IsRed(t *testing.T) {
	p := mustParse(t, jsonDanger) // cost = 15.30
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "$15.30") {
		t.Errorf("cost not shown, got: %q", plain)
	}
}

// ─── Progress bar width ───────────────────────────────────────────────────────

func TestRenderBarWidth10(t *testing.T) {
	p := mustParse(t, jsonNormal)
	opts := DefaultOptions()
	opts.ASCIIMode = true
	line1, _ := renderWith(p, GitInfo{}, opts)
	// Count # and - characters in the bar
	barStart := strings.Index(line1, "#")
	if barStart == -1 {
		barStart = strings.Index(line1, "-")
	}
	if barStart == -1 {
		t.Fatal("no bar found in ASCII line1")
	}
	// The bar should be exactly "####------" (10 chars)
	bar := ""
	for _, ch := range line1[barStart:] {
		if ch == '#' || ch == '-' {
			bar += string(ch)
		} else {
			break
		}
	}
	if len(bar) != 10 {
		t.Errorf("bar width should be 10, got %d: %q", len(bar), bar)
	}
}

// ─── Nerd Font mode ────────────────────────────────────────────────────────────

func TestRenderNerdFontBranchSymbol(t *testing.T) {
	p := mustParse(t, jsonNormal)
	opts := DefaultOptions()
	opts.NerdFont = true
	_, line2 := renderWith(p, GitInfo{Branch: "main"}, opts)
	// Nerd Font branch symbol: " " (nf-dev-git_branch U+E0A0)
	if !strings.Contains(line2, " ") {
		t.Errorf("nerd font mode should use  branch symbol, got: %q", line2)
	}
}

// ─── Powerline separators ─────────────────────────────────────────────────────

func TestRenderPowerlineSeparator(t *testing.T) {
	p := mustParse(t, jsonNormal)
	opts := DefaultOptions()
	opts.NerdFont = true
	opts.Powerline = true
	line1, _ := renderWith(p, GitInfo{}, opts)
	// Powerline separator U+E0B0
	const powerlineSep = "\ue0b0"
	if !strings.Contains(line1, powerlineSep) {
		t.Errorf("powerline mode should use \ue0b0 separator, got: %q", line1)
	}
}

func TestRenderNoPowerlineByDefault(t *testing.T) {
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	const powerlineSep = "\ue0b0"
	if strings.Contains(line1, powerlineSep) {
		t.Errorf("default mode should NOT use powerline separator")
	}
	if !strings.Contains(line1, "│") {
		t.Errorf("default mode should use │ separator, got: %q", line1)
	}
}

// ─── 200k label ───────────────────────────────────────────────────────────────

func TestRenderContextLabel200k(t *testing.T) {
	p := mustParse(t, jsonWarning) // context_window_size: 200000
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "200k") {
		t.Errorf("200k label missing, got: %q", plain)
	}
}

// ─── Directory display (line 2) ──────────────────────────────────────────────

func TestDirectoryDisplay_CurrentEqualsProject(t *testing.T) {
	dir := filepath.Join(string(filepath.Separator)+"Users", "dev", "my-project")
	got := directoryDisplay(dir, dir)
	if got != "my-project" {
		t.Errorf("current==project: got %q, want %q", got, "my-project")
	}
}

func TestDirectoryDisplay_CurrentIsDescendant(t *testing.T) {
	project := filepath.Join(string(filepath.Separator)+"Users", "dev", "my-project")
	current := filepath.Join(project, "internal", "renderer")
	got := directoryDisplay(current, project)
	want := "my-project/internal/renderer"
	if got != want {
		t.Errorf("descendant: got %q, want %q", got, want)
	}
}

func TestDirectoryDisplay_ProjectEmpty(t *testing.T) {
	current := filepath.Join(string(filepath.Separator)+"Users", "dev", "my-project")
	got := directoryDisplay(current, "")
	if got != "my-project" {
		t.Errorf("empty project: got %q, want %q", got, "my-project")
	}
}

func TestDirectoryDisplay_CurrentEmpty(t *testing.T) {
	got := directoryDisplay("", "")
	if got != "." {
		t.Errorf("empty current: got %q, want %q", got, ".")
	}
}

func TestDirectoryDisplay_CurrentIsDot(t *testing.T) {
	got := directoryDisplay(".", "")
	if got != "." {
		t.Errorf("current=.: got %q, want %q", got, ".")
	}
}

func TestDirectoryDisplay_CurrentNotDescendant(t *testing.T) {
	// current is outside project — fallback to base of current
	project := filepath.Join(string(filepath.Separator)+"Users", "dev", "project-a")
	current := filepath.Join(string(filepath.Separator)+"Users", "dev", "project-b")
	got := directoryDisplay(current, project)
	if got != "project-b" {
		t.Errorf("non-descendant fallback: got %q, want %q", got, "project-b")
	}
}

// ─── resolveProjectRoot via directoryDisplay: git-root fallback ──────────────

func TestDirectoryDisplay_GitRootFromSubfolder(t *testing.T) {
	// tmpDir/proj/.git/ (dir) + tmpDir/proj/sub/
	tmp := t.TempDir()
	proj := filepath.Join(tmp, "proj")
	sub := filepath.Join(proj, "sub")
	if err := os.MkdirAll(filepath.Join(proj, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	// payload's project_dir == current (Claude Code subfolder-start behavior)
	got := directoryDisplay(sub, sub)
	want := "proj/sub"
	if got != want {
		t.Errorf("git-root walk: got %q, want %q", got, want)
	}
}

func TestDirectoryDisplay_GitRootAsFile(t *testing.T) {
	// .git as a regular file (worktree / submodule-style structure)
	tmp := t.TempDir()
	proj := filepath.Join(tmp, "wt")
	sub := filepath.Join(proj, "src")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, ".git"), []byte("gitdir: /elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := directoryDisplay(sub, sub)
	want := "wt/src"
	if got != want {
		t.Errorf(".git-as-file: got %q, want %q", got, want)
	}
}

func TestDirectoryDisplay_SubmoduleFirstGitWins(t *testing.T) {
	// parent/.git (dir) + parent/sub/.git (file) + parent/sub/src/
	// Expected: submodule (sub) wins; display is "sub/src", not "parent/sub/src"
	tmp := t.TempDir()
	parent := filepath.Join(tmp, "parent")
	sub := filepath.Join(parent, "sub")
	src := filepath.Join(sub, "src")
	if err := os.MkdirAll(filepath.Join(parent, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, ".git"), []byte("gitdir: ../.git/modules/sub\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := directoryDisplay(src, src)
	want := "sub/src"
	if got != want {
		t.Errorf("submodule first-git-wins: got %q, want %q", got, want)
	}
}

func TestDirectoryDisplay_NoGitFallbackToBase(t *testing.T) {
	// Pure non-git directory tree → walk finds nothing → base name fallback
	tmp := t.TempDir()
	folder := filepath.Join(tmp, "standalone", "deep", "folder")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	got := directoryDisplay(folder, folder)
	// May match base name OR (if a stray .git exists in ancestors of tmp) a
	// "<root>/standalone/deep/folder" path. In the latter case, skip.
	if got == "folder" {
		return // expected path
	}
	if strings.HasSuffix(got, "/standalone/deep/folder") {
		t.Skipf("ancestor of t.TempDir()=%q contains a .git entry; skipping (got=%q)", tmp, got)
	}
	t.Errorf("no-git fallback: got %q, want %q", got, "folder")
}

func TestDirectoryDisplay_PayloadAncestorBeatsGitWalk(t *testing.T) {
	// Payload project_dir is a strict ancestor → use it, do NOT walk for .git.
	// Setup: tmp/outer/inner/.git + tmp/outer/inner/src; payload=tmp/outer.
	// Without the "payload wins" rule, walk would pick "inner" and show "inner/src".
	// With the rule, result is "outer/inner/src".
	tmp := t.TempDir()
	outer := filepath.Join(tmp, "outer")
	inner := filepath.Join(outer, "inner")
	src := filepath.Join(inner, "src")
	if err := os.MkdirAll(filepath.Join(inner, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	got := directoryDisplay(src, outer)
	want := "outer/inner/src"
	if got != want {
		t.Errorf("payload-ancestor-wins: got %q, want %q", got, want)
	}
}

func TestRenderLine2UsesProjectRelative(t *testing.T) {
	// Use platform-appropriate paths via filepath.Join; JSON-escape backslashes.
	project := filepath.Join(string(filepath.Separator)+"Users", "dev", "my-project")
	current := filepath.Join(project, "internal", "renderer")
	projectJSON := strings.ReplaceAll(project, `\`, `\\`)
	currentJSON := strings.ReplaceAll(current, `\`, `\\`)
	jsonSubdir := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"` + currentJSON + `","project_dir":"` + projectJSON + `"}}`
	p := mustParse(t, jsonSubdir)
	_, line2 := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line2)
	if !strings.Contains(plain, "my-project/internal/renderer") {
		t.Errorf("line2 should show project-relative path, got: %q", plain)
	}
}

// ─── 1M label color based on exceeds_200k_tokens ─────────────────────────────

func TestRenderContextLabel1M_NotExceeding_IsGray(t *testing.T) {
	// jsonNormal: 1M context, no exceeds_200k_tokens field → gray
	p := mustParse(t, jsonNormal)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	if !strings.Contains(line1, ansiGray+"1M") {
		t.Errorf("1M without exceeds_200k_tokens should be gray, got: %q", line1)
	}
	if strings.Contains(line1, ansiRed+"1M") {
		t.Errorf("1M without exceeds_200k_tokens should NOT be red, got: %q", line1)
	}
}

func TestRenderContextLabel1M_Exceeding_IsRed(t *testing.T) {
	jsonExceed := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/x"},"exceeds_200k_tokens":true}`
	p := mustParse(t, jsonExceed)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	if !strings.Contains(line1, ansiRed+"1M") {
		t.Errorf("1M with exceeds_200k_tokens=true should be red, got: %q", line1)
	}
}

func TestRenderContextLabel200k_ExceedsFlagIgnored(t *testing.T) {
	// 200k label stays gray even if exceeds_200k_tokens is true
	jsonExceed200k := `{"model":{"display_name":"Claude Sonnet 4.6"},"context_window":{"used_percentage":75,"context_window_size":200000},"cost":{"total_cost_usd":3.20,"total_duration_ms":725000},"workspace":{"current_dir":"/tmp/x"},"exceeds_200k_tokens":true}`
	p := mustParse(t, jsonExceed200k)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	if !strings.Contains(line1, ansiGray+"200k") {
		t.Errorf("200k should remain gray regardless of exceeds flag, got: %q", line1)
	}
	if strings.Contains(line1, ansiRed+"200k") {
		t.Errorf("200k should NOT be red, got: %q", line1)
	}
}

// ─── Seven-day pace arrow unit tests ─────────────────────────────────────────

// computePaceArrow uses a 604800-second (7-day) window. Tests anchor "now"
// so the calculation is deterministic regardless of real clock.

func TestComputePaceArrow_OverPace(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	// elapsed = 2 days (172800s) → expected ≈ 28.57%; used=55 → deviation ≈ +26.43 → magnitude 26
	resetsAt := now.Add(5 * 24 * time.Hour).Unix()
	rl := model.RateLimit{UsedPercentage: 55, ResetsAt: resetsAt, Present: true}
	got := computePaceArrow(rl, now, DefaultOptions())
	if !strings.Contains(got, "▲26%") {
		t.Errorf("over-pace should contain ▲26%%, got: %q", got)
	}
	if !strings.Contains(got, ansiRed) {
		t.Errorf("over-pace should use ansiRed, got: %q", got)
	}
}

func TestComputePaceArrow_UnderPace(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	// elapsed = 2 days → expected ≈ 28.57%; used=20 → deviation ≈ -8.57 → magnitude 9
	resetsAt := now.Add(5 * 24 * time.Hour).Unix()
	rl := model.RateLimit{UsedPercentage: 20, ResetsAt: resetsAt, Present: true}
	got := computePaceArrow(rl, now, DefaultOptions())
	if !strings.Contains(got, "▼9%") {
		t.Errorf("under-pace should contain ▼9%%, got: %q", got)
	}
	if !strings.Contains(got, ansiGray) {
		t.Errorf("under-pace should use ansiGray, got: %q", got)
	}
}

func TestComputePaceArrow_WithinTolerance(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	// elapsed = 2 days → expected ≈ 28.57%; used=30 → deviation ≈ +1.43 (|dev|≤5) → ≈
	resetsAt := now.Add(5 * 24 * time.Hour).Unix()
	rl := model.RateLimit{UsedPercentage: 30, ResetsAt: resetsAt, Present: true}
	got := computePaceArrow(rl, now, DefaultOptions())
	if !strings.Contains(got, "≈") {
		t.Errorf("within tolerance should contain ≈, got: %q", got)
	}
	if !strings.Contains(got, ansiGray) {
		t.Errorf("within tolerance should use ansiGray, got: %q", got)
	}
	if strings.ContainsAny(got, "▲▼") {
		t.Errorf("within tolerance should NOT contain an arrow, got: %q", got)
	}
}

func TestComputePaceArrow_WithinToleranceBoundary(t *testing.T) {
	// Exactly +5 and -5 deviation must still be "within tolerance" (not over/under-pace).
	now := time.Unix(1_000_000_000, 0)
	// elapsed = 2 days → expected ≈ 28.5714%
	resetsAt := now.Add(5 * 24 * time.Hour).Unix()
	const expected = 100.0 * 172800.0 / 604800.0 // ≈ 28.5714
	for _, dev := range []float64{-5, 0, +5} {
		rl := model.RateLimit{UsedPercentage: expected + dev, ResetsAt: resetsAt, Present: true}
		got := computePaceArrow(rl, now, DefaultOptions())
		if !strings.Contains(got, "≈") {
			t.Errorf("boundary deviation %v should contain ≈, got: %q", dev, got)
		}
		if strings.ContainsAny(got, "▲▼") {
			t.Errorf("boundary deviation %v should NOT contain an arrow, got: %q", dev, got)
		}
	}
}

func TestComputePaceArrow_NearReset(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	// remaining = 60000s < 60480s (10% of 604800) → suppressed regardless of deviation
	resetsAt := now.Unix() + 60000
	rl := model.RateLimit{UsedPercentage: 99, ResetsAt: resetsAt, Present: true}
	got := computePaceArrow(rl, now, DefaultOptions())
	if got != "" {
		t.Errorf("near-reset should suppress arrow, got: %q", got)
	}
}

func TestComputePaceArrow_ResetsAtAbsent(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	rl := model.RateLimit{UsedPercentage: 80, ResetsAt: 0, Present: true}
	got := computePaceArrow(rl, now, DefaultOptions())
	if got != "" {
		t.Errorf("resets_at=0 should suppress arrow, got: %q", got)
	}
}

func TestComputePaceArrow_ASCIIOverPace(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	resetsAt := now.Add(5 * 24 * time.Hour).Unix()
	rl := model.RateLimit{UsedPercentage: 55, ResetsAt: resetsAt, Present: true}
	opts := DefaultOptions()
	opts.ASCIIMode = true
	got := computePaceArrow(rl, now, opts)
	if got != "^26%" {
		t.Errorf("ASCII over-pace should be '^26%%' with no color, got: %q", got)
	}
}

func TestComputePaceArrow_ASCIIUnderPace(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	resetsAt := now.Add(5 * 24 * time.Hour).Unix()
	rl := model.RateLimit{UsedPercentage: 20, ResetsAt: resetsAt, Present: true}
	opts := DefaultOptions()
	opts.ASCIIMode = true
	got := computePaceArrow(rl, now, opts)
	if got != "v9%" {
		t.Errorf("ASCII under-pace should be 'v9%%' with no color, got: %q", got)
	}
}

func TestComputePaceArrow_ASCIIWithinTolerance(t *testing.T) {
	now := time.Unix(1_000_000_000, 0)
	// elapsed = 2 days → expected ≈ 28.57%; used=30 → deviation ≈ +1.43 → ~
	resetsAt := now.Add(5 * 24 * time.Hour).Unix()
	rl := model.RateLimit{UsedPercentage: 30, ResetsAt: resetsAt, Present: true}
	opts := DefaultOptions()
	opts.ASCIIMode = true
	got := computePaceArrow(rl, now, opts)
	if got != "~" {
		t.Errorf("ASCII within tolerance should be '~' with no color, got: %q", got)
	}
}

func TestRenderSevenDayOverPaceArrow(t *testing.T) {
	// Integration: 7d over-pace should show "▲<N>%" between % and countdown.
	// Use real-clock-based resetsAt that is ~5 days in the future.
	// elapsed ≈ 2 days (~28.57% expected); used=55 → deviation ≈ +26.43 → magnitude 26.
	resetsAt := time.Now().Add(5 * 24 * time.Hour).Unix()
	jsonPace := fmt.Sprintf(`{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/x"},"rate_limits":{"seven_day":{"used_percentage":55,"resets_at":%d}}}`, resetsAt)
	p := mustParse(t, jsonPace)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "7d:55% ▲26%") {
		t.Errorf("7d over-pace should render '7d:55%% ▲26%%', got: %q", plain)
	}
}

func TestRenderSevenDayWithinToleranceApprox(t *testing.T) {
	// Integration: 7d within tolerance should show " ≈" between % and countdown.
	// elapsed ≈ 2 days (~28.57% expected); used=30 → deviation ≈ +1.43 (|dev|≤5) → ≈.
	resetsAt := time.Now().Add(5 * 24 * time.Hour).Unix()
	jsonPace := fmt.Sprintf(`{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/x"},"rate_limits":{"seven_day":{"used_percentage":30,"resets_at":%d}}}`, resetsAt)
	p := mustParse(t, jsonPace)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if !strings.Contains(plain, "7d:30% ≈") {
		t.Errorf("7d within tolerance should render '7d:30%% ≈', got: %q", plain)
	}
	if strings.ContainsAny(plain, "▲▼") {
		t.Errorf("within tolerance should NOT contain arrow, got: %q", plain)
	}
}

func TestRenderFiveHourNeverShowsArrow(t *testing.T) {
	// Even with over-pace-like usage, 5h never shows an arrow.
	resetsAt := time.Now().Add(2 * time.Hour).Unix()
	jsonPace := fmt.Sprintf(`{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/x"},"rate_limits":{"five_hour":{"used_percentage":90,"resets_at":%d}}}`, resetsAt)
	p := mustParse(t, jsonPace)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	if strings.Contains(plain, "▲") || strings.Contains(plain, "▼") {
		t.Errorf("5h should never have arrow, got: %q", plain)
	}
}

// ─── formatCountdown unit tests ───────────────────────────────────────────────

func TestFormatCountdown_AboveOneDay(t *testing.T) {
	// 50 hours from now → (2d 2h)
	resetsAt := time.Now().Add(50 * time.Hour).Unix()
	result := formatCountdown(resetsAt)
	if !strings.Contains(result, "d") {
		t.Errorf(">=24h should show days, got: %q", result)
	}
	if !strings.Contains(result, "h") {
		t.Errorf(">=24h should show hours, got: %q", result)
	}
	if !strings.HasPrefix(result, "(") || !strings.HasSuffix(result, ")") {
		t.Errorf("result should be wrapped in parens, got: %q", result)
	}
}

func TestFormatCountdown_AboveOneHour(t *testing.T) {
	// 90 minutes from now → should show (Xh Ym)
	resetsAt := time.Now().Add(90 * time.Minute).Unix()
	result := formatCountdown(resetsAt)
	if !strings.Contains(result, "h") || !strings.Contains(result, "m") {
		t.Errorf(">=60 min should show (Xh Ym), got: %q", result)
	}
	if !strings.HasPrefix(result, "(") || !strings.HasSuffix(result, ")") {
		t.Errorf("result should be wrapped in parens, got: %q", result)
	}
}

func TestFormatCountdown_BelowOneHour(t *testing.T) {
	// 30 minutes from now → should show (Ym), no hours
	resetsAt := time.Now().Add(30 * time.Minute).Unix()
	result := formatCountdown(resetsAt)
	if strings.Contains(result, "h") {
		t.Errorf("<60 min should NOT show hours, got: %q", result)
	}
	if !strings.Contains(result, "m") {
		t.Errorf("<60 min should show minutes, got: %q", result)
	}
	if !strings.HasPrefix(result, "(") || !strings.HasSuffix(result, ")") {
		t.Errorf("result should be wrapped in parens, got: %q", result)
	}
}

func TestFormatCountdown_Expired(t *testing.T) {
	// In the past → should show (now)
	resetsAt := time.Now().Add(-5 * time.Minute).Unix()
	result := formatCountdown(resetsAt)
	if result != "(now)" {
		t.Errorf("expired should show (now), got: %q", result)
	}
}

// ─── Rate limit countdown integration ─────────────────────────────────────────

func TestRenderRateLimitCountdownShownAbove80(t *testing.T) {
	// rate limit >= 80% and resets_at present → countdown appears in line1
	futureTs := time.Now().Add(90 * time.Minute).Unix()
	jsonWith := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/x"},"rate_limits":{"five_hour":{"used_percentage":85,"resets_at":` + itoa(futureTs) + `}}}`
	p := mustParse(t, jsonWith)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	// Countdown should appear: either "h" or "m" in parentheses after percentage
	if !strings.Contains(plain, "(") || !strings.Contains(plain, ")") {
		t.Errorf("line1 should contain countdown in parens when rate >= 80%%, got: %q", plain)
	}
}

func TestRenderRateLimitCountdownShownBelow80(t *testing.T) {
	// rate limit < 80% but resets_at present → countdown always shown
	futureTs := time.Now().Add(90 * time.Minute).Unix()
	jsonWith := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/x"},"rate_limits":{"five_hour":{"used_percentage":50,"resets_at":` + itoa(futureTs) + `}}}`
	p := mustParse(t, jsonWith)
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	// Countdown should appear regardless of pct
	if !strings.Contains(plain, "(") || !strings.Contains(plain, ")") {
		t.Errorf("line1 should contain countdown even when rate < 80%%, got: %q", plain)
	}
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// stripANSI removes ANSI escape sequences for plain-text comparison.
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm'
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
			continue
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

// TestMain allows reading env vars for integration-style checks.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
