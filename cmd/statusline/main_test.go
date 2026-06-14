package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"claude-code-statusline/internal/renderer"
)

const cliPayload = `{"model":{"display_name":"Claude Opus 4.6"},"effort":{"level":"high"},"thinking":{"enabled":true},"fast_mode":true,"context_window":{"used_percentage":73,"context_window_size":200000,"current_usage":{"input_tokens":1,"cache_creation_input_tokens":1302,"cache_read_input_tokens":144198}},"cost":{"total_cost_usd":0.85,"total_lines_added":10,"total_lines_removed":2,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/project"},"agent":{"name":"code-reviewer"},"rate_limits":{"five_hour":{"used_percentage":15},"seven_day":{"used_percentage":8}}}`

func TestParseCLIConfig(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		getenv func(string) string
		check  func(t *testing.T, cfg cliConfig)
	}{
		{
			name: "default options",
			check: func(t *testing.T, cfg cliConfig) {
				if cfg.Options.ASCIIMode || cfg.Options.NerdFont || cfg.Options.Powerline || cfg.Options.TrueColor {
					t.Fatalf("default options enabled unexpected modes: %+v", cfg.Options)
				}
				if len(cfg.Warnings) != 0 {
					t.Fatalf("default config warnings = %v, want none", cfg.Warnings)
				}
			},
		},
		{
			name: "ascii flag",
			args: []string{"--ascii"},
			check: func(t *testing.T, cfg cliConfig) {
				if !cfg.Options.ASCIIMode {
					t.Fatal("--ascii should enable ASCII mode")
				}
			},
		},
		{
			name: "nerdfont implies powerline",
			args: []string{"--nerdfont"},
			check: func(t *testing.T, cfg cliConfig) {
				if !cfg.Options.NerdFont || !cfg.Options.Powerline {
					t.Fatalf("--nerdfont should enable NerdFont and Powerline, got %+v", cfg.Options)
				}
			},
		},
		{
			name: "powerline only",
			args: []string{"--powerline"},
			check: func(t *testing.T, cfg cliConfig) {
				if cfg.Options.NerdFont || !cfg.Options.Powerline {
					t.Fatalf("--powerline should not enable NerdFont, got %+v", cfg.Options)
				}
			},
		},
		{
			name: "hide keys trim dedupe and merge",
			args: []string{"--hide", "effort, duration", "--hide=rate,effort"},
			check: func(t *testing.T, cfg cliConfig) {
				for _, key := range []renderer.SectionKey{renderer.SectionEffort, renderer.SectionDuration, renderer.SectionRate} {
					if !cfg.Options.HiddenSections[key] {
						t.Fatalf("hidden key %q missing from %+v", key, cfg.Options.HiddenSections)
					}
				}
			},
		},
		{
			name: "unknown hide key warns and keeps known keys",
			args: []string{"--hide", "effort,duraton,rate"},
			check: func(t *testing.T, cfg cliConfig) {
				if !cfg.Options.HiddenSections[renderer.SectionEffort] || !cfg.Options.HiddenSections[renderer.SectionRate] {
					t.Fatalf("known hide keys should still apply: %+v", cfg.Options.HiddenSections)
				}
				if cfg.Options.HiddenSections[renderer.SectionDuration] {
					t.Fatal("misspelled duraton must not hide duration")
				}
				requireWarning(t, cfg.Warnings, "duraton")
			},
		},
		{
			name:   "removed project env vars ignored",
			getenv: func(string) string { return "1" },
			check: func(t *testing.T, cfg cliConfig) {
				if cfg.Options.ASCIIMode || cfg.Options.NerdFont || cfg.Options.Powerline {
					t.Fatalf("CLAUDE_STATUSLINE_* style env values must be ignored, got %+v", cfg.Options)
				}
			},
		},
		{
			name: "colorterm truecolor",
			getenv: func(key string) string {
				if key == "COLORTERM" {
					return "truecolor"
				}
				return ""
			},
			check: func(t *testing.T, cfg cliConfig) {
				if !cfg.Options.TrueColor {
					t.Fatal("COLORTERM=truecolor should enable true color")
				}
			},
		},
		{
			name: "version flag",
			args: []string{"--version"},
			check: func(t *testing.T, cfg cliConfig) {
				if !cfg.ShowVersion {
					t.Fatal("--version should set ShowVersion")
				}
			},
		},
		{
			name: "ascii conflict falls back to ascii",
			args: []string{"--ascii", "--nerdfont", "--powerline"},
			check: func(t *testing.T, cfg cliConfig) {
				if !cfg.Options.ASCIIMode || cfg.Options.NerdFont || cfg.Options.Powerline {
					t.Fatalf("ASCII conflict should keep only ASCII, got %+v", cfg.Options)
				}
				requireWarning(t, cfg.Warnings, "conflicts")
			},
		},
		{
			name: "unknown flag warns without hard failure",
			args: []string{"--unknown"},
			check: func(t *testing.T, cfg cliConfig) {
				requireWarning(t, cfg.Warnings, "invalid command-line config")
			},
		},
		{
			name: "missing hide value warns without hard failure",
			args: []string{"--hide"},
			check: func(t *testing.T, cfg cliConfig) {
				requireWarning(t, cfg.Warnings, "invalid command-line config")
			},
		},
		{
			name: "positional argument warns without hard failure",
			args: []string{"extra"},
			check: func(t *testing.T, cfg cliConfig) {
				requireWarning(t, cfg.Warnings, "positional")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := tt.getenv
			if getenv == nil {
				getenv = func(string) string { return "" }
			}
			cfg := parseCLIConfig(tt.args, getenv)
			tt.check(t, cfg)
		})
	}
}

func TestMainToleratesCLIConfig(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOut    []string
		absentOut  []string
		wantErr    string
		wantNoErr  bool
		wantNoRead bool
	}{
		{
			name:      "unknown hide key still renders",
			args:      []string{"--hide", "duraton"},
			wantOut:   []string{"Claude Opus 4.6", "3m42s"},
			wantErr:   "duraton",
			wantNoErr: false,
		},
		{
			name:      "ascii nerdfont conflict renders ascii",
			args:      []string{"--ascii", "--nerdfont"},
			wantOut:   []string{"<>", "#######---", "effort:high think fast"},
			wantErr:   "conflicts",
			wantNoErr: false,
		},
		{
			name:      "unknown flag still renders",
			args:      []string{"--unknown"},
			wantOut:   []string{"Claude Opus 4.6", "$0.85"},
			wantErr:   "invalid command-line config",
			wantNoErr: false,
		},
		{
			name:      "positional arg still renders",
			args:      []string{"extra"},
			wantOut:   []string{"Claude Opus 4.6", "$0.85"},
			wantErr:   "positional",
			wantNoErr: false,
		},
		{
			name:       "version does not render or read stdin",
			args:       []string{"--version"},
			wantOut:    []string{"test-version"},
			absentOut:  []string{"Claude Opus 4.6"},
			wantNoErr:  true,
			wantNoRead: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader(cliPayload)
			if tt.wantNoRead {
				stdin = strings.NewReader("")
			}
			code := runWithGit(tt.args, stdin, &stdout, &stderr, func(string) string { return "" }, "test-version", fakeGit)
			if code != 0 {
				t.Fatalf("runWithGit() exit = %d, want 0", code)
			}
			for _, want := range tt.wantOut {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout missing %q: %q", want, stdout.String())
				}
			}
			for _, absent := range tt.absentOut {
				if strings.Contains(stdout.String(), absent) {
					t.Fatalf("stdout should not contain %q: %q", absent, stdout.String())
				}
			}
			if tt.wantNoErr {
				if stderr.String() != "" {
					t.Fatalf("stderr = %q, want empty", stderr.String())
				}
				return
			}
			if !strings.Contains(stderr.String(), tt.wantErr) {
				t.Fatalf("stderr missing %q: %q", tt.wantErr, stderr.String())
			}
			if strings.TrimSpace(stdout.String()) == "" {
				t.Fatal("stdout must not be blank for config warnings")
			}
		})
	}
}

func fakeGit(string, time.Duration) (string, bool) {
	return "main", true
}

func requireWarning(t *testing.T, warnings []string, want string) {
	t.Helper()
	for _, warning := range warnings {
		if strings.Contains(warning, want) {
			return
		}
	}
	t.Fatalf("warnings %v do not contain %q", warnings, want)
}
