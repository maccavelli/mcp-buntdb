package mcpbuntdb

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/tidwall/buntdb"
)

const memoryDBPath = ":memory:"

var (
	lockFiles   = make(map[string]*os.File)
	lockFilesMu sync.Mutex
)

func closeLockFile(f *os.File) {
	if f == nil {
		return
	}
	if err := f.Close(); err != nil {
		slog.Debug("store: close lock file", "error", err)
	}
}

// OpenBuntDB initializes a BuntDB store at the specified dbPath.
// If the database cannot be opened due to lock contention (e.g., from multiple IDE instances),
// it automatically degrades to an in-memory BuntDB instance.
// If config is non-nil, it will be applied to the database.
func OpenBuntDB(dbPath string, config *buntdb.Config) (*buntdb.DB, error) {
	if dbPath != memoryDBPath {
		lockPath := filepath.Clean(dbPath + ".lock")
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to open lock file: %w", err)
		}

		err = flock(lockFile)
		if err != nil {
			slog.Warn("store: BuntDB lock conflict detected (likely multi-IDE instantiation). Degrading to in-memory mode.", "original_path", dbPath, "error", err)
			closeLockFile(lockFile)
			dbPath = memoryDBPath
		} else {
			lockFilesMu.Lock()
			if old, exists := lockFiles[dbPath]; exists {
				closeLockFile(old)
			}
			lockFiles[dbPath] = lockFile
			lockFilesMu.Unlock()
		}
	}

	database, err := buntdb.Open(dbPath)
	if err != nil {
		if dbPath != memoryDBPath {
			lockFilesMu.Lock()
			if lf, exists := lockFiles[dbPath]; exists {
				closeLockFile(lf)
				delete(lockFiles, dbPath)
			}
			lockFilesMu.Unlock()
		}
		return nil, err
	}

	if config != nil {
		if err := database.SetConfig(*config); err != nil {
			slog.Warn("store: failed to apply BuntDB config", "error", err)
		}
	}

	return database, nil
}

// ReleaseDBLock closes and removes the advisory lock file associated with dbPath.
// Call this when shutting down a store opened via OpenBuntDB so the path can be
// reopened (e.g. dashboard snapshot polling in the same process).
func ReleaseDBLock(dbPath string) {
	if dbPath == "" || dbPath == memoryDBPath {
		return
	}
	lockFilesMu.Lock()
	defer lockFilesMu.Unlock()
	if lf, exists := lockFiles[dbPath]; exists {
		closeLockFile(lf)
		delete(lockFiles, dbPath)
	}
}
