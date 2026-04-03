// Command statusline is the Claude Code statusLine hook binary.
// It reads a JSON payload from stdin and writes two ANSI-colored lines to stdout.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"claude-code-statusline/internal/gitcache"
	"claude-code-statusline/internal/model"
	"claude-code-statusline/internal/renderer"
)

// version is set at build time via -ldflags "-X main.version=v1.2.3".
// Local dev builds report "dev".
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	p, err := model.ParsePayload(os.Stdin)
	if err != nil {
		// Fallback: single gray line, exit 0 so Claude Code still works
		fmt.Printf("\033[90m─ │ parse error\033[0m\n")
		os.Exit(0)
	}

	// Resolve options from environment
	opts := renderer.Options{
		ASCIIMode: os.Getenv("CLAUDE_STATUSLINE_ASCII") == "1",
		NerdFont:  os.Getenv("CLAUDE_STATUSLINE_NERDFONT") == "1",
		TrueColor: isenv("COLORTERM", "truecolor", "24bit"),
	}
	// Powerline follows NerdFont unless explicitly overridden
	opts.Powerline = opts.NerdFont || os.Getenv("CLAUDE_STATUSLINE_POWERLINE") == "1"

	// Git info via cache
	cacheFile := gitcache.DefaultCacheFile()
	branch, dirty := gitcache.Get(p.Workspace.CurrentDir, cacheFile, 5*time.Second)

	git := renderer.GitInfo{
		Branch: branch,
		Dirty:  dirty,
	}

	line1, line2 := renderer.Render(p, git, opts)
	fmt.Printf("%s\n%s", line1, line2)
}

// isenv returns true if the environment variable key equals any of the given values.
func isenv(key string, values ...string) bool {
	v := os.Getenv(key)
	for _, want := range values {
		if strings.EqualFold(v, want) {
			return true
		}
	}
	return false
}
