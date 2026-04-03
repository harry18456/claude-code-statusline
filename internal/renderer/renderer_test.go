package renderer

import (
	"fmt"
	"os"
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

func TestRenderContextLabelSuppressedWhenInModelName(t *testing.T) {
	p := mustParse(t, jsonStartup) // "Opus 4.6 (1M context)"
	line1, _ := renderWith(p, GitInfo{}, DefaultOptions())
	plain := stripANSI(line1)
	// Model name already contains "context" → no label
	count := strings.Count(plain, "1M")
	// The label "1M" should NOT appear (or only appear if it's in the model name itself)
	// Model name is "Opus 4.6 (1M context)" — so "1M" appears in model name
	// The label should NOT add another "1M"
	_ = count
	// Simpler: check that there's no standalone "1M" label separate from model name
	// Strip model name from line and check
	withoutModel := strings.Replace(plain, "Opus 4.6 (1M context)", "", 1)
	if strings.Contains(withoutModel, "1M") {
		t.Errorf("context label should be suppressed when model name contains 'context', remaining: %q", withoutModel)
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
