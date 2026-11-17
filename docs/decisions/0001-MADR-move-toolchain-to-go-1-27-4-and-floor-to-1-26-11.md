---
status: accepted
date: 2026-11-17
decision-makers: mcp-buntdb maintainers
consulted: "govulncheck v1.9.0 (vulnerability database read 2026-11-17) under four toolchains, 2026-11-17; go-selfupdate-lib's 0045-MADR and 0045-PLAN, read-only; the sibling repositories' go.mod files, read-only"
informed: the programs that require this module
---
# Build and test with Go 1.27.4 through the `toolchain` line, and raise the floor to `go 1.26.11`

## Context and Problem Statement

The owner's instruction of 2026-11-17, relayed by the coordinator:

> "go-llmprovider-sdk, go-tui-lib, acp-go-sdk2, gobble-cli, magic-cli-remote, go-selfupdate-lib, mcp-server-duckduckgo, mcp-server-magictools, mcp-server-recall, mcp-server-sequential-thinking, mcplib, mcp-buntdb, and mcp-server-socratic-thinker all need to have their go floor pinned to 1.26.11, and go version/toolchain pinned to 1.27.4."

"Floor" is the `go` directive in `go.mod`; "version/toolchain" is the
`toolchain` line. A read-only survey then found that four of the thirteen
need Go 1.27 and cannot build at 1.26.11. This repository is not one of
them: it builds and vets at 1.26.11 (E3). The owner ruled on the survey,
labels quoted exactly and only as labels:

* **D1, the four that need Go 1.27.** "Their floor is 1.27.4: raise
  go-tui-lib, leave the other three (Recommended)". Asked: for code that
  needs Go 1.27 the floor and the toolchain are the same version; the
  other nine get floor 1.26.11 and toolchain 1.27.4. This repository is
  one of the nine.
* **D2, the seven ready repositories.** "One short record pair per
  repository, all approved now through CI (Recommended)". Asked: a pair
  modelled on go-selfupdate-lib's is drafted in each repository; each is
  committed, executed (`go.mod`, docs, gates) and pushed to its `main` on
  this approval, libraries first then the servers; the coordinator stops
  and comes back to the owner on any failure or anything unexpected; tags
  and releases wait for the owner's word. This repository is one of the
  seven.

Everything in this record other than the quoted instruction and the
quoted labels is evidence or the author's text.

**Evidence index.** Read or observed on 2026-11-17 at HEAD `23a4cd8` on
`main`, level with its upstream ref (the public remote's `main`), tree
clean, unless it says otherwise.

* **E1, the pins [read].** `go.mod:3` is `go 1.26.6`; there is no
  `toolchain` line. `README.md:92` says "Requires **Go 1.26.6** or later."
  `.github/workflows/ci.yml:15` is `go-version-file: go.mod`. A
  whole-repository search for `1.2[67]`, `go-version` and `toolchain`
  (excluding `go.sum` and `.git`) finds only those three lines. The
  `Makefile` names no Go version. The repository has no `AGENTS.md`,
  `CLAUDE.md`, `GEMINI.md`, changelog, `SECURITY.md`, `docs/README.md` or
  other docs index, and no test files.
* **E2, CI [read].** One workflow, `ci.yml`, one job (`validate`) on
  `ubuntu-latest` with `actions/setup-go@v7` and `go-version-file:
  go.mod`. Its README at v7.0.0, as go-selfupdate-lib's 0018-MADR quoted
  it: "If the `toolchain` directive is present, its version is used;
  otherwise, the action falls back to the `go` directive." So after this
  change CI builds with Go 1.27.4; today it builds with 1.26.6. There is
  **no job that builds and tests on the floor.**
* **E3, builds [observed on this Mac, scratch copy].** `go build ./...` and
  `go vet ./...` pass under `GOTOOLCHAIN=go1.26.11 GOPROXY=off` and under
  `go1.27.4`; `go mod tidy -diff` shows nothing.
* **E4, `govulncheck ./...` at `23a4cd8`, v1.9.0 [observed].** Under
  `go1.27.4` and `go1.26.11` it exits 0, "No vulnerabilities found."
  Under `go1.27.3` and `go1.26.10` it **also exits 0**: GO-2026-6859
  ("HTTP/1 server crash due to MaxBytesReader header race in net/http",
  found in `stdlib@go1.27.3` and `stdlib@go1.26.10`, fixed in
  `stdlib@go1.27.4` and `stdlib@go1.26.11`) is reported only as a
  vulnerability in a module the code requires but does not call. So,
  unlike go-selfupdate-lib, this repository has no red gate to turn
  green: the move is the owner's fleet standard, not a fix for a finding
  here.
* **E5, who requires this module [read, read-only `grep` of siblings'
  `go.mod`].**

  | Repository | its `go` line | requires this module |
  | :--- | :--- | :--- |
  | mcp-server-duckduckgo | `go 1.27.4` | `v1.0.1` |
  | mcp-server-socratic-thinker | `go 1.26.6` | `v0.2.0` |
  | mcp-server-brainstorm (outside the thirteen) | `go 1.26.6` | `v0.2.0` |
  | mcp-server-go-modernizer (outside the thirteen) | `go 1.26.6` | `v0.2.0` |
  | mcp-server-magicdev (outside the thirteen) | `go 1.26.6` | `v0.2.0` |

  The latest tag is `v1.0.1`; `HEAD` is four commits after it.
* **E6, records here [observed].** None. A scan of the whole repository
  for `NNNN-{MADR,PLAN,REPORT,GATES}-*.md` finds no file, and there is no
  record tooling, so this pair starts at `0001` in a new
  `docs/decisions/`. `.markdownlint-cli2.jsonc` already excludes MADR and
  PLAN files.

The question: how should this repository's builds and its declared floor
follow the owner's instruction?

## Decision Drivers

* **The owner's instruction and D1, D2:** floor `1.26.11`, toolchain
  `1.27.4`, through CI, no tag.
* **Build and test on a standard library `govulncheck` reports clean** (E4),
  even though nothing here reaches GO-2026-6859.
* **The tree states one floor and one toolchain.**
* **No gate is loosened and no job is added** beyond the change.
* **No compatibility engineering:** the owner is the only user of the
  fleet's repositories.

## Considered Options

* **A. `go 1.26.11` and `toolchain go1.27.4`** (the owner's instruction).
* **B. A different floor:** keep `go 1.26.6` and add only
  `toolchain go1.27.4`.
* **C. Do nothing.**

## Decision Outcome

Chosen option: **"A"**, the owner's instruction as ruled in D1 and D2,
because it puts CI on a toolchain with a clean standard library, declares
a floor that matches the fleet's, and costs three lines.

1. **`go.mod`:** `go 1.26.6` becomes `go 1.26.11` (`:3`), and a
   `toolchain go1.27.4` line is added at `:5`. No `require` changes;
   `go mod tidy -diff` stays clean.
2. **CI:** unchanged. `go-version-file: go.mod` follows the `toolchain`
   line (E2), so CI builds with 1.27.4.
3. **Docs:** `README.md:92` says "Requires **Go 1.26.11** or later; built
   and tested with Go 1.27.4." There is no other current statement.
4. **No release is made by this plan** (see Consequences).

### Consequences

* Good, because CI builds and tests with Go 1.27.4 and the declared floor
  names a patch release without GO-2026-6859.
* Neutral, because E4 shows no finding here: the change does not alter any
  `govulncheck` result.
* Bad, because **CI has no floor job**: nothing builds or tests on 1.26.11
  in CI, so the declared floor is proven only by the local run on this Mac
  (E3). This is a named gap; adding a floor job is for the owner to decide
  fleet-wide, and this record does not add one.
* **A release is not needed for consumers to benefit.** A `toolchain` line
  in a dependency's `go.mod` does not affect the module that requires it:
  consumers build with their own toolchain. The only part of this change
  that reaches a consumer is the floor, and that is a cost, not a benefit:
  a consumer that takes a release carrying `go 1.26.11` has its own `go`
  line raised to at least 1.26.11 by `go get`. If the owner later tags
  one (a patch release, `v1.0.2` as the next number after `v1.0.1`), each
  consumer in E5 must bump its `require` to it to take it. Whether a
  release is wanted is not decided here; the tag waits for the owner's
  word.
* **Would raising the floor break a consumer?** No. mcp-server-duckduckgo
  declares `go 1.27.4`, above 1.26.11. The four on `go 1.26.6`
  (socratic-thinker, brainstorm, go-modernizer, magicdev) declare less than
  1.26.11, and each is on `v0.2.0`, so none is affected until it bumps to
  a release carrying the new floor, at which point its `go` line rises to
  1.26.11 (the fleet's floor for those in the thirteen).

### Confirmation

* `go.mod` says `go 1.26.11` and `toolchain go1.27.4`; `go mod tidy -diff`
  and `go.sum` show no other movement.
* The PLAN's version-agreement check passes on the tree and fails on
  scratch copies with a planted wrong toolchain, a wrong floor, and a
  stray old pin in `README.md`.
* `go build`, `go vet` pass under `GOTOOLCHAIN=go1.26.11` and `go1.27.4`;
  `govulncheck ./...` exits 0 under both, with `GOPROXY=off`.
* CI on `main` passes; its `setup-go` log names Go 1.27.4 (observed after
  the push, by the coordinator).

## Pros and Cons of the Options

### A. Floor 1.26.11 and toolchain 1.27.4 (chosen)

* Good, because it is the owner's instruction, and one pair of numbers
  fleet-wide.
* Good, because it needs no change to CI.
* Bad, because it raises the floor by five patch releases for any future
  consumer (E5), for a library with no reachable finding (E4).

### B. Keep `go 1.26.6`; add only the toolchain line

* Good, because consumers' minimum would not move.
* Bad, because it departs from the owner's instruction, and CI would test
  on 1.27.4 while the floor says 1.26.6, a floor nothing tests.

### C. Do nothing

* Good, because nothing changes and nothing is at risk.
* Bad, because it ignores the owner's instruction, leaves CI on 1.26.6
  with GO-2026-6859 in its standard library (not reachable, E4), and
  leaves this repository the odd one out.

## More Information

### Relation to earlier records

None here: this is the repository's first record (E6). It amends nothing.

### Not verified

* Whether `actions/setup-go@v7` (a moving major tag) still behaves as its
  README at v7.0.0 says (E2): quoted from go-selfupdate-lib's record, not
  re-read, and no CI run was observed before this record.
* That Go 1.27.4 and 1.26.11 install on the runner: the manifest was not
  read here. CI on `main` after the push is the proof.
* What the four commits after `v1.0.1` change, and so whether a release
  would carry anything else.
* That any code here needs the floor 1.26.11: it does not (`go build`
  passes at `go1.26.6` in the same scratch copy, 2026-11-17). The floor is
  the owner's fleet standard, not a requirement of the code.
* Why 1.27.4 and 1.26.11: the owner's instruction is the decision either
  way.

### Links

* [GO-2026-6859](https://pkg.go.dev/vuln/GO-2026-6859).
* [Go downloads](https://go.dev/dl/): `go1.27.4`, `go1.26.11`.

## Amendments

### A1 (2026-11-17): CI's golangci-lint pin moves to v2.15.0

**What this corrects.** Decision Outcome item 2 says CI is unchanged, and
E2 says CI builds with 1.27.4 after the change. The second held; the
first did not. The pair was pushed to `main` as `54b8ebf` (records) and
`e9b1ad8` (`go.mod`, `README.md`). CI on `e9b1ad8` **failed**, in the
step "Run go install
github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1"
(`.github/workflows/ci.yml:20`, exit code 2). The log reads
"could not import os (/opt/hostedtoolcache/go/1.27.4/x64/src/os/dir.go:8:2
...", "could not load export data", "export data version 5 is greater
th[an the maximum supported version 4]" and "typecheck: 4". It also shows
the runner's Go is `1.27.4`: the `toolchain` line took effect there,
which settles the first not-verified item of E2. CI on `54b8ebf`, which
changed Markdown only, passed. So the first push left `main` red until
the change below lands.

**Cause.** golangci-lint v2.13.1, built with Go 1.27.4 by that same step,
cannot read Go 1.27.4's export data; per the coordinator's finding in a
sibling repository, v2.14.0 and v2.15.0 can. Reproduced here: v2.13.1
under Go 1.27.4 exits 2 with four "export data version" typecheck errors
on this repository; v2.15.0 reports 0 issues (A1's change, below). It was
not foreseen: the common brief and the survey looked at Go versions, not
at lint tool versions, and the local `make lint` passed because the local
binary is newer than CI's pin.

**The owner's ruling.** Chosen label: "v2.15.0 everywhere, in each
repository's Go-pins record (Recommended)". Asked: six repositories pin
v2.13.1 in CI, which cannot type-check under Go 1.27.4; each record is
amended to include the CI pin; any new finding the newer linter reports
is brought to the owner with the code rather than suppressed; this
repository gets a follow-up commit under its record.

**The change.** `.github/workflows/ci.yml:20`:
`golangci-lint@v2.13.1` becomes `golangci-lint@v2.15.0`, by tag, as the
line already pinned by tag. Nothing else in the workflow moves. No other
file states the lint version: the Makefile's install hint says `@latest`
and stays. `make lint` with v2.15.0 reports 0 issues under Go 1.27.4 and
under `GOTOOLCHAIN=go1.26.11`.

**Consequence.** The Confirmation gains: CI on `main` passes after this
change, with the `golangci-lint` step on v2.15.0. This amends E2 and
Decision Outcome item 2; it rewrites nothing above.
