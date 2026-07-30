//go:build windows

package mcpbuntdb

import (
	"os"

	"golang.org/x/sys/windows"
)

func flock(f *os.File) error {
	h := windows.Handle(f.Fd())
	var overlapped windows.Overlapped
	return windows.LockFileEx(h, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
}
