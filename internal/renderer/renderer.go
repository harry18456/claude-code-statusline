// Package renderer assembles the two-line ANSI status output for Claude Code.
package renderer

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

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

func ctxLabel(size int64, modelName string) string {
	nameLower := strings.ToLower(modelName)
	if strings.Contains(nameLower, "context") {
		return ""
	}
	switch {
	case size >= 1_000_000:
		return " " + ansiGray + "1M" + ansiReset
	case size >= 200_000:
		return " " + ansiGray + "200k" + ansiReset
	default:
		return ""
	}
}

// ─── Rate limits ──────────────────────────────────────────────────────────────

func formatRate(label string, rl model.RateLimit) string {
	if !rl.Present {
		return ""
	}
	pct := int(math.Round(rl.UsedPercentage))
	if pct >= 80 {
		return fmt.Sprintf("%s%s:%d%%%s", ansiRed, label, pct, ansiReset)
	}
	return fmt.Sprintf("%s%s:%d%%%s", ansiGray, label, pct, ansiReset)
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
	label := ctxLabel(p.ContextWindow.ContextWindowSize, p.Model.DisplayName)

	// ── Cost ──────────────────────────────────────────────────────────────────
	costFmt := fmt.Sprintf("%.2f", p.Cost.TotalCostUSD)
	cColor := costColor(p.Cost.TotalCostUSD)

	// ── Duration ──────────────────────────────────────────────────────────────
	durStr := formatDuration(p.Cost.TotalDurationMs)

	// ── Rate limits ───────────────────────────────────────────────────────────
	r5h := formatRate("5h", p.RateLimits.FiveHour)
	r7d := formatRate("7d", p.RateLimits.SevenDay)

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
	dir := p.Workspace.CurrentDir
	if dir == "" || dir == "." {
		dir = "."
	} else {
		dir = filepath.Base(dir)
	}
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
