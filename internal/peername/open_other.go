//go:build !windows

package peername

import "os"

// openShared opens path read-only. On non-Windows systems an open file never
// blocks another process from renaming over or deleting it.
func openShared(path string) (*os.File, error) {
	return os.Open(path)
}
