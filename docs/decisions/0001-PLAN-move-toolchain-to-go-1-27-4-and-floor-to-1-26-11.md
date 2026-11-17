---
status: in-progress
date: 2026-11-17
associated-madr: "0001-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md"
---
# Implement Go 1.27.4 as the toolchain and `go 1.26.11` as the floor

Associated MADR: [0001-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md](0001-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md)
(option A, as the owner ruled).

## Goal

`go.mod` says `go 1.26.11` and `toolchain go1.27.4`, and every current
statement in the tree agrees. The plan stops before any tag or release.

## Scope

### In scope

| File | Line (at `23a4cd8`) | Change |
| :--- | :--- | :--- |
| `go.mod` | 3 | `go 1.26.6` becomes `go 1.26.11` |
| `go.mod` | after 3 | add a blank line and `toolchain go1.27.4` (lines 4 and 5) |
| `README.md` | 92 | "Requires **Go 1.26.6** or later." becomes "Requires **Go 1.26.11** or later; built and tested with Go 1.27.4." |

`.github/workflows/ci.yml` is **not** changed: `go-version-file: go.mod`
(`:15`) follows the `toolchain` line (MADR E2).

### Out of scope

* A floor job in CI: a named gap in the MADR's consequences, for the
  owner to decide fleet-wide.
* The other repositories, including every consumer (MADR E5).
* Any tag or release; they wait for the owner's word.
* An index: the repository has no `docs/README.md` or other records
  index, so none is created.

## Implementation Steps

1. Edit by a Python script under the scratch folder, run by path, with an
   `assert` on each old string's count before replacing; read each result
   back.
2. `go mod tidy -diff` shows nothing; `go.sum` is unchanged; `git diff
   --stat` names only `go.mod` and `README.md`.
3. The version-agreement check (below) passes on the tree and fails on
   three planted copies.
4. Gates: `make fmt`-equivalent `gofmt -l`, `make vet`, `make lint`,
   `make test` (there are no test files: "no test files" is the result),
   `govulncheck ./...` under `go1.27.4` and `go1.26.11` with
   `GOPROXY=off`, the floor build and vet under `GOTOOLCHAIN=go1.26.11`,
   and the `GOOS=linux` and `GOOS=windows` builds (`flock_unix.go`,
   `flock_windows.go`) with `CGO_ENABLED=0`.
5. Stage the record files only; leave `go.mod` and `README.md` unstaged.
   No commit, push or tag.

The version-agreement check, `check_versions.py ROOT` in the scratch
folder (not added to the tree), reads the floor `F` and toolchain `T`
from `ROOT/go.mod`, and asserts: `go.mod` has exactly one `go` and one
`toolchain` line; `README.md` says `Go F or later` and `Go T`; every
`1.2x.y` or `go1.2x.y` string outside `docs/decisions/` and `go.sum`
equals `F` or `T`; `ci.yml` still reads `go-version-file: go.mod` and has
no literal version.

## Verification

* **V1:** the agreement check exits 0 on the tree and non-zero on a
  wrong toolchain in `go.mod`, a wrong floor in `go.mod`, and a stray old
  pin in `README.md`.
* **V2:** `govulncheck ./...` exits 0 under `go1.27.4` and `go1.26.11`.
* **V3:** build and vet pass under `go1.26.11` and `go1.27.4`, and
  cross-builds pass for linux and windows.
* **V4:** `go mod tidy -diff` is empty.
* **V5:** CI on `main` passes and its `setup-go` log names Go 1.27.4
  (after the coordinator's push).

## Rollout and Rollback

Revert the `go.mod` and `README.md` change by a new commit. Nothing is
tagged, so no consumer is affected (MADR consequences).

## Execution Record

### Approval (2026-11-17)

The owner's rulings D1 and D2 in the MADR: "Their floor is 1.26.11 ...",
as quoted there, and "One short record pair per repository, all approved
now through CI (Recommended)". The MADR is `accepted` and this plan
`in-progress`.

### Baseline (2026-11-17, scratch copy of `23a4cd8`)

* `go build` and `go vet` pass under `go1.26.11` and `go1.27.4`; `go build`
  passes under `go1.26.6`; `go mod tidy -diff` exits 0.
* `govulncheck` v1.9.0: exit 0 under all four of `go1.27.3`, `go1.26.10`,
  `go1.27.4`, `go1.26.11`; the first two report GO-2026-6859 only as an
  uncalled vulnerability in a required module (MADR E4).

### The change and the gates (2026-11-17)

Staged: the record pair only. `go.mod` and `README.md` are changed and
unstaged; nothing is committed, pushed or tagged. HEAD was `23a4cd8`,
clean, level with its upstream, when this began.

**The change.** One script with an `assert` on each old string's count
made both edits: `go.mod:3` `go 1.26.6` to `go 1.26.11`, a blank line
and `toolchain go1.27.4` at `:5`; `README.md:92` "Requires **Go 1.26.6**
or later." to "Requires **Go 1.26.11** or later; built and tested with
Go 1.27.4.". `git diff` was read back. `ci.yml` is unchanged.

**The version-agreement check** (`check_versions.py`, scratch folder, not
in the tree): on the tree `F=1.26.11 T=1.27.4`, "0 failures", exit 0.
Planted, each on a scratch copy of `git archive HEAD` plus the diff, each
exit 1:

* wrong toolchain (`go.mod` `toolchain go1.27.3`): "`go.mod:3/5`: target
  is 1.26.11 / 1.27.4, got 1.26.11 / 1.27.3" and "`README.md:92`: ... does
  not say 'Requires **Go 1.26.11** or later; built and tested with Go
  1.27.3.'";
* wrong floor (`go.mod` `go 1.26.10`): "`go.mod:3/5`: ... got 1.26.10 /
  1.27.4" and "`README.md:92`: ... does not say 'Requires **Go 1.26.10**
  or later; ...'";
* stray old pin (`README.md` "Go 1.27.3."): "`README.md:92`: ... does not
  say '... built and tested with Go 1.27.4.'".

**Gates** (local time, this Mac, exit status taken before any filter;
the two heavy runs went through `run-heavy` and waited in its queue
behind other agents' runs):

| Gate | Start | End | Exit | Result |
| :--- | :--- | :--- | :--- | :--- |
| `go mod tidy -diff`, `GOPROXY=off` | 13:34:01 | 13:34:01 | 0 | empty |
| `gofmt -l .` | 13:34:01 | 13:34:01 | 0 | empty |
| `make vet` | 13:34:01 | 13:34:02 | 0 | clean |
| `make lint` | 13:34:02 | 13:34:04 | 0 | "0 issues." |
| `make test` (`run-heavy`) | 13:34:04 | 13:37:31 | 0 | "[no test files]" |
| floor `go build` and `go vet`, `GOTOOLCHAIN=go1.26.11 GOPROXY=off` | 13:37:31 | 13:37:31 | 0 | clean |
| floor `go test -count=1 -timeout 20m ./...` (`run-heavy`) | 13:37:32 | 13:39:22 | 0 | "[no test files]" |
| `govulncheck ./...`, `go1.27.4`, `GOPROXY=off` | 13:39:22 | 13:39:23 | 0 | "No vulnerabilities found." |
| `govulncheck ./...`, `go1.26.11`, `GOPROXY=off` | 13:39:23 | 13:39:24 | 0 | "No vulnerabilities found." |
| `GOOS=linux CGO_ENABLED=0 go build ./...` | 13:39:24 | 13:39:25 | 0 | clean |
| `GOOS=windows CGO_ENABLED=0 go build ./...` and `go vet ./...` | 13:39:25 | 13:39:26 | 0 | clean |

Not done, and why: no commit, push or tag (the coordinator commits). The
repository has no tests, so the test gates show no behaviour; CI has no
floor job, so the floor is proven only by the local run above (MADR
consequences). The `toolchain` line taking effect on a runner is not
shown locally; CI's `setup-go` log will show it.

## Deviations

### D1 (2026-11-17): CI's golangci-lint pin (v2.13.1) fails under Go 1.27.4

**Found.** After the pair and the change were pushed (`54b8ebf`,
`e9b1ad8`), CI on `e9b1ad8` failed in the step that installs
`golangci-lint@v2.13.1` (`.github/workflows/ci.yml:20`): "could not import
os (/opt/hostedtoolcache/go/1.27.4/x64/src/os/dir.go:8:2 ...", "export
data version 5 is greater th[an the maximum supported version 4]",
"typecheck: 4", "Process completed with exit code 2". The scope table said
`ci.yml` was not changed; that was wrong. Not foreseen: the brief and the
survey did not look at lint tool versions, and the local `make lint` uses
a newer binary. `main` was red from that push until this lands. CI on
`54b8ebf` (Markdown only) passed.

**Decision.** The owner's ruling, label "v2.15.0 everywhere, in each
repository's Go-pins record (Recommended)": pin v2.15.0, bring any new
finding to the owner rather than suppress it. MADR amendment A1 in
`0001-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md` records
it.

**Added to scope.** `.github/workflows/ci.yml:20`, only:
`golangci-lint@v2.13.1` becomes `golangci-lint@v2.15.0` (pinned by tag, as
before). A follow-up commit follows the coordinator's push.

**Checks run (this Mac, exit status before any filter).**

| Check | Start | End | Exit | Result |
| :--- | :--- | :--- | :--- | :--- |
| `go install ...golangci-lint@v2.15.0` into a scratch `GOBIN`, `GOTOOLCHAIN=local` (the module cache lacked it, so the proxy was used for this download) | 13:47:38 | 13:47:43 | 0 | "golangci-lint has version 2.15.0 built with go1.27.4" |
| `make lint` with that binary, Go 1.27.4, `GOPROXY=off` | 13:47:49 | 13:47:56 | 0 | "0 issues." |
| the same under `GOTOOLCHAIN=go1.26.11` | 13:47:56 | 13:48:01 | 0 | "0 issues." |
| `make lint` with v2.13.1 installed the same way, Go 1.27.4 (reproduction) | after 13:48 | | 2 | "typecheck: 4", "export data version" |
| `actionlint .github/workflows/ci.yml` | | | 0 | clean |
| `make vet` | | | 0 | clean |

No finding was reported by v2.15.0, so nothing was fixed, suppressed or
excluded. The version-agreement check gained the lint pin (`v2.15.0`
as the only `golangci-lint@` pin in `ci.yml`): exit 0 on the tree, and
exit 1 on a scratch copy with the pin set back to v2.13.1
("`.github/workflows/ci.yml`: golangci-lint pin is ['v2.13.1'], want
['v2.15.0']").

### A2 (2026-11-17): the floor job

**Decision.** The owner's ruling, label "Only for the two libraries
(mcplib, mcp-buntdb)": add a CI job on the floor. MADR amendment A2 in
`0001-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md` records
it. This is a scope addition, not a failure.

**Added to scope.** `.github/workflows/ci.yml`, one new job `go126`
(`go-version: '1.26.11'`, `GOTOOLCHAIN: local`, the `GOVERSION` test,
`go build ./...`, `go vet ./...`, `go test ./...`; actions by tag). The
Out of scope line "A floor job in CI" no longer holds.

**Checks run.**

| Check | Exit | Result |
| :--- | :--- | :--- |
| the job's steps under `GOTOOLCHAIN=go1.26.11 GOPROXY=off` on this Mac | 0 | "[no test files]" |
| `actionlint .github/workflows/ci.yml` | 0 | clean |
| version-agreement check, extended: `go126`'s `go-version` equals `go.mod`'s `go` line, its `GOVERSION` test names it, `GOTOOLCHAIN: local`, no `go-version-file` in it | 0 | "0 failures" |
| the same on a scratch copy with `go-version: '1.26.10'` | 1 | "go126 go-version is ['1.26.10'], want ['1.26.11']" |
| the same with the test set to `go1.26.10` | 1 | "go126 GOVERSION test is not go1.26.11" |

The `validate` job's `go.mod`-driven CI run on `a453c3b` passed (run
44162759936): `setup-go` printed "Successfully set up Go version 1.27.4",
and the lint step ended "0 issues.". The first CI run on this change's
push is the proof of `go126` on a runner. With no test files, the job
proves only that the module builds and vets under 1.26.11 on Linux.
