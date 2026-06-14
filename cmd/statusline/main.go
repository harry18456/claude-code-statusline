// Command statusline is the Claude Code statusLine hook binary.
// It reads a JSON payload from stdin and writes two ANSI-colored lines to stdout.
package main

import (
	"flag"
	"fmt"
	"io"
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

type cliConfig struct {
	Options     renderer.Options
	ShowVersion bool
	Warnings    []string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv, version))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string, buildVersion string) int {
	return runWithGit(args, stdin, stdout, stderr, getenv, buildVersion, gitcache.Get)
}

func runWithGit(args []string, stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string, buildVersion string, getGit func(string, time.Duration) (string, bool)) int {
	cfg := parseCLIConfig(args, getenv)
	if cfg.ShowVersion {
		fmt.Fprintln(stdout, buildVersion)
		return 0
	}

	for _, warning := range cfg.Warnings {
		fmt.Fprintln(stderr, "statusline warning:", warning)
	}

	p, err := model.ParsePayload(stdin)
	if err != nil {
		// Fallback: single gray line, exit 0 so Claude Code still works.
		fmt.Fprint(stdout, "\033[90m─ │ parse error\033[0m\n")
		return 0
	}

	// Git info via cache
	branch, dirty := getGit(p.Workspace.CurrentDir, 5*time.Second)

	git := renderer.GitInfo{
		Branch: branch,
		Dirty:  dirty,
	}

	line1, line2 := renderer.Render(p, git, cfg.Options)
	fmt.Fprint(stdout, line1, "\n", line2)
	return 0
}

func parseCLIConfig(args []string, getenv func(string) string) cliConfig {
	opts := renderer.DefaultOptions()
	opts.TrueColor = envEquals(getenv("COLORTERM"), "truecolor", "24bit")

	var hideValues []string
	var showVersion bool

	fs := flag.NewFlagSet("statusline", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&opts.ASCIIMode, "ascii", false, "enable pure ASCII rendering")
	fs.BoolVar(&opts.NerdFont, "nerdfont", false, "enable Nerd Font symbols")
	fs.BoolVar(&opts.Powerline, "powerline", false, "enable Powerline separators")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.Func("hide", "comma-separated section keys to hide", func(value string) error {
		hideValues = append(hideValues, value)
		return nil
	})

	var warnings []string
	if err := fs.Parse(args); err != nil {
		warnings = append(warnings, fmt.Sprintf("invalid command-line config: %v", err))
	}
	if positional := fs.Args(); len(positional) > 0 {
		warnings = append(warnings, "ignoring positional arguments: "+strings.Join(positional, " "))
	}

	opts.Powerline = opts.Powerline || opts.NerdFont
	if opts.ASCIIMode && (opts.NerdFont || opts.Powerline) {
		warnings = append(warnings, "--ascii conflicts with --nerdfont/--powerline; using ASCII rendering")
		opts.NerdFont = false
		opts.Powerline = false
	}

	hidden := make(map[renderer.SectionKey]bool)
	for _, value := range hideValues {
		for _, rawKey := range strings.Split(value, ",") {
			name := strings.TrimSpace(rawKey)
			if name == "" {
				continue
			}
			key, ok := renderer.SectionKeyForName(name)
			if !ok {
				warnings = append(warnings, "unknown --hide key ignored: "+name)
				continue
			}
			hidden[key] = true
		}
	}
	if len(hidden) > 0 {
		opts.HiddenSections = hidden
	}

	return cliConfig{
		Options:     opts,
		ShowVersion: showVersion,
		Warnings:    warnings,
	}
}

// envEquals returns true if value equals any of the given values, case-insensitively.
func envEquals(value string, values ...string) bool {
	for _, want := range values {
		if strings.EqualFold(value, want) {
			return true
		}
	}
	return false
}
