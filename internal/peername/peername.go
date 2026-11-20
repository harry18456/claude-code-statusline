// Package peername resolves the Claude Code session peer name: the name that
// other Claude Code sessions use as the SendMessage address for a session.
//
// Claude Code records the name in its session registry, one JSON file per
// running process at <config>/sessions/<pid>.json. The registry is an
// undocumented internal format, so every failure resolves to "no name" and
// never surfaces as an error.
package peername

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// maxRegistryFileSize bounds how much of one registry file is read.
// Real entries are under 1 KiB; larger files are skipped.
const maxRegistryFileSize = 64 * 1024

// registryEntry holds the registry fields used for the lookup.
// Every other field in the file is ignored.
type registryEntry struct {
	SessionID string       `json:"sessionId"`
	Name      string       `json:"name"`
	UpdatedAt lenientFloat `json:"updatedAt"`
}

// lenientFloat decodes a JSON number. Any other JSON value, such as a string
// or null, decodes to 0 instead of failing: updatedAt only breaks ties, so
// its shape must never decide whether an entry matches.
type lenientFloat float64

func (f *lenientFloat) UnmarshalJSON(b []byte) error {
	*f = 0
	var v float64
	if err := json.Unmarshal(b, &v); err == nil {
		*f = lenientFloat(v)
	}
	return nil
}

// ConfigDirs returns the candidate Claude Code config directories in lookup
// order: $CLAUDE_CONFIG_DIR when it is not blank, then <home>/.claude.
// A directory that repeats after filepath.Clean is listed once.
// A nil getenv or an empty home contributes no entry.
func ConfigDirs(getenv func(string) string, home string) []string {
	var dirs []string
	seen := make(map[string]bool)
	add := func(dir string) {
		if dir == "" {
			return
		}
		dir = filepath.Clean(dir)
		if seen[dir] {
			return
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}

	if getenv != nil {
		add(strings.TrimSpace(getenv("CLAUDE_CONFIG_DIR")))
	}
	if home != "" {
		add(filepath.Join(home, ".claude"))
	}
	return dirs
}

// Lookup returns the sanitized peer name that the session registry records
// for sessionID, or "" when no usable entry exists.
//
// Config directories are searched in order and the first directory with a
// match wins. Inside one directory, the match with the greatest updatedAt
// wins; ties keep the earliest file name. Lookup never reads a file when
// sessionID is empty.
func Lookup(sessionID string, configDirs []string) string {
	if sessionID == "" {
		return ""
	}
	for _, dir := range configDirs {
		if dir == "" {
			continue
		}
		if name := lookupDir(filepath.Join(dir, "sessions"), sessionID); name != "" {
			return name
		}
	}
	return ""
}

// lookupDir scans the regular *.json files directly inside sessionsDir.
func lookupDir(sessionsDir, sessionID string) string {
	entries, err := os.ReadDir(sessionsDir) // sorted by file name
	if err != nil {
		return ""
	}

	bestName := ""
	var bestUpdatedAt lenientFloat
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		reg, ok := readEntry(filepath.Join(sessionsDir, entry.Name()))
		if !ok || reg.SessionID != sessionID {
			continue
		}
		name := sanitizeName(reg.Name)
		if name == "" {
			continue
		}
		if bestName == "" || reg.UpdatedAt > bestUpdatedAt {
			bestName = name
			bestUpdatedAt = reg.UpdatedAt
		}
	}
	return bestName
}

// readEntry decodes one registry file. It reports false for any open or read
// error, for anything that is not a regular file once opened, for files
// larger than maxRegistryFileSize, and for any JSON error, including a
// sessionId or name with the wrong type.
func readEntry(path string) (registryEntry, bool) {
	f, err := openShared(path)
	if err != nil {
		return registryEntry{}, false
	}
	defer f.Close()

	// Check the open handle, not the directory entry: the entry can be
	// swapped for a pipe or device after ReadDir, and reading one can block.
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() {
		return registryEntry{}, false
	}

	data, err := io.ReadAll(io.LimitReader(f, maxRegistryFileSize+1))
	if err != nil || len(data) > maxRegistryFileSize {
		return registryEntry{}, false
	}

	var reg registryEntry
	if err := json.Unmarshal(data, &reg); err != nil {
		return registryEntry{}, false
	}
	return reg, true
}

// sanitizeName removes every character that can break the two-line output,
// inject terminal escape sequences, or change the line's appearance while
// staying invisible: controls (Cc), format characters such as bidirectional
// overrides and zero-width characters (Cf), and line or paragraph
// separators (Zl, Zp). It then trims surrounding whitespace.
func sanitizeName(name string) string {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.In(r, unicode.Cc, unicode.Cf, unicode.Zl, unicode.Zp) {
			return -1
		}
		return r
	}, name)
	return strings.TrimSpace(cleaned)
}
