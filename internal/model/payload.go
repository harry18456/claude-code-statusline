package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type tolerantInt64 int64

func (t *tolerantInt64) UnmarshalJSON(b []byte) error {
	*t = 0
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		return nil
	}

	var i int64
	if err := json.Unmarshal(b, &i); err == nil {
		*t = tolerantInt64(i)
		return nil
	}

	var f float64
	if err := json.Unmarshal(b, &f); err == nil {
		*t = tolerantInt64(int64(f))
		return nil
	}

	return nil
}

type ContextWindow struct {
	UsedPercentage    float64 `json:"used_percentage"`
	ContextWindowSize int64   `json:"context_window_size"`
}

func (c *ContextWindow) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		*c = ContextWindow{}
		return nil
	}

	var raw struct {
		UsedPercentage    float64       `json:"used_percentage"`
		ContextWindowSize tolerantInt64 `json:"context_window_size"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	c.UsedPercentage = raw.UsedPercentage
	c.ContextWindowSize = int64(raw.ContextWindowSize)
	return nil
}

type Cost struct {
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
	TotalDurationMs   int64   `json:"total_duration_ms"`
}

func (c *Cost) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		*c = Cost{}
		return nil
	}

	var raw struct {
		TotalCostUSD      float64       `json:"total_cost_usd"`
		TotalLinesAdded   int           `json:"total_lines_added"`
		TotalLinesRemoved int           `json:"total_lines_removed"`
		TotalDurationMs   tolerantInt64 `json:"total_duration_ms"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	c.TotalCostUSD = raw.TotalCostUSD
	c.TotalLinesAdded = raw.TotalLinesAdded
	c.TotalLinesRemoved = raw.TotalLinesRemoved
	c.TotalDurationMs = int64(raw.TotalDurationMs)
	return nil
}

// RateLimit represents a single rate limit entry.
// Present is true only when the JSON field was explicitly included.
// ResetsAt is 0 when the field is absent.
type RateLimit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64
	Present        bool `json:"-"`
}

func (r *RateLimit) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		*r = RateLimit{}
		return nil
	}

	var raw struct {
		UsedPercentage float64       `json:"used_percentage"`
		ResetsAt       tolerantInt64 `json:"resets_at"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	r.UsedPercentage = raw.UsedPercentage
	r.ResetsAt = int64(raw.ResetsAt)
	r.Present = true
	return nil
}

// RateLimits holds the optional 5-hour and 7-day rate limit values.
type RateLimits struct {
	FiveHour RateLimit `json:"five_hour"`
	SevenDay RateLimit `json:"seven_day"`
}

// Payload is the top-level Claude Code statusLine JSON payload.
type Payload struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`

	ContextWindow ContextWindow `json:"context_window"`

	Cost Cost `json:"cost"`

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

	RateLimits RateLimits `json:"rate_limits"`

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

	var p Payload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	return &p, nil
}
