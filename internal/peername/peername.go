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

// maxBridgeIDLength bounds an accepted bridgeSessionId. Real values are
// "session_" plus 24 alphanumeric characters.
const maxBridgeIDLength = 128

// Info is the registry data resolved for one session.
type Info struct {
	// Name is the sanitized peer name: the SendMessage address for sessions
	// on the same machine. Empty when no usable entry matched.
	Name string
	// BridgeSessionID is the validated Remote Control bridge id of the
	// winning entry. Sessions on other machines reach this one at
	// "bridge:" + BridgeSessionID. Empty when Remote Control is off, the
	// value is unusable, or Name is empty.
	BridgeSessionID string
}

// registryEntry holds the registry fields used for the lookup.
// Every other field in the file is ignored.
type registryEntry struct {
	SessionID       string        `json:"sessionId"`
	Name            string        `json:"name"`
	UpdatedAt       lenientFloat  `json:"updatedAt"`
	BridgeSessionID lenientString `json:"bridgeSessionId"`
}

// lenientString decodes a JSON string. Any other JSON value, such as null
// (Remote Control off) or a number, decodes to "" instead of failing: the
// bridge id is optional, so its shape must never decide whether an entry
// matches.
type lenientString string

func (s *lenientString) UnmarshalJSON(b []byte) error {
	*s = ""
	var v string
	if err := json.Unmarshal(b, &v); err == nil {
		*s = lenientString(v)
	}
	return nil
}

// validBridgeID reports whether id is 1 to maxBridgeIDLength characters,
// each an ASCII letter, digit, '_' or '-'. The id is printed as-is and
// copied as an address, so anything else is rejected.
func validBridgeID(id string) bool {
	if id == "" || len(id) > maxBridgeIDLength {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_', c == '-':
		default:
			return false
		}
	}
	return true
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

// Lookup returns the registry data for sessionID: the sanitized peer name
// and, when Remote Control is connected, the validated bridge id. It returns
// an empty Info when no usable entry exists.
//
// Config directories are searched in order and the first directory with a
// match wins. Inside one directory, the match with the greatest updatedAt
// wins; ties keep the earliest file name. Both values come from the winning
// entry. Lookup never reads a file when sessionID is empty.
func Lookup(sessionID string, configDirs []string) Info {
	if sessionID == "" {
		return Info{}
	}
	for _, dir := range configDirs {
		if dir == "" {
			continue
		}
		if info := lookupDir(filepath.Join(dir, "sessions"), sessionID); info.Name != "" {
			return info
		}
	}
	return Info{}
}

// lookupDir scans the regular *.json files directly inside sessionsDir.
func lookupDir(sessionsDir, sessionID string) Info {
	entries, err := os.ReadDir(sessionsDir) // sorted by file name
	if err != nil {
		return Info{}
	}

	var best Info
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
		if best.Name == "" || reg.UpdatedAt > bestUpdatedAt {
			best = Info{Name: name}
			if id := string(reg.BridgeSessionID); validBridgeID(id) {
				best.BridgeSessionID = id
			}
			bestUpdatedAt = reg.UpdatedAt
		}
	}
	return best
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
