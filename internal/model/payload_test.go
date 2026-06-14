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
	if !p.RateLimits.FiveHour.Present {
		t.Error("5h rate should be present")
	}
	if p.RateLimits.SevenDay.UsedPercentage != 8 {
		t.Errorf("7d rate: got %v, want 8", p.RateLimits.SevenDay.UsedPercentage)
	}
	if !p.RateLimits.SevenDay.Present {
		t.Error("7d rate should be present")
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

func TestParsePayload_CurrentUsage(t *testing.T) {
	tests := []struct {
		name              string
		json              string
		wantAvailable     bool
		wantInput         int64
		wantOutput        int64
		wantCacheCreation int64
		wantCacheRead     int64
	}{
		{
			name:              "object",
			json:              `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":73,"context_window_size":200000,"current_usage":{"input_tokens":1,"output_tokens":374,"cache_creation_input_tokens":1302,"cache_read_input_tokens":144198}},"cost":{"total_cost_usd":0.85},"workspace":{"current_dir":"/Users/dev/my-project"}}`,
			wantAvailable:     true,
			wantInput:         1,
			wantOutput:        374,
			wantCacheCreation: 1302,
			wantCacheRead:     144198,
		},
		{
			name:          "null",
			json:          `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":73,"context_window_size":200000,"current_usage":null},"cost":{"total_cost_usd":0.85},"workspace":{"current_dir":"/Users/dev/my-project"}}`,
			wantAvailable: false,
		},
		{
			name:          "absent",
			json:          `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":73,"context_window_size":200000},"cost":{"total_cost_usd":0.85},"workspace":{"current_dir":"/Users/dev/my-project"}}`,
			wantAvailable: false,
		},
		{
			name:              "decimal",
			json:              `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":73,"context_window_size":200000,"current_usage":{"input_tokens":1.0,"output_tokens":374.0,"cache_creation_input_tokens":1302.0,"cache_read_input_tokens":144198.0}},"cost":{"total_cost_usd":0.85},"workspace":{"current_dir":"/Users/dev/my-project"}}`,
			wantAvailable:     true,
			wantInput:         1,
			wantOutput:        374,
			wantCacheCreation: 1302,
			wantCacheRead:     144198,
		},
		{
			name:              "scientific",
			json:              `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":73,"context_window_size":200000,"current_usage":{"input_tokens":1e0,"output_tokens":3.74e2,"cache_creation_input_tokens":1.302e3,"cache_read_input_tokens":1.44198e5}},"cost":{"total_cost_usd":0.85},"workspace":{"current_dir":"/Users/dev/my-project"}}`,
			wantAvailable:     true,
			wantInput:         1,
			wantOutput:        374,
			wantCacheCreation: 1302,
			wantCacheRead:     144198,
		},
		{
			name:              "wrong scalar token fields",
			json:              `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":73,"context_window_size":200000,"current_usage":{"input_tokens":"wide","output_tokens":true,"cache_creation_input_tokens":1302,"cache_read_input_tokens":144198}},"cost":{"total_cost_usd":0.85},"workspace":{"current_dir":"/Users/dev/my-project"}}`,
			wantAvailable:     true,
			wantInput:         0,
			wantOutput:        0,
			wantCacheCreation: 1302,
			wantCacheRead:     144198,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePayload(strings.NewReader(tt.json))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := p.ContextWindow.CurrentUsage
			if !tt.wantAvailable {
				if got != nil {
					t.Fatalf("CurrentUsage should be nil, got %#v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("CurrentUsage should be available")
			}
			if got.InputTokens != tt.wantInput {
				t.Errorf("InputTokens: got %v, want %v", got.InputTokens, tt.wantInput)
			}
			if got.OutputTokens != tt.wantOutput {
				t.Errorf("OutputTokens: got %v, want %v", got.OutputTokens, tt.wantOutput)
			}
			if got.CacheCreationInputTokens != tt.wantCacheCreation {
				t.Errorf("CacheCreationInputTokens: got %v, want %v", got.CacheCreationInputTokens, tt.wantCacheCreation)
			}
			if got.CacheReadInputTokens != tt.wantCacheRead {
				t.Errorf("CacheReadInputTokens: got %v, want %v", got.CacheReadInputTokens, tt.wantCacheRead)
			}
		})
	}
}

func TestParsePayload_ResetsAtPresent(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":85,"context_window_size":200000},"cost":{"total_cost_usd":2.50,"total_duration_ms":300000},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":85,"resets_at":1700000000},"seven_day":{"used_percentage":62,"resets_at":1700100000}}}`

	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.RateLimits.FiveHour.ResetsAt != 1700000000 {
		t.Errorf("five_hour resets_at: got %v, want 1700000000", p.RateLimits.FiveHour.ResetsAt)
	}
	if p.RateLimits.SevenDay.ResetsAt != 1700100000 {
		t.Errorf("seven_day resets_at: got %v, want 1700100000", p.RateLimits.SevenDay.ResetsAt)
	}
}

func TestParsePayload_ResetsAtAbsent(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":85,"context_window_size":200000},"cost":{"total_cost_usd":2.50,"total_duration_ms":300000},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":85},"seven_day":{"used_percentage":62}}}`

	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.RateLimits.FiveHour.ResetsAt != 0 {
		t.Errorf("five_hour resets_at should be 0 when absent, got %v", p.RateLimits.FiveHour.ResetsAt)
	}
	if p.RateLimits.SevenDay.ResetsAt != 0 {
		t.Errorf("seven_day resets_at should be 0 when absent, got %v", p.RateLimits.SevenDay.ResetsAt)
	}
}

func TestParsePayload_TolerantIntegerNumbers(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		wantCtxSize    int64
		wantDuration   int64
		wantFiveReset  int64
		wantSevenReset int64
	}{
		{
			name:           "decimal",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000.0},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000.0},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":15,"resets_at":1700000000.0},"seven_day":{"used_percentage":8,"resets_at":1700100000.0}}}`,
			wantCtxSize:    1000000,
			wantDuration:   222000,
			wantFiveReset:  1700000000,
			wantSevenReset: 1700100000,
		},
		{
			name:           "scientific",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1e6},"cost":{"total_cost_usd":0.85,"total_duration_ms":2.22e5},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":15,"resets_at":1.7e9},"seven_day":{"used_percentage":8,"resets_at":1.7001e9}}}`,
			wantCtxSize:    1000000,
			wantDuration:   222000,
			wantFiveReset:  1700000000,
			wantSevenReset: 1700100000,
		},
		{
			name:           "fractional truncates",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000.9},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000.9},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":15,"resets_at":1700000000.9},"seven_day":{"used_percentage":8,"resets_at":1700100000.9}}}`,
			wantCtxSize:    1000000,
			wantDuration:   222000,
			wantFiveReset:  1700000000,
			wantSevenReset: 1700100000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePayload(strings.NewReader(tt.json))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Model.DisplayName != "Claude Opus 4.6" {
				t.Errorf("model display name: got %q, want %q", p.Model.DisplayName, "Claude Opus 4.6")
			}
			if p.Cost.TotalCostUSD != 0.85 {
				t.Errorf("cost: got %v, want 0.85", p.Cost.TotalCostUSD)
			}
			if p.ContextWindow.ContextWindowSize != tt.wantCtxSize {
				t.Errorf("ctx size: got %v, want %v", p.ContextWindow.ContextWindowSize, tt.wantCtxSize)
			}
			if p.Cost.TotalDurationMs != tt.wantDuration {
				t.Errorf("duration ms: got %v, want %v", p.Cost.TotalDurationMs, tt.wantDuration)
			}
			if !p.RateLimits.FiveHour.Present {
				t.Error("five_hour should be present")
			}
			if p.RateLimits.FiveHour.ResetsAt != tt.wantFiveReset {
				t.Errorf("five_hour resets_at: got %v, want %v", p.RateLimits.FiveHour.ResetsAt, tt.wantFiveReset)
			}
			if !p.RateLimits.SevenDay.Present {
				t.Error("seven_day should be present")
			}
			if p.RateLimits.SevenDay.ResetsAt != tt.wantSevenReset {
				t.Errorf("seven_day resets_at: got %v, want %v", p.RateLimits.SevenDay.ResetsAt, tt.wantSevenReset)
			}
		})
	}
}

func TestParsePayload_UnconvertibleIntegerFields(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		wantCtxSize    int64
		wantDuration   int64
		wantFiveUsed   float64
		wantFiveReset  int64
		wantSevenUsed  float64
		wantSevenReset int64
	}{
		{
			name:           "context size string",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":"wide"},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":15,"resets_at":1700000000}}}`,
			wantCtxSize:    0,
			wantDuration:   222000,
			wantFiveUsed:   15,
			wantFiveReset:  1700000000,
			wantSevenUsed:  0,
			wantSevenReset: 0,
		},
		{
			name:           "context size bool",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":true},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":15,"resets_at":1700000000}}}`,
			wantCtxSize:    0,
			wantDuration:   222000,
			wantFiveUsed:   15,
			wantFiveReset:  1700000000,
			wantSevenUsed:  0,
			wantSevenReset: 0,
		},
		{
			name:           "duration string",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":"slow"},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":15,"resets_at":1700000000}}}`,
			wantCtxSize:    1000000,
			wantDuration:   0,
			wantFiveUsed:   15,
			wantFiveReset:  1700000000,
			wantSevenUsed:  0,
			wantSevenReset: 0,
		},
		{
			name:           "duration bool",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":false},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":15,"resets_at":1700000000}}}`,
			wantCtxSize:    1000000,
			wantDuration:   0,
			wantFiveUsed:   15,
			wantFiveReset:  1700000000,
			wantSevenUsed:  0,
			wantSevenReset: 0,
		},
		{
			name:           "resets_at string",
			json:           `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"rate_limits":{"five_hour":{"used_percentage":85,"resets_at":"soon"},"seven_day":{"used_percentage":62,"resets_at":"later"}}}`,
			wantCtxSize:    1000000,
			wantDuration:   222000,
			wantFiveUsed:   85,
			wantFiveReset:  0,
			wantSevenUsed:  62,
			wantSevenReset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePayload(strings.NewReader(tt.json))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Model.DisplayName != "Claude Opus 4.6" {
				t.Errorf("model display name: got %q, want %q", p.Model.DisplayName, "Claude Opus 4.6")
			}
			if p.Cost.TotalCostUSD != 0.85 {
				t.Errorf("cost: got %v, want 0.85", p.Cost.TotalCostUSD)
			}
			if p.ContextWindow.ContextWindowSize != tt.wantCtxSize {
				t.Errorf("ctx size: got %v, want %v", p.ContextWindow.ContextWindowSize, tt.wantCtxSize)
			}
			if p.Cost.TotalDurationMs != tt.wantDuration {
				t.Errorf("duration ms: got %v, want %v", p.Cost.TotalDurationMs, tt.wantDuration)
			}
			if !p.RateLimits.FiveHour.Present {
				t.Error("five_hour should be present")
			}
			if p.RateLimits.FiveHour.UsedPercentage != tt.wantFiveUsed {
				t.Errorf("five_hour used_percentage: got %v, want %v", p.RateLimits.FiveHour.UsedPercentage, tt.wantFiveUsed)
			}
			if p.RateLimits.FiveHour.ResetsAt != tt.wantFiveReset {
				t.Errorf("five_hour resets_at: got %v, want %v", p.RateLimits.FiveHour.ResetsAt, tt.wantFiveReset)
			}
			if tt.wantSevenUsed == 0 && tt.wantSevenReset == 0 {
				if p.RateLimits.SevenDay.Present {
					t.Error("seven_day should not be present")
				}
				return
			}
			if !p.RateLimits.SevenDay.Present {
				t.Error("seven_day should be present")
			}
			if p.RateLimits.SevenDay.UsedPercentage != tt.wantSevenUsed {
				t.Errorf("seven_day used_percentage: got %v, want %v", p.RateLimits.SevenDay.UsedPercentage, tt.wantSevenUsed)
			}
			if p.RateLimits.SevenDay.ResetsAt != tt.wantSevenReset {
				t.Errorf("seven_day resets_at: got %v, want %v", p.RateLimits.SevenDay.ResetsAt, tt.wantSevenReset)
			}
		})
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

func TestParsePayload_ExceedsTokens200kTrue(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"exceeds_200k_tokens":true}`
	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.ExceedsTokens200k {
		t.Errorf("ExceedsTokens200k: got false, want true")
	}
}

func TestParsePayload_ExceedsTokens200kFalse(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"},"exceeds_200k_tokens":false}`
	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ExceedsTokens200k {
		t.Errorf("ExceedsTokens200k: got true, want false")
	}
}

func TestParsePayload_ExceedsTokens200kAbsent(t *testing.T) {
	json := `{"model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":1000000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/Users/dev/my-project"}}`
	p, err := ParsePayload(strings.NewReader(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ExceedsTokens200k {
		t.Errorf("ExceedsTokens200k (absent): got true, want false")
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
