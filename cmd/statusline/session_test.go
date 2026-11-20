package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"claude-code-statusline/internal/peername"
	"claude-code-statusline/internal/renderer"
)

// sessionTestID is fake on purpose: it must never match a real registry
// entry on the machine that runs the tests.
const sessionTestID = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"

// sessionPayload carries a session_name that differs from the registry name,
// so tests can prove the registry name is the one rendered.
const sessionPayload = `{"session_id":"` + sessionTestID + `","session_name":"Refactor config loader","model":{"display_name":"Claude Opus 4.6"},"context_window":{"used_percentage":42,"context_window_size":200000},"cost":{"total_cost_usd":0.85,"total_duration_ms":222000},"workspace":{"current_dir":"/tmp/project"}}`

func TestResolveSession(t *testing.T) {
	type call struct {
		sessionID string
		dirs      []string
	}
	home := filepath.Join("home", "dev")
	custom := filepath.Join("custom", "config")
	getenv := func(key string) string {
		if key == "CLAUDE_CONFIG_DIR" {
			return custom
		}
		return ""
	}
	homeOK := func() (string, error) { return home, nil }
	homeErr := func() (string, error) { return "", errors.New("no home") }

	tests := []struct {
		name      string
		opts      renderer.Options
		sessionID string
		homeDir   func() (string, error)
		wantCalls []call
		wantName  string
	}{
		{
			name:      "visible session looks up env dir then home dir",
			sessionID: sessionTestID,
			homeDir:   homeOK,
			wantCalls: []call{{sessionID: sessionTestID, dirs: []string{custom, filepath.Join(home, ".claude")}}},
			wantName:  "stub-name",
		},
		{
			name:      "hidden bridge still looks up the name",
			opts:      renderer.Options{HiddenSections: map[renderer.SectionKey]bool{renderer.SectionBridge: true}},
			sessionID: sessionTestID,
			homeDir:   homeOK,
			wantCalls: []call{{sessionID: sessionTestID, dirs: []string{custom, filepath.Join(home, ".claude")}}},
			wantName:  "stub-name",
		},
		{
			name:      "hidden session skips lookup",
			opts:      renderer.Options{HiddenSections: map[renderer.SectionKey]bool{renderer.SectionSession: true}},
			sessionID: sessionTestID,
			homeDir:   homeOK,
		},
		{
			name:      "empty session id skips lookup",
			sessionID: "",
			homeDir:   homeOK,
		},
		{
			name:      "home error still searches env dir",
			sessionID: sessionTestID,
			homeDir:   homeErr,
			wantCalls: []call{{sessionID: sessionTestID, dirs: []string{custom}}},
			wantName:  "stub-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []call
			lookup := func(sessionID string, dirs []string) peername.Info {
				calls = append(calls, call{sessionID: sessionID, dirs: dirs})
				return peername.Info{Name: "stub-name", BridgeSessionID: "stub_bridge"}
			}
			got := resolveSession(tt.opts, tt.sessionID, getenv, tt.homeDir, lookup)
			if got.PeerName != tt.wantName {
				t.Fatalf("PeerName = %q, want %q", got.PeerName, tt.wantName)
			}
			wantBridge := ""
			if tt.wantName != "" {
				wantBridge = "stub_bridge"
			}
			if got.BridgeSessionID != wantBridge {
				t.Fatalf("BridgeSessionID = %q, want %q", got.BridgeSessionID, wantBridge)
			}
			if len(calls) != len(tt.wantCalls) {
				t.Fatalf("lookup calls = %+v, want %+v", calls, tt.wantCalls)
			}
			for i := range calls {
				if calls[i].sessionID != tt.wantCalls[i].sessionID ||
					strings.Join(calls[i].dirs, "|") != strings.Join(tt.wantCalls[i].dirs, "|") {
					t.Fatalf("lookup call %d = %+v, want %+v", i, calls[i], tt.wantCalls[i])
				}
			}
		})
	}
}

func TestMainSessionPeerName(t *testing.T) {
	validRegistry := map[string]string{
		"42184.json":         `{"pid":42184,"sessionId":"` + sessionTestID + `","name":"demo-project-07","nameSource":"derived","updatedAt":1795139396728}`,
		"42184.ecac27da.key": "secret",
		"14620.json":         `{"sessionId":"99999999-8888-4777-8666-555555555555","name":"other-project-00"}`,
	}
	brokenRegistry := map[string]string{
		"39840.json":    strings.Repeat(" ", 600),
		"40404.json":    `{"sessionId":`,
		"41688.abc.key": `{"sessionId":"` + sessionTestID + `","name":"from-key-file"}`,
	}

	tests := []struct {
		name      string
		args      []string
		registry  map[string]string
		wantLine2 string // suffix of line 2 after ANSI is stripped
		absent    []string
	}{
		{
			name:      "default mode shows registry name",
			registry:  validRegistry,
			wantLine2: "⇄ demo-project-07",
			absent:    []string{"Refactor config loader"},
		},
		{
			name:      "ascii mode shows peer label",
			args:      []string{"--ascii"},
			registry:  validRegistry,
			wantLine2: "| peer:demo-project-07",
			absent:    []string{"⇄", "Refactor config loader"},
		},
		{
			name:     "hide session omits the name",
			args:     []string{"--hide", "session"},
			registry: validRegistry,
			absent:   []string{"demo-project-07", "⇄", "peer:"},
		},
		{
			name:     "broken registry omits the name",
			registry: brokenRegistry,
			absent:   []string{"from-key-file", "⇄", "peer:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emptyHome := t.TempDir()
			t.Setenv("HOME", emptyHome)
			t.Setenv("USERPROFILE", emptyHome)

			configDir := t.TempDir()
			sessionsDir := filepath.Join(configDir, "sessions")
			if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
				t.Fatal(err)
			}
			for name, content := range tt.registry {
				if err := os.WriteFile(filepath.Join(sessionsDir, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			getenv := func(key string) string {
				if key == "CLAUDE_CONFIG_DIR" {
					return configDir
				}
				return ""
			}

			var stdout, stderr bytes.Buffer
			var code int
			processStderr := captureProcessStderr(t, func() {
				code = runWithGit(tt.args, strings.NewReader(sessionPayload), &stdout, &stderr, getenv, "test-version", fakeGit)
			})
			if code != 0 {
				t.Fatalf("runWithGit() exit = %d, want 0", code)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr should be empty, got %q", stderr.String())
			}
			if processStderr != "" {
				t.Fatalf("process stderr and log output should be empty, got %q", processStderr)
			}

			lines := strings.Split(stdout.String(), "\n")
			if len(lines) != 2 {
				t.Fatalf("stdout should have 2 lines, got %d: %q", len(lines), stdout.String())
			}
			line2 := stripTestANSI(lines[1])
			if tt.wantLine2 != "" && !strings.HasSuffix(line2, tt.wantLine2) {
				t.Fatalf("line2 should end with %q, got %q", tt.wantLine2, line2)
			}
			for _, absent := range tt.absent {
				if strings.Contains(stdout.String(), absent) {
					t.Fatalf("stdout should not contain %q: %q", absent, stdout.String())
				}
			}
			if strings.HasSuffix(line2, " │ ") || strings.HasSuffix(line2, " | ") {
				t.Fatalf("line2 should not end with a separator, got %q", line2)
			}
		})
	}
}

// captureProcessStderr runs fn while os.Stderr and the standard logger write
// to a pipe, and returns what they wrote. It catches output that bypasses the
// stderr writer injected into runWithGit.
func captureProcessStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStderr := os.Stderr
	origLog := log.Writer()
	os.Stderr = w
	log.SetOutput(w)
	// Restore and close even when fn ends the test early with t.Fatal;
	// closing w also ends the reader goroutine.
	t.Cleanup(func() {
		os.Stderr = origStderr
		log.SetOutput(origLog)
		w.Close()
	})

	// Drain the pipe while fn runs, so a large write cannot block fn.
	captured := make(chan string, 1)
	go func() {
		out, _ := io.ReadAll(r)
		r.Close()
		captured <- string(out)
	}()

	fn()
	os.Stderr = origStderr
	log.SetOutput(origLog)
	w.Close()
	return <-captured
}

// stripTestANSI removes SGR escape sequences such as "\033[90m".
func stripTestANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func TestMainBridgeAddress(t *testing.T) {
	const bridgeID = "session_01EXAMPLEb2kQ7mZp4Lx9VdR"
	entry := func(bridgeJSON string) map[string]string {
		return map[string]string{
			"42184.json": `{"pid":42184,"sessionId":"` + sessionTestID + `","name":"demo-project-07","nameSource":"derived","updatedAt":1795139396728,"bridgeSessionId":` + bridgeJSON + `}`,
		}
	}

	tests := []struct {
		name      string
		args      []string
		registry  map[string]string
		wantLine2 string // suffix of line 2 after ANSI is stripped
		absent    []string
	}{
		{
			name:      "connected shows name and bridge",
			registry:  entry(`"` + bridgeID + `"`),
			wantLine2: "⇄ demo-project-07 · bridge:" + bridgeID,
		},
		{
			name:      "ascii shows peer and bridge labels",
			args:      []string{"--ascii"},
			registry:  entry(`"` + bridgeID + `"`),
			wantLine2: "| peer:demo-project-07 bridge:" + bridgeID,
			absent:    []string{"⇄", " · "},
		},
		{
			name:      "hide bridge keeps the name",
			args:      []string{"--hide", "bridge"},
			registry:  entry(`"` + bridgeID + `"`),
			wantLine2: "⇄ demo-project-07",
			absent:    []string{"bridge:"},
		},
		{
			name:      "disconnected shows only the name",
			registry:  entry(`null`),
			wantLine2: "⇄ demo-project-07",
			absent:    []string{"bridge:"},
		},
		{
			name:      "invalid bridge value shows only the name",
			registry:  entry(`"session 01; rm -rf"`),
			wantLine2: "⇄ demo-project-07",
			absent:    []string{"bridge:", "rm -rf"},
		},
		{
			name:     "hide session hides both",
			args:     []string{"--hide", "session"},
			registry: entry(`"` + bridgeID + `"`),
			absent:   []string{"demo-project-07", "bridge:", "⇄"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emptyHome := t.TempDir()
			t.Setenv("HOME", emptyHome)
			t.Setenv("USERPROFILE", emptyHome)

			configDir := t.TempDir()
			sessionsDir := filepath.Join(configDir, "sessions")
			if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
				t.Fatal(err)
			}
			for name, content := range tt.registry {
				if err := os.WriteFile(filepath.Join(sessionsDir, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			getenv := func(key string) string {
				if key == "CLAUDE_CONFIG_DIR" {
					return configDir
				}
				return ""
			}

			var stdout, stderr bytes.Buffer
			var code int
			processStderr := captureProcessStderr(t, func() {
				code = runWithGit(tt.args, strings.NewReader(sessionPayload), &stdout, &stderr, getenv, "test-version", fakeGit)
			})
			if code != 0 {
				t.Fatalf("runWithGit() exit = %d, want 0", code)
			}
			if stderr.Len() != 0 || processStderr != "" {
				t.Fatalf("stderr should be empty, got %q / %q", stderr.String(), processStderr)
			}
			lines := strings.Split(stdout.String(), "\n")
			if len(lines) != 2 {
				t.Fatalf("stdout should have 2 lines, got %d: %q", len(lines), stdout.String())
			}
			line2 := stripTestANSI(lines[1])
			if tt.wantLine2 != "" && !strings.HasSuffix(line2, tt.wantLine2) {
				t.Fatalf("line2 should end with %q, got %q", tt.wantLine2, line2)
			}
			if strings.HasSuffix(line2, " │ ") || strings.HasSuffix(line2, " | ") {
				t.Fatalf("line2 should not end with a separator, got %q", line2)
			}
			for _, absent := range tt.absent {
				if strings.Contains(line2, absent) {
					t.Fatalf("line2 should not contain %q, got %q", absent, line2)
				}
			}
		})
	}
}
