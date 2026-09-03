> **Migration notice.** This repository was previously a one-way published
> export of a privately hosted project; early history contains squashed
> “Sync from internal source …” commits from that period. `origin` is now
> [`github.com/maccavelli/mcp-buntdb`](https://github.com/maccavelli/mcp-buntdb),
> `main` tracks it directly, and normal commits/PRs land here.

# mcp-buntdb

`mcp-buntdb` is a lightweight Go helper library providing concurrency-safe initialization
and cross-process advisory file locking for [BuntDB](https://github.com/tidwall/buntdb)
in Model Context Protocol (MCP) server fleets and CLI tools.

It wraps BuntDB initialization with non-blocking file locking. When lock contention occurs—such
as when multiple IDE instances or background processes access the same database file
simultaneously—it automatically degrades to an in-memory database (`:memory:`) rather than
crashing the server or risking database log corruption.

---

## Why It Exists

[BuntDB](https://github.com/tidwall/buntdb) is an embeddable, in-memory key/value database
backed by an append-only file (AOF) on disk. By design, BuntDB does not implement native
cross-process file locking.

In MCP environments, servers persist lightweight state (such as session metadata,
authentication tokens, scraping caches, and metrics) in shared user-level directories
(e.g., `~/.cache` or `~/.local/share`). This introduces concurrency challenges:

- **Multi-Instance Contention:** Multiple IDE windows, orchestrated subagents, standalone
  CLIs, or background daemons frequently boot against identical data paths.
- **Log Corruption Risk:** Multiple concurrent writers pointing to an uncoordinated BuntDB
  AOF file can corrupt the append-only log.
- **Resilience Over Hard Failures:** Crashing secondary instances when a lock is busy breaks
  MCP tool routing. Secondary instances must remain operational and continue answering requests.

`mcp-buntdb` resolves this with a clear ownership boundary: **one process holds exclusive write
ownership over the disk file, while concurrent peer processes seamlessly degrade to in-memory mode.**

---

## Architecture & Design

`mcp-buntdb` is a focused library consisting of two exported functions without background
daemons or heavy store wrappers.

```text
                    ┌────────────────────────┐
                    │ OpenBuntDB(path, cfg)  │
                    └───────────┬────────────┘
                                │
                      Is path == ":memory:"?
                                │
                    ┌───────────┴───────────┐
                   Yes                      No
                    │                       │
                    │         Attempt Non-Blocking flock on
                    │              "{path}.lock"
                    │                       │
                    │             ┌─────────┴─────────┐
                    │          Success             Conflict / Held
                    │             │                   │
                    │             │             Log Warning (slog) &
                    │             │             Degrade path to ":memory:"
                    │             │                   │
                    └─────────► ┌─┴───────────────────┴─┐
                                │   buntdb.Open(target) │
                                └───────────┬───────────┘
                                            │
                                Apply optional *buntdb.Config
                                            │
                                    Return *buntdb.DB
```

### 1. Cross-Platform Advisory Locking

- For disk paths, `OpenBuntDB` opens a companion lock file (`<dbPath>.lock`) with `0600` permissions.
- It requests a non-blocking exclusive lock:
  - **Unix (`flock_unix.go`):** `syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)`
  - **Windows (`flock_windows.go`):** `windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, ...)`

### 2. Automatic Memory Fallback

- If the lock succeeds, the file descriptor is recorded in a thread-safe registry (`sync.Mutex`),
  and the target file is opened.
- If the lock fails due to contention, `OpenBuntDB` logs a warning via `log/slog`, closes the
  unacquired lock descriptor, and opens `:memory:`. The caller receives an operational `*buntdb.DB`.

### 3. Error Boundaries & Ownership

- **Degraded on Contention Only:** Only lock collisions trigger `:memory:` fallback. System-level
  filesystem errors (permissions, directory missing) or invalid file formats fail fast and return an error.
- **Caller Responsibilities:** Callers maintain ownership of data modeling, indexes, spatial/JSON
  queries, transaction lifecycles, and directory creation.

### 4. Lock Lifecycle (`ReleaseDBLock`)

- Closing the `*buntdb.DB` instance does not close the lock file descriptor.
- Call `ReleaseDBLock(dbPath)` during server shutdown or before in-process reopenings (e.g., periodic
  dashboard polling). The `.lock` file remains on disk as metadata, but the lock descriptor is released.

---

## Installation & Requirements

Requires **Go 1.26.6** or newer.

```bash
go get github.com/maccavelli/mcp-buntdb
```

---

## Usage

### Basic Initialization & Graceful Cleanup

```go
package main

import (
    "log"

    "github.com/tidwall/buntdb"
    mcpbuntdb "github.com/maccavelli/mcp-buntdb"
)

func main() {
    dbPath := "/path/to/server_state.db"

    // Configure BuntDB options (optional; pass nil to use defaults)
    config := &buntdb.Config{
        AutoShrinkPercentage: 50,
        AutoShrinkMinSize:    25 * 1024 * 1024,
    }

    // Opens the database on disk, or degrades to :memory: if locked
    db, err := mcpbuntdb.OpenBuntDB(dbPath, config)
    if err != nil {
        log.Fatalf("failed to open database: %v", err)
    }
    defer func() {
        _ = db.Close()
        mcpbuntdb.ReleaseDBLock(dbPath)
    }()

    // Use standard BuntDB transactions
    err = db.Update(func(tx *buntdb.Tx) error {
        _, _, err := tx.Set("session:123", `{"status":"active"}`, nil)
        return err
    })
    if err != nil {
        log.Printf("transaction failed: %v", err)
    }
}
```

### Explicit In-Memory Mode

Pass `":memory:"` directly to bypass file locking entirely:

```go
db, err := mcpbuntdb.OpenBuntDB(":memory:", nil)
if err != nil {
    log.Fatalf("failed to open in-memory db: %v", err)
}
defer db.Close()
```

---

## API Reference

### `OpenBuntDB`

```go
func OpenBuntDB(dbPath string, config *buntdb.Config) (*buntdb.DB, error)
```

- **`dbPath`**: Path to the BuntDB file on disk, or `":memory:"`.
- **`config`**: Optional pointer to `buntdb.Config`. When provided, applied via `db.SetConfig(*config)`.
  If `SetConfig` fails, a warning is logged and the database instance is still returned.
- **Returns**: Active `*buntdb.DB` instance or an initialization error.

### `ReleaseDBLock`

```go
func ReleaseDBLock(dbPath string)
```

- **`dbPath`**: Path to the database whose lock file descriptor should be closed and removed from the active process registry.
- **Behavior**: Thread-safe. Safe no-op if `dbPath` is empty, `":memory:"`, or not currently tracked.

---

## Behavior Matrix

| Scenario | Lock File Action | Resulting Storage | Error / Log Behavior |
| :--- | :--- | :--- | :--- |
| **First Process on Path** | Acquires `.lock` exclusively | Disk AOF file (`dbPath`) | Returns `*buntdb.DB`, no errors |
| **Concurrent Process on Path** | Lock collision detected | Memory (`:memory:`) | Logs `slog.Warn`, returns working `*buntdb.DB` |
| **Explicit `":memory:"`** | Bypasses lock check | Memory (`:memory:`) | Returns `*buntdb.DB`, no errors |
| **Invalid Directory / Permission** | Lock creation fails | None | Returns `error` immediately |
| **Corrupted Database File** | Lock acquired | None | Returns `buntdb.Open` `error` |

---

## Development

The project uses a standard `Makefile` for developer workflows:

```bash
# Run test suite
make test

# Run vet static analysis
make vet

# Run golangci-lint
make lint

# Format source files
make fmt

# Tidy module dependencies
make tidy
```

---

## License

Refer to the repository license for terms of distribution and use.
