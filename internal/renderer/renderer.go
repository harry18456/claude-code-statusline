// Package renderer assembles the two-line ANSI status output for Claude Code.
package renderer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-code-statusline/internal/model"
)

// GitInfo carries branch and dirty-state resolved by the gitcache package.
type GitInfo struct {
	Branch string
	Dirty  bool
}

// Options controls rendering behavior driven by environment variables.
type Options struct {
	ASCIIMode bool // CLAUDE_STATUSLINE_ASCII=1
	NerdFont  bool // CLAUDE_STATUSLINE_NERDFONT=1
	Powerline bool // CLAUDE_STATUSLINE_POWERLINE=1 (or follows NerdFont)
	TrueColor bool // COLORTERM=truecolor|24bit
}

// DefaultOptions returns Options with all features disabled (safest fallback).
func DefaultOptions() Options {
	return Options{}
}

// ─── ANSI helpers ─────────────────────────────────────────────────────────────

const (
	ansiReset   = "\033[0m"
	ansiCyan    = "\033[36m"
	ansiBlue    = "\033[34m"
	ansiGray    = "\033[90m"
	ansiYellow  = "\033[33m"
	ansiGreen   = "\033[32m"
	ansiRed     = "\033[31m"
	ansiMagenta = "\033[35m"
)

func ansiRGB(r, g, b int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// Anthropic brand purple RGB(114,102,234).
const ansiPurple = "\033[35m" // ANSI fallback

func purpleCode(trueColor bool) string {
	if trueColor {
		return ansiRGB(114, 102, 234)
	}
	return ansiPurple
}

// ─── Symbols ──────────────────────────────────────────────────────────────────

type symbols struct {
	brand  string
	branch string
	warn   string
	time   string
	cost   string
	sep    string
}

func symbolSet(opts Options) symbols {
	if opts.ASCIIMode {
		return symbols{
			brand:  "<>",
			branch: ">",
			warn:   "!",
			time:   "",
			cost:   "",
			sep:    " | ",
		}
	}
	sep := " │ "
	if opts.Powerline {
		sep = " \ue0b0 "
	}
	if opts.NerdFont {
		return symbols{
			brand:  "◆",
			branch: " ",
			warn:   " \uf026", // nf-mdi-alert  (󰀦)
			time:   "\uf017 ", // nf-fa-clock_o (󰔟)
			cost:   " ",       // nf-fa-money
			sep:    sep,
		}
	}
	return symbols{
		brand:  "◆",
		branch: "⎇ ",
		warn:   " ⚠",
		time:   "",
		cost:   "",
		sep:    sep,
	}
}

// ─── Progress bar ─────────────────────────────────────────────────────────────

// Gradient colors (10 cells, green → yellow → red).
var gradR = [10]int{46, 116, 186, 241, 239, 236, 233, 231, 211, 192}
var gradG = [10]int{204, 195, 186, 196, 161, 126, 101, 76, 66, 57}
var gradB = [10]int{113, 89, 64, 15, 24, 34, 44, 60, 50, 43}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func buildBar(pct int, opts Options) string {
	filled := clamp(pct/10, 0, 10)

	if opts.ASCIIMode {
		var sb strings.Builder
		for i := range 10 {
			if i < filled {
				sb.WriteByte('#')
			} else {
				sb.WriteByte('-')
			}
		}
		return sb.String()
	}

	if opts.TrueColor {
		var sb strings.Builder
		for i := range 10 {
			if i < filled {
				sb.WriteString(ansiRGB(gradR[i], gradG[i], gradB[i]))
				sb.WriteString("█")
			} else {
				sb.WriteString(ansiRGB(60, 60, 60))
				sb.WriteString("░")
			}
		}
		sb.WriteString(ansiReset)
		return sb.String()
	}

	// ANSI fallback: solid color
	var barColor string
	switch {
	case pct >= 90:
		barColor = ansiRed
	case pct >= 70:
		barColor = ansiYellow
	default:
		barColor = ansiGreen
	}

	var bar strings.Builder
	for i := range 10 {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	return barColor + bar.String() + ansiReset
}

func pctColor(pct int) string {
	switch {
	case pct >= 90:
		return ansiRed
	case pct >= 70:
		return ansiYellow
	default:
		return ansiGreen
	}
}

// ─── Cost ─────────────────────────────────────────────────────────────────────

func costColor(usd float64) string {
	switch {
	case usd >= 10:
		return ansiRed
	case usd > 0:
		return ansiYellow
	default:
		return ansiGray
	}
}

// ─── Duration ─────────────────────────────────────────────────────────────────

func formatDuration(ms int64) string {
	if ms <= 0 {
		return ""
	}
	totalSec := ms / 1000
	mins := totalSec / 60
	secs := totalSec % 60
	if mins == 0 && secs == 0 {
		return "" // sub-second: suppress
	}
	return fmt.Sprintf("%dm%ds", mins, secs)
}

// ─── Context window label ─────────────────────────────────────────────────────

func ctxLabel(size int64, modelName string, exceeds200k bool) string {
	switch {
	case size >= 1_000_000:
		color := ansiGray
		if exceeds200k {
			color = ansiRed
		}
		return " " + color + "1M" + ansiReset
	case size >= 200_000:
		return " " + ansiGray + "200k" + ansiReset
	default:
		return ""
	}
}

// ─── Directory display ───────────────────────────────────────────────────────

// resolveProjectRoot determines the project root via a three-step fallback:
//  1. If payloadProjectDir is a strict ancestor of currentDir, return it.
//  2. Otherwise walk upward from currentDir; the first directory containing a
//     .git entry (file or directory) wins.
//  3. Otherwise return "" (no root detected).
//
// The walk uses os.Stat only — no git CLI, so it works without git installed
// and handles worktrees/submodules (where .git is a file) transparently.
func resolveProjectRoot(currentDir, payloadProjectDir string) string {
	if currentDir == "" {
		return ""
	}
	if payloadProjectDir != "" && payloadProjectDir != currentDir {
		rel, err := filepath.Rel(payloadProjectDir, currentDir)
		if err == nil && rel != "." && rel != "" && !strings.HasPrefix(rel, "..") {
			return payloadProjectDir
		}
	}
	p := currentDir
	for {
		if _, err := os.Stat(filepath.Join(p, ".git")); err == nil {
			return p
		}
		parent := filepath.Dir(p)
		if parent == p {
			return ""
		}
		p = parent
	}
}

// directoryDisplay returns the string shown on line 2 for the working directory.
// Resolves a project root via resolveProjectRoot, then displays
// "<root_base>" when current equals root, or "<root_base>/<relative>" when
// current is a descendant. Falls back to filepath.Base(currentDir) when no
// root is resolved. Returns "." for empty or "." current.
func directoryDisplay(currentDir, projectDir string) string {
	if currentDir == "" || currentDir == "." {
		return "."
	}
	root := resolveProjectRoot(currentDir, projectDir)
	if root == "" {
		return filepath.Base(currentDir)
	}
	if currentDir == root {
		return filepath.Base(root)
	}
	rel, err := filepath.Rel(root, currentDir)
	if err != nil || rel == "." || rel == "" || strings.HasPrefix(rel, "..") {
		return filepath.Base(currentDir)
	}
	return filepath.Base(root) + "/" + filepath.ToSlash(rel)
}

// ─── Rate limits ──────────────────────────────────────────────────────────────

// formatCountdown returns the remaining time until resetsAt as a parenthesised
// string: "(Xd Yh)" if >= 24h, "(Xh Ym)" if >= 60 min, "(Ym)" if < 60 min, "(now)" if expired.
// Returns "" when resetsAt is zero (absent).
func formatCountdown(resetsAt int64) string {
	if resetsAt == 0 {
		return ""
	}
	remaining := time.Until(time.Unix(resetsAt, 0))
	if remaining <= 0 {
		return "(now)"
	}
	totalMin := int(remaining.Minutes())
	hours := totalMin / 60
	mins := totalMin % 60
	if hours >= 24 {
		days := hours / 24
		remHours := hours % 24
		return fmt.Sprintf("(%dd %dh)", days, remHours)
	}
	if hours > 0 {
		return fmt.Sprintf("(%dh %dm)", hours, mins)
	}
	return fmt.Sprintf("(%dm)", mins)
}

// Seven-day rate-limit window length in seconds. Used only to derive
// expected_pct and deviation; the pace indicator is shown for the entire
// window as long as resets_at is present and has not elapsed.
const sevenDayWindowSeconds = int64(604800)

// One day in seconds. expected_pct uses day-level granularity:
// elapsed_days = ceil(elapsed / dayLengthSeconds), clamped to [0, 7].
const dayLengthSeconds = int64(86400)

// computePaceArrow returns a colored pace indicator for a seven_day rate
// limit, or "" when no indicator should be shown. Any non-zero deviation
// produces a directional arrow with magnitude floored at 1; the neutral "≈"
// only appears on exact deviation == 0 (rare).
// Caller is responsible for passing only seven_day rate limits — the function
// does not self-check label.
func computePaceArrow(rl model.RateLimit, now time.Time, opts Options) string {
	if !rl.Present || rl.ResetsAt == 0 {
		return ""
	}
	remaining := rl.ResetsAt - now.Unix()
	if remaining <= 0 {
		return ""
	}
	elapsed := sevenDayWindowSeconds - remaining
	elapsedDays := min(int64(math.Ceil(float64(elapsed)/float64(dayLengthSeconds))), int64(7))
	expectedPct := float64(elapsedDays) * (100.0 / 7.0)
	deviation := rl.UsedPercentage - expectedPct
	magnitude := clamp(int(math.Round(min(math.Abs(deviation), 100))), 0, 100)
	if magnitude == 0 && deviation != 0 {
		magnitude = 1
	}
	switch {
	case deviation > 0:
		if opts.ASCIIMode {
			return fmt.Sprintf("^%d%%", magnitude)
		}
		return fmt.Sprintf("%s▲%d%%%s", ansiRed, magnitude, ansiReset)
	case deviation < 0:
		if opts.ASCIIMode {
			return fmt.Sprintf("v%d%%", magnitude)
		}
		return fmt.Sprintf("%s▼%d%%%s", ansiGray, magnitude, ansiReset)
	}
	if opts.ASCIIMode {
		return "~"
	}
	return ansiGray + "≈" + ansiReset
}

func formatRate(label string, rl model.RateLimit, now time.Time, opts Options) string {
	if !rl.Present {
		return ""
	}
	pct := clamp(int(math.Round(min(max(rl.UsedPercentage, 0), 100))), 0, 100)
	color := ansiGray
	if pct >= 80 {
		color = ansiRed
	}
	arrow := ""
	if label == "7d" {
		arrow = computePaceArrow(rl, now, opts)
	}
	arrowSegment := ""
	if arrow != "" {
		arrowSegment = " " + arrow
	}
	countdown := formatCountdown(rl.ResetsAt)
	if countdown != "" {
		return fmt.Sprintf("%s%s:%d%%%s%s %s", color, label, pct, ansiReset, arrowSegment, countdown)
	}
	return fmt.Sprintf("%s%s:%d%%%s%s", color, label, pct, ansiReset, arrowSegment)
}

// ─── Render ───────────────────────────────────────────────────────────────────

// Render produces the two status lines.
// It does NOT run git itself — the caller passes pre-resolved GitInfo.
func Render(p *model.Payload, git GitInfo, opts Options) (line1, line2 string) {
	sym := symbolSet(opts)

	// ── Percent ──────────────────────────────────────────────────────────────
	pct := clamp(int(p.ContextWindow.UsedPercentage), 0, 100)

	// ── Progress bar ──────────────────────────────────────────────────────────
	bar := buildBar(pct, opts)

	// ── Warning symbol ────────────────────────────────────────────────────────
	warnStr := ""
	if pct >= 90 {
		warnStr = ansiRed + sym.warn + ansiReset
	}

	// ── Context label ─────────────────────────────────────────────────────────
	label := ctxLabel(p.ContextWindow.ContextWindowSize, p.Model.DisplayName, p.ExceedsTokens200k)

	// ── Cost ──────────────────────────────────────────────────────────────────
	costFmt := fmt.Sprintf("%.2f", p.Cost.TotalCostUSD)
	cColor := costColor(p.Cost.TotalCostUSD)

	// ── Duration ──────────────────────────────────────────────────────────────
	durStr := formatDuration(p.Cost.TotalDurationMs)

	// ── Rate limits ───────────────────────────────────────────────────────────
	now := time.Now()
	r5h := formatRate("5h", p.RateLimits.FiveHour, now, opts)
	r7d := formatRate("7d", p.RateLimits.SevenDay, now, opts)

	rateParts := []string{}
	if r5h != "" {
		rateParts = append(rateParts, r5h)
	}
	if r7d != "" {
		rateParts = append(rateParts, r7d)
	}

	// ── Brand diamond ─────────────────────────────────────────────────────────
	purple := purpleCode(opts.TrueColor)

	// ── Line 1 ────────────────────────────────────────────────────────────────
	var l1 strings.Builder
	l1.WriteString(purple + sym.brand + ansiReset)
	l1.WriteString(" " + ansiCyan + p.Model.DisplayName + ansiReset)
	l1.WriteString(sym.sep + bar + " " + pctColor(pct) + fmt.Sprintf("%d%%", pct) + ansiReset)
	l1.WriteString(warnStr)
	l1.WriteString(label)
	l1.WriteString(sym.sep + cColor + sym.cost + "$" + costFmt + ansiReset)
	if durStr != "" {
		l1.WriteString(sym.sep + ansiGray + sym.time + durStr + ansiReset)
	}
	if len(rateParts) > 0 {
		l1.WriteString(sym.sep + strings.Join(rateParts, " "))
	}

	// ── Line 2 ────────────────────────────────────────────────────────────────
	var parts []string

	// Branch
	branch := git.Branch
	if branch == "" {
		branch = p.Worktree.Branch
	}
	if branch != "" {
		dirty := ""
		if git.Dirty {
			dirty = "*"
		}
		parts = append(parts, ansiGray+sym.branch+branch+dirty+ansiReset)
	}

	// Lines added/removed
	if p.Cost.TotalLinesAdded > 0 || p.Cost.TotalLinesRemoved > 0 {
		parts = append(parts, fmt.Sprintf("%s+%d%s/%s-%d%s",
			ansiGreen, p.Cost.TotalLinesAdded, ansiReset,
			ansiRed, p.Cost.TotalLinesRemoved, ansiReset))
	}

	// Dirname
	dir := directoryDisplay(p.Workspace.CurrentDir, p.Workspace.ProjectDir)
	parts = append(parts, ansiBlue+dir+ansiReset)

	// Agent / Worktree indicator
	if p.Worktree.Name != "" {
		parts = append(parts, ansiYellow+"⚙ worktree:"+p.Worktree.Name+ansiReset)
	} else if p.Agent.Name != "" {
		parts = append(parts, ansiYellow+"⚙ "+p.Agent.Name+ansiReset)
	}

	line1 = l1.String()
	line2 = strings.Join(parts, sym.sep)
	return line1, line2
}
