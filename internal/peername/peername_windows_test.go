//go:build windows

package peername

import (
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// TestOpenSharedFallsBackToOSOpen proves the os.Open fallback runs when
// CreateFile fails. CreateFile without FILE_FLAG_BACKUP_SEMANTICS cannot open
// a directory while os.Open can, so a directory exercises the fallback on
// every machine, whatever its long-path setting.
func TestOpenSharedFallsBackToOSOpen(t *testing.T) {
	dir := t.TempDir()
	if f, err := createFileShared(dir); err == nil {
		f.Close()
		t.Fatal("createFileShared(directory) succeeded; the test no longer exercises the fallback")
	}
	f, err := openShared(dir)
	if err != nil {
		t.Fatalf("openShared(directory) error = %v; want the os.Open fallback to succeed", err)
	}
	f.Close()
}

// TestReadEntrySkipsNamedPipe proves readEntry checks the open handle before
// reading. A registry entry swapped for a named pipe between ReadDir and
// open would otherwise block the statusline until the pipe server writes.
func TestReadEntrySkipsNamedPipe(t *testing.T) {
	const (
		pipeAccessOutbound = 0x00000002
		pipeTypeByteWait   = 0x00000000
	)
	name := `\\.\pipe\peername-test-` + strconv.Itoa(os.Getpid())
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		t.Fatal(err)
	}
	createNamedPipe := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateNamedPipeW")
	h, _, callErr := createNamedPipe.Call(
		uintptr(unsafe.Pointer(namePtr)),
		pipeAccessOutbound,
		pipeTypeByteWait,
		1,    // max instances
		4096, // out buffer
		4096, // in buffer
		0,    // default timeout
		0,    // security attributes
	)
	if syscall.Handle(h) == syscall.InvalidHandle {
		t.Skipf("CreateNamedPipeW failed: %v", callErr)
	}
	defer syscall.CloseHandle(syscall.Handle(h))

	done := make(chan bool, 1)
	go func() {
		_, ok := readEntry(name)
		done <- ok
	}()

	select {
	case ok := <-done:
		if ok {
			t.Fatal("readEntry() on a named pipe reported success")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("readEntry() blocked on a named pipe")
	}
}
