//go:build windows

package peername

import (
	"os"
	"syscall"
)

// openShared opens path read-only. Unlike os.Open, it also grants
// FILE_SHARE_DELETE, so Claude Code can still delete or rewrite its registry
// file while the statusline holds it open. Windows still refuses to rename
// another file over it during that window.
//
// If CreateFile fails, openShared retries with os.Open. That covers paths
// longer than MAX_PATH on systems without long-path support, which os.Open
// handles by adding the \\?\ prefix.
func openShared(path string) (*os.File, error) {
	if f, err := createFileShared(path); err == nil {
		return f, nil
	}
	return os.Open(path)
}

func createFileShared(path string) (*os.File, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := syscall.CreateFile(
		p,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(h), path), nil
}
