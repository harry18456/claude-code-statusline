package model

import (
	"strings"
	"testing"
)

func TestParsePayload_Normal(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"main"},"rate_limits":{"five_hour":{"used_percentage":15},"seven_day":{"used_percentage":8}}}`

	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Model.DisplayName != "Claude Opus 4.6" {
		t.Errorf("model display name: got %q, want %q", p.Model.DisplayName, "Claude Opus 4.6")
	}
	if p.ContextWindow.UsedPercentage != 42 {
		t.Errorf("ctx pct: got %v, want 42", p.ContextWindow.UsedPercentage)
	}
	if p.ContextWindow.ContextWindowSize != 1000000 {
		t.Errorf("ctx size: got %v, want 1000000", p.ContextWindow.ContextWindowSize)
	}
	if p.Cost.TotalCostUSD != 0.85 {
		t.Errorf("cost: got %v, want 0.85", p.Cost.TotalCostUSD)
	}
	if p.Cost.TotalLinesAdded != 150 {
		t.Errorf("lines added: got %v, want 150", p.Cost.TotalLinesAdded)
	}
	if p.Cost.TotalLinesRemoved != 30 {
		t.Errorf("lines removed: got %v, want 30", p.Cost.TotalLinesRemoved)
	}
	if p.Cost.TotalDurationMs != 222000 {
		t.Errorf("duration ms: got %v, want 222000", p.Cost.TotalDurationMs)
	}
	if p.Workspace.CurrentDir != "/Users/dev/my-project" {
		t.Errorf("workspace dir: got %q, want %q", p.Workspace.CurrentDir, "/Users/dev/my-project")
	}
	if p.Worktree.Branch != "main" {
		t.Errorf("branch: got %q, want %q", p.Worktree.Branch, "main")
	}
	if p.RateLimits.FiveHour.UsedPercentage != 15 {
		t.Errorf("5h rate: got %v, want 15", p.RateLimits.FiveHour.UsedPercentage)
	}
	if p.RateLimits.SevenDay.UsedPercentage != 8 {
		t.Errorf("7d rate: got %v, want 8", p.RateLimits.SevenDay.UsedPercentage)
	}
}

func TestParsePayload_Agent(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"main"},"agent":{"name":"code-reviewer"}}`

	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Agent.Name != "code-reviewer" {
		t.Errorf("agent name: got %q, want %q", p.Agent.Name, "code-reviewer")
	}
}

func TestParsePayload_Worktree(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_lines_added":150,"total_lines_removed":30,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"worktree-my-feature","name":"my-feature","path":"/path/to/worktree"}}`

	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Worktree.Name != "my-feature" {
		t.Errorf("worktree name: got %q, want %q", p.Worktree.Name, "my-feature")
	}
}

func TestParsePayload_MissingFields(t *testing.T) {
	// Minimal payload — only model
	json := `{"model":{"display_name":"Opus 4.6 (1M context)"},"context_window":{"used_percentage":0,"context_window_size":1000000},"cost":{"total_cost_usd":0,"total_duration_ms":0},"workspace":{"current_dir":"/Users/dev/my-project"}}`

	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Model.DisplayName != "Opus 4.6 (1M context)" {
		t.Errorf("model: got %q", p.Model.DisplayName)
	}
	// Missing fields should be zero values
	if p.Agent.Name != "" {
		t.Errorf("agent name should be empty, got %q", p.Agent.Name)
	}
	if p.Worktree.Name != "" {
		t.Errorf("worktree name should be empty, got %q", p.Worktree.Name)
	}
	// rate_limits absent — UsedPercentage should be 0 (zero value)
	if p.RateLimits.FiveHour.UsedPercentage != 0 {
		t.Errorf("5h rate should be 0 when absent, got %v", p.RateLimits.FiveHour.UsedPercentage)
	}
}

func TestParsePayload_EmptyInput(t *testing.T) {
	_, err := ParsePayload(strings.NewReader(""))
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestParsePayload_InvalidJSON(t *testing.T) {
	_, err := ParsePayload(strings.NewReader("{invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParsePayload_RateLimitAbsent(t *testing.T) {
	// rate_limits absent — check that HasFiveHour / HasSevenDay are false
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":75,"context_window_size":200000},"cost":{"total_cost_usd":3.20,"total_lines_added":280,"total_lines_removed":45,"total_duration_ms":725000},"workspace":{"current_dir":"/Users/dev/my-project"},"worktree":{"branch":"feat/auth"},"rate_limits":{"five_hour":{"used_percentage":48}}}`

	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.RateLimits.FiveHour.Present {
		t.Error("five_hour should be present")
	}
	if p.RateLimits.SevenDay.Present {
		t.Error("seven_day should NOT be present")
	}
}
