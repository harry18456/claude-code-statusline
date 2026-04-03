// Command debug-tee intercepts the Claude Code statusLine JSON payload,
// appends it to a log file, then forwards it to the real statusline binary.
//
// Env vars:
//
//	STATUSLINE_REAL  path to the real statusline binary (default: statusline.exe next to this binary)
//	STATUSLINE_LOG   log file path (default: $TEMP/cc-statusline-debug.jsonl)
package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	// Append raw JSON + newline to log file
	logPath := os.Getenv("STATUSLINE_LOG")
	if logPath == "" {
		logPath = filepath.Join(os.TempDir(), "cc-statusline-debug.jsonl")
	}
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		f.Write(data)
		f.WriteString("\n")
		f.Close()
	}

	// Forward stdin to real binary
	real := os.Getenv("STATUSLINE_REAL")
	if real == "" {
		exe, _ := os.Executable()
		real = filepath.Join(filepath.Dir(exe), "statusline.exe")
	}

	cmd := exec.Command(real)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	os.Exit(cmd.ProcessState.ExitCode())
}
