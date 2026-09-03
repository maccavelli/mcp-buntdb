# mcp-buntdb

`mcp-buntdb` is a lightweight Go library providing concurrency-safe initialization and cross-process advisory file locking for [BuntDB](https://github.com/tidwall/buntdb).

It ensures that when multiple Model Context Protocol (MCP) server instances or
CLI tools access the same database path simultaneously, one process acquires exclusive
file access while secondary processes automatically degrade to an in-memory database
(`:memory:`) instead of crashing or corrupting the disk log.

---

## Why It Exists

[BuntDB](https://github.com/tidwall/buntdb) is an in-memory key/value database backed by an append-only file (AOF) on disk. BuntDB does not implement native cross-process file locking.

In MCP environments, servers store state (session metadata, auth tokens, scraper caches, metrics) in shared user-level directories (e.g., `~/.cache` or `~/.local/share`). This creates concurrency hazards:

- **Multi-Instance Contention:** Multiple IDE windows, subagents, or standalone processes frequently launch against the same database file path.
- **Log Corruption Risk:** Multiple concurrent writers appending to an uncoordinated BuntDB AOF file can corrupt database integrity.
- **Resilience Over Crashing:** Failing or crashing secondary processes when a database file is busy breaks MCP tool routing. Secondary servers must start successfully and continue answering requests.

`mcp-buntdb` establishes a strict ownership boundary: **one process owns the on-disk file; all concurrent peer processes safely operate in memory.**

---

## How It Is Designed

`mcp-buntdb` is a minimal, zero-dependency helper wrapping BuntDB initialization:

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
- Acquires a non-blocking exclusive file lock:
  - **Unix (`flock_unix.go`):** `syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)`
  - **Windows (`flock_windows.go`):** `windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, ...)`

### 2. Graceful In-Memory Fallback

- If the lock is acquired, the file descriptor is retained in an internal thread-safe registry (`sync.Mutex`), and the target file is opened.
- If the lock is held by another process, `OpenBuntDB` logs a warning via `log/slog`, closes the lock descriptor, and opens `:memory:`. The caller receives an operational `*buntdb.DB`.

### 3. Explicit Lock Lifecycle Management

- Closing `*buntdb.DB` does not close the lock file descriptor held by the process.
- Call `ReleaseDBLock(dbPath)` during server shutdown or prior to in-process reopenings (e.g., periodic dashboard polling). The `.lock` file remains on disk, but the lock descriptor is released.

### 4. Error Boundaries

- Only lock collisions trigger fallback to `:memory:`.
- Hard errors (missing parent directories, permission failures, corrupt database files) fail fast and return an error directly to the caller.

---

## What It Is For

- **MCP Server Fleets:** Multi-tenant or multi-process AI tooling where multiple instances share user-level cache or state directories.
- **Embedded Go Daemons & CLIs:** Lightweight tools requiring embedded key/value storage with built-in concurrency safety.
- **High-Availability Workloads:** Applications that must stay online and functional even when disk lock contention occurs.

---

## Installation

Requires **Go 1.26.6** or later.

```bash
go get github.com/maccavelli/mcp-buntdb
```

---

## Usage

### Standard Open with Graceful Cleanup

```go
package main

import (
    "log"

    "github.com/tidwall/buntdb"
    mcpbuntdb "github.com/maccavelli/mcp-buntdb"
)

func main() {
    dbPath := "/path/to/server_state.db"

    // Optional BuntDB configuration (pass nil for defaults)
    config := &buntdb.Config{
        AutoShrinkPercentage: 50,
        AutoShrinkMinSize:    25 * 1024 * 1024,
    }

    // Opens on disk, or degrades to in-memory if locked
    db, err := mcpbuntdb.OpenBuntDB(dbPath, config)
    if err != nil {
        log.Fatalf("failed to open database: %v", err)
    }
    defer func() {
        _ = db.Close()
        mcpbuntdb.ReleaseDBLock(dbPath)
    }()

    // Execute standard BuntDB operations
    err = db.Update(func(tx *buntdb.Tx) error {
        _, _, err := tx.Set("session:active", "true", nil)
        return err
    })
    if err != nil {
        log.Printf("write failed: %v", err)
    }
}
```

### Explicit In-Memory Mode

Pass `":memory:"` to bypass file locking entirely:

```go
db, err := mcpbuntdb.OpenBuntDB(":memory:", nil)
if err != nil {
    log.Fatalf("failed to open in-memory db: %v", err)
}
defer db.Close()
```

---

## API Reference

| Function | Signature | Description |
| :--- | :--- | :--- |
| `OpenBuntDB` | `func OpenBuntDB(dbPath string, config *buntdb.Config) (*buntdb.DB, error)` | Acquires advisory lock (if disk path), opens BuntDB, applies optional config. Degrades to `:memory:` on lock conflict. |
| `ReleaseDBLock` | `func ReleaseDBLock(dbPath string)` | Closes the lock file descriptor for `dbPath` and removes it from the process registry. Safe no-op for `""` and `":memory:"`. |

---

## Behavior Matrix

| Scenario | Lock Action | Storage Target | Outcome |
| :--- | :--- | :--- | :--- |
| **First Process on Path** | Acquires `.lock` exclusively | Disk AOF file (`dbPath`) | Returns `*buntdb.DB`, persists to disk |
| **Concurrent Process on Path** | Lock contention detected | Memory (`:memory:`) | Logs warning (`slog.Warn`), returns `*buntdb.DB` |
| **Explicit `":memory:"`** | Bypasses locking | Memory (`:memory:`) | Returns `*buntdb.DB` in-memory |
| **Invalid Directory / Permissions** | Lock creation fails | None | Returns filesystem error |
| **Corrupted AOF File** | Lock acquired | None | Returns `buntdb.Open` error |

---

## Development

```bash
# Run tests
make test

# Static analysis
make vet

# Run linter
make lint

# Format source code
make fmt

# Tidy Go modules
make tidy
```
