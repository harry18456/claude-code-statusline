package peername

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testSessionID is fake on purpose: it must never match a real registry
// entry on the machine that runs the tests.
const testSessionID = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"

// writeRegistry creates <configDir>/sessions/<name> with content.
func writeRegistry(t *testing.T, configDir, name, content string) {
	t.Helper()
	path := filepath.Join(configDir, "sessions", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func entryJSON(sessionID, name string) string {
	return `{"sessionId":"` + sessionID + `","name":"` + name + `"}`
}

// paddedEntry returns a matching registry entry padded to exactly size bytes.
func paddedEntry(t *testing.T, size int) string {
	t.Helper()
	prefix := `{"sessionId":"` + testSessionID + `","name":"padded-name","pad":"`
	suffix := `"}`
	n := size - len(prefix) - len(suffix)
	if n < 0 {
		t.Fatalf("size %d too small", size)
	}
	return prefix + strings.Repeat("x", n) + suffix
}

func TestLookupMatchingEntry(t *testing.T) {
	cfg := t.TempDir()
	writeRegistry(t, cfg, "14620.json", entryJSON("99999999-8888-4777-8666-555555555555", "other-project-00"))
	writeRegistry(t, cfg, "42184.json", `{"pid":42184,"sessionId":"`+testSessionID+`","name":"demo-project-07","nameSource":"derived","status":"idle","updatedAt":1795139396728}`)
	writeRegistry(t, cfg, "42184.ecac27da.key", "secret")

	if got := Lookup(testSessionID, []string{cfg}); got != "demo-project-07" {
		t.Fatalf("Lookup() = %q, want %q", got, "demo-project-07")
	}
}

func TestLookupConfigDirPrecedence(t *testing.T) {
	// The later directory holds the newer entry, so an implementation that
	// merges every directory and then picks the greatest updatedAt fails.
	envDir, homeDir := t.TempDir(), t.TempDir()
	writeRegistry(t, envDir, "1.json", `{"sessionId":"`+testSessionID+`","name":"from-env","updatedAt":100}`)
	writeRegistry(t, homeDir, "1.json", `{"sessionId":"`+testSessionID+`","name":"from-home","updatedAt":200}`)

	if got := Lookup(testSessionID, []string{envDir, homeDir}); got != "from-env" {
		t.Fatalf("Lookup() = %q, want %q", got, "from-env")
	}
}

func TestLookupLenientUpdatedAt(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name:  "string updatedAt still matches",
			files: map[string]string{"1.json": `{"sessionId":"s-1","name":"string-ts","updatedAt":"2026-11-20T02:00:00Z"}`},
			want:  "string-ts",
		},
		{
			name:  "null updatedAt still matches",
			files: map[string]string{"1.json": `{"sessionId":"s-1","name":"null-ts","updatedAt":null}`},
			want:  "null-ts",
		},
		{
			name:  "object updatedAt still matches",
			files: map[string]string{"1.json": `{"sessionId":"s-1","name":"object-ts","updatedAt":{"ms":1}}`},
			want:  "object-ts",
		},
		{
			name: "non-numeric updatedAt counts as zero",
			files: map[string]string{
				"a.json": `{"sessionId":"s-1","name":"string-ts","updatedAt":"9999"}`,
				"b.json": `{"sessionId":"s-1","name":"numeric-ts","updatedAt":5}`,
			},
			want: "numeric-ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := t.TempDir()
			for name, content := range tt.files {
				writeRegistry(t, cfg, name, content)
			}
			if got := Lookup("s-1", []string{cfg}); got != tt.want {
				t.Fatalf("Lookup() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLookupFallsBackToLaterConfigDir(t *testing.T) {
	envDir, homeDir := t.TempDir(), t.TempDir()
	writeRegistry(t, envDir, "1.json", entryJSON("other-session", "other-name"))
	writeRegistry(t, homeDir, "1.json", entryJSON(testSessionID, "from-home"))

	if got := Lookup(testSessionID, []string{envDir, homeDir}); got != "from-home" {
		t.Fatalf("Lookup() = %q, want %q", got, "from-home")
	}
}

func TestLookupMostRecentlyUpdatedWins(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name: "later file newer",
			files: map[string]string{
				"a.json": `{"sessionId":"s-1","name":"old-name","updatedAt":100}`,
				"b.json": `{"sessionId":"s-1","name":"new-name","updatedAt":200}`,
			},
			want: "new-name",
		},
		{
			name: "earlier file newer",
			files: map[string]string{
				"a.json": `{"sessionId":"s-1","name":"new-name","updatedAt":200}`,
				"b.json": `{"sessionId":"s-1","name":"old-name","updatedAt":100}`,
			},
			want: "new-name",
		},
		{
			name: "tie keeps first file name",
			files: map[string]string{
				"a.json": `{"sessionId":"s-1","name":"first","updatedAt":100}`,
				"b.json": `{"sessionId":"s-1","name":"second","updatedAt":100}`,
			},
			want: "first",
		},
		{
			name: "missing updatedAt keeps first file name",
			files: map[string]string{
				"a.json": `{"sessionId":"s-1","name":"first"}`,
				"b.json": `{"sessionId":"s-1","name":"second"}`,
			},
			want: "first",
		},
		{
			name: "present updatedAt beats missing",
			files: map[string]string{
				"a.json": `{"sessionId":"s-1","name":"no-timestamp"}`,
				"b.json": `{"sessionId":"s-1","name":"timestamped","updatedAt":5}`,
			},
			want: "timestamped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := t.TempDir()
			for name, content := range tt.files {
				writeRegistry(t, cfg, name, content)
			}
			if got := Lookup("s-1", []string{cfg}); got != tt.want {
				t.Fatalf("Lookup() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLookupSanitizesName(t *testing.T) {
	// Code points are built with fmt %c so the source holds no invisible
	// characters; the test JSON-encodes the name like Claude Code would.
	const (
		nl   = rune(0x0a)   // line feed
		cr   = rune(0x0d)   // carriage return
		tab  = rune(0x09)   // horizontal tab
		esc  = rune(0x1b)   // escape (C0)
		csi  = rune(0x9b)   // control sequence introducer (C1)
		rlo  = rune(0x202e) // right-to-left override (Cf)
		pdf  = rune(0x202c) // pop directional formatting (Cf)
		lri  = rune(0x2066) // left-to-right isolate (Cf)
		pdi  = rune(0x2069) // pop directional isolate (Cf)
		zwsp = rune(0x200b) // zero width space (Cf)
		zwj  = rune(0x200d) // zero width joiner (Cf)
		bom  = rune(0xfeff) // zero width no-break space (Cf)
		ls   = rune(0x2028) // line separator (Zl)
		ps   = rune(0x2029) // paragraph separator (Zp)
	)
	tests := []struct {
		name    string
		rawName string
		want    string
	}{
		{name: "newline and ESC removed", rawName: fmt.Sprintf("peer%cname%c[31m", nl, esc), want: "peername[31m"},
		{name: "surrounding whitespace trimmed", rawName: fmt.Sprintf("  %c spaced %c%c", tab, cr, nl), want: "spaced"},
		{name: "removal happens before trimming", rawName: fmt.Sprintf("%c spaced", esc), want: "spaced"},
		{name: "C1 control removed", rawName: fmt.Sprintf("a%cb", csi), want: "ab"},
		{name: "bidi override removed", rawName: fmt.Sprintf("peer%cname%c", rlo, pdf), want: "peername"},
		{name: "bidi isolate removed", rawName: fmt.Sprintf("%cpeer%c", lri, pdi), want: "peer"},
		{name: "zero-width characters removed", rawName: fmt.Sprintf("a%cb%cc%cd", zwsp, zwj, bom), want: "abcd"},
		{name: "line and paragraph separators removed", rawName: fmt.Sprintf("x%cy%cz", ls, ps), want: "xyz"},
		{name: "non-ASCII kept", rawName: "session id 顯示查詢", want: "session id 顯示查詢"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := json.Marshal(map[string]string{"sessionId": testSessionID, "name": tt.rawName})
			if err != nil {
				t.Fatal(err)
			}
			cfg := t.TempDir()
			writeRegistry(t, cfg, "1.json", string(content))
			if got := Lookup(testSessionID, []string{cfg}); got != tt.want {
				t.Fatalf("Lookup() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLookupControlOnlyNameIsNotAMatch(t *testing.T) {
	cfg := t.TempDir()
	writeRegistry(t, cfg, "a.json", `{"sessionId":"`+testSessionID+`","name":"\n\t\u001b","updatedAt":200}`)
	writeRegistry(t, cfg, "b.json", `{"sessionId":"`+testSessionID+`","name":"usable","updatedAt":100}`)

	if got := Lookup(testSessionID, []string{cfg}); got != "usable" {
		t.Fatalf("Lookup() = %q, want %q", got, "usable")
	}
}

func TestLookupFileSizeLimit(t *testing.T) {
	t.Run("exactly limit is read", func(t *testing.T) {
		cfg := t.TempDir()
		writeRegistry(t, cfg, "1.json", paddedEntry(t, maxRegistryFileSize))
		if got := Lookup(testSessionID, []string{cfg}); got != "padded-name" {
			t.Fatalf("Lookup() = %q, want %q", got, "padded-name")
		}
	})
	t.Run("one byte over limit is skipped", func(t *testing.T) {
		cfg := t.TempDir()
		writeRegistry(t, cfg, "1.json", paddedEntry(t, maxRegistryFileSize+1))
		if got := Lookup(testSessionID, []string{cfg}); got != "" {
			t.Fatalf("Lookup() = %q, want empty", got)
		}
	})
}

func TestLookupUnusableRegistry(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, cfg string)
	}{
		{name: "sessions directory missing", setup: func(t *testing.T, cfg string) {}},
		{name: "whitespace-only file", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", strings.Repeat(" ", 600))
		}},
		{name: "empty file", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", "")
		}},
		{name: "partially written file", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", `{"sessionId":`)
		}},
		{name: "empty name", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", entryJSON(testSessionID, ""))
		}},
		{name: "name wrong type", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", `{"sessionId":"`+testSessionID+`","name":42}`)
		}},
		{name: "session id wrong type", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", `{"sessionId":42,"name":"x"}`)
		}},
		{name: "top-level array", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", `[1,2]`)
		}},
		{name: "top-level null", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", `null`)
		}},
		{name: "key file is never read", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.abc.key", entryJSON(testSessionID, "from-key-file"))
		}},
		{name: "json suffix directory is skipped", setup: func(t *testing.T, cfg string) {
			if err := os.MkdirAll(filepath.Join(cfg, "sessions", "40404.json"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "nested sessions file is not read", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, filepath.Join("nested", "40404.json"), entryJSON(testSessionID, "nested-name"))
		}},
		{name: "session id differs by case", setup: func(t *testing.T, cfg string) {
			writeRegistry(t, cfg, "40404.json", entryJSON(strings.ToUpper(testSessionID), "upper-name"))
		}},
		{name: "sessions is a file", setup: func(t *testing.T, cfg string) {
			if err := os.WriteFile(filepath.Join(cfg, "sessions"), []byte(entryJSON(testSessionID, "x")), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := t.TempDir()
			tt.setup(t, cfg)
			if got := Lookup(testSessionID, []string{cfg}); got != "" {
				t.Fatalf("Lookup() = %q, want empty", got)
			}
		})
	}
}

func TestLookupSkipsSymlinks(t *testing.T) {
	cfg := t.TempDir()
	target := filepath.Join(t.TempDir(), "target.json")
	if err := os.WriteFile(target, []byte(entryJSON(testSessionID, "via-symlink")), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cfg, "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(cfg, "sessions", "link.json")); err != nil {
		t.Skipf("symlink not supported here: %v", err)
	}
	if got := Lookup(testSessionID, []string{cfg}); got != "" {
		t.Fatalf("Lookup() = %q, want empty", got)
	}
}

func TestLookupEmptyInputs(t *testing.T) {
	cfg := t.TempDir()
	writeRegistry(t, cfg, "1.json", entryJSON("", "ghost"))

	if got := Lookup("", []string{cfg}); got != "" {
		t.Fatalf("Lookup(empty id) = %q, want empty", got)
	}
	if got := Lookup(testSessionID, nil); got != "" {
		t.Fatalf("Lookup(nil dirs) = %q, want empty", got)
	}
	if got := Lookup(testSessionID, []string{""}); got != "" {
		t.Fatalf("Lookup(empty dir) = %q, want empty", got)
	}
}

func TestConfigDirs(t *testing.T) {
	home := filepath.Join("home", "dev")
	homeClaude := filepath.Join(home, ".claude")
	custom := filepath.Join("custom", "claude-config")

	tests := []struct {
		name   string
		getenv func(string) string
		home   string
		want   []string
	}{
		{
			name:   "unset env uses home",
			getenv: func(string) string { return "" },
			home:   home,
			want:   []string{homeClaude},
		},
		{
			name:   "env first then home",
			getenv: envWith(custom),
			home:   home,
			want:   []string{custom, homeClaude},
		},
		{
			name:   "blank env ignored",
			getenv: envWith("   "),
			home:   home,
			want:   []string{homeClaude},
		},
		{
			name:   "env padded with whitespace is trimmed",
			getenv: envWith("  " + custom + "  "),
			home:   home,
			want:   []string{custom, homeClaude},
		},
		{
			name:   "env equal to home config is de-duplicated",
			getenv: envWith(homeClaude + string(filepath.Separator)),
			home:   home,
			want:   []string{homeClaude},
		},
		{
			name:   "empty home adds no home entry",
			getenv: envWith(custom),
			home:   "",
			want:   []string{custom},
		},
		{
			name:   "nil getenv uses home",
			getenv: nil,
			home:   home,
			want:   []string{homeClaude},
		},
		{
			name:   "nothing available",
			getenv: nil,
			home:   "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConfigDirs(tt.getenv, tt.home)
			if strings.Join(got, "|") != strings.Join(tt.want, "|") {
				t.Fatalf("ConfigDirs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func envWith(configDir string) func(string) string {
	return func(key string) string {
		if key == "CLAUDE_CONFIG_DIR" {
			return configDir
		}
		return ""
	}
}

// TestOpenSharedLongPath guards the Windows fallback for paths longer than
// MAX_PATH when long-path support is disabled system-wide.
func TestOpenSharedLongPath(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 6; i++ {
		dir = filepath.Join(dir, strings.Repeat(string(rune('a'+i)), 50))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "1.json")
	if len(path) <= 260 {
		t.Fatalf("test path is only %d characters; want more than 260", len(path))
	}
	if err := os.WriteFile(path, []byte(`{"v":1}`), 0o600); err != nil {
		t.Fatal(err)
	}

	f, err := openShared(path)
	if err != nil {
		t.Fatalf("openShared(long path) error = %v", err)
	}
	f.Close()
}

// TestOpenSharedAllowsRewriteAndDeleteWhileOpen guards the Windows share
// mode: Claude Code must still be able to rewrite or delete its registry file
// while the statusline holds it open. A plain os.Open handle on Windows makes
// the delete fail with a sharing violation.
func TestOpenSharedAllowsRewriteAndDeleteWhileOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "1.json")
	if err := os.WriteFile(path, []byte(`{"v":1}`), 0o600); err != nil {
		t.Fatal(err)
	}

	f, err := openShared(path)
	if err != nil {
		t.Fatalf("openShared() error = %v", err)
	}
	defer f.Close()

	if err := os.WriteFile(path, []byte(`{"v":2}`), 0o600); err != nil {
		t.Fatalf("rewrite of an open registry file failed: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("delete of an open registry file failed: %v", err)
	}
}
