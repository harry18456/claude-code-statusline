package model

import (
	"encoding/json"
	"fmt"
	"io"
)

// RateLimit represents a single rate limit entry.
// Present is true only when the JSON field was explicitly included.
// ResetsAt is 0 when the field is absent.
type RateLimit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64
	Present        bool `json:"-"`
}

type rateLimitRaw struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

// RateLimits holds the optional 5-hour and 7-day rate limit values.
type RateLimits struct {
	FiveHour RateLimit
	SevenDay RateLimit
}

type rateLimitsRaw struct {
	FiveHour *rateLimitRaw `json:"five_hour"`
	SevenDay *rateLimitRaw `json:"seven_day"`
}

// Payload is the top-level Claude Code statusLine JSON payload.
type Payload struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`

	ContextWindow struct {
		UsedPercentage    float64 `json:"used_percentage"`
		ContextWindowSize int64   `json:"context_window_size"`
	} `json:"context_window"`

	Cost struct {
		TotalCostUSD      float64 `json:"total_cost_usd"`
		TotalLinesAdded   int     `json:"total_lines_added"`
		TotalLinesRemoved int     `json:"total_lines_removed"`
		TotalDurationMs   int64   `json:"total_duration_ms"`
	} `json:"cost"`

	Workspace struct {
		CurrentDir string `json:"current_dir"`
		ProjectDir string `json:"project_dir"`
	} `json:"workspace"`

	Worktree struct {
		Branch string `json:"branch"`
		Name   string `json:"name"`
		Path   string `json:"path"`
	} `json:"worktree"`

	Agent struct {
		Name string `json:"name"`
	} `json:"agent"`

	RateLimits RateLimits

	ExceedsTokens200k bool `json:"exceeds_200k_tokens"`
}

// payloadJSON mirrors Payload but uses the raw rate-limit type for presence detection.
type payloadJSON struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`

	ContextWindow struct {
		UsedPercentage    float64 `json:"used_percentage"`
		ContextWindowSize int64   `json:"context_window_size"`
	} `json:"context_window"`

	Cost struct {
		TotalCostUSD      float64 `json:"total_cost_usd"`
		TotalLinesAdded   int     `json:"total_lines_added"`
		TotalLinesRemoved int     `json:"total_lines_removed"`
		TotalDurationMs   int64   `json:"total_duration_ms"`
	} `json:"cost"`

	Workspace struct {
		CurrentDir string `json:"current_dir"`
		ProjectDir string `json:"project_dir"`
	} `json:"workspace"`

	Worktree struct {
		Branch string `json:"branch"`
		Name   string `json:"name"`
		Path   string `json:"path"`
	} `json:"worktree"`

	Agent struct {
		Name string `json:"name"`
	} `json:"agent"`

	RateLimits rateLimitsRaw `json:"rate_limits"`

	ExceedsTokens200k bool `json:"exceeds_200k_tokens"`
}

// ParsePayload reads and parses a Claude Code JSON payload from r.
// Returns an error if the input is empty or contains invalid JSON.
func ParsePayload(r io.Reader) (*Payload, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	var raw payloadJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	p := &Payload{
		Model:             raw.Model,
		ContextWindow:     raw.ContextWindow,
		Cost:              raw.Cost,
		Workspace:         raw.Workspace,
		Worktree:          raw.Worktree,
		Agent:             raw.Agent,
		ExceedsTokens200k: raw.ExceedsTokens200k,
	}

	if raw.RateLimits.FiveHour != nil {
		p.RateLimits.FiveHour = RateLimit{
			UsedPercentage: raw.RateLimits.FiveHour.UsedPercentage,
			ResetsAt:       raw.RateLimits.FiveHour.ResetsAt,
			Present:        true,
		}
	}
	if raw.RateLimits.SevenDay != nil {
		p.RateLimits.SevenDay = RateLimit{
			UsedPercentage: raw.RateLimits.SevenDay.UsedPercentage,
			ResetsAt:       raw.RateLimits.SevenDay.ResetsAt,
			Present:        true,
		}
	}

	return p, nil
}
