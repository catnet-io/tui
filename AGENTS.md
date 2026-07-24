# AGENTS.md — catnet-io/tui

This file provides persistent context for AI coding agents (Antigravity, Jules, OpenHands, Claude Code) working in the `catnet-io/tui` repository.

---

## What this repository is

`catnet-io/tui` is the terminal UI frontend for CatNet.
Built with Bubble Tea (Go). It is a pure consumer of `catnet-io/engine`.
This repository is in **bootstrap phase** — the MVP has not been delivered yet.

**Module path:** `github.com/catnet-io/tui`  
**Binary name:** `catnet-tui`  
**Go version:** 1.26.4  
**Engine dependency:** `github.com/catnet-io/engine@v0.5.x` (use latest stable)  
**Visibility:** Private until MVP is functional

---

## MVP definition (what must exist before going public)

1. Single TUI screen: target input → scan → live host list
2. Progress bar updating in real time from `pkg/scan.Engine.ScanStream`
3. `q` or `Ctrl+C` cancels scan without goroutine leak
4. CI passing (`go build`, `go test -race`, `go vet`, `golangci-lint`, `govulncheck`)

Do not add features beyond this list until the MVP checklist is complete.

---

## Architecture (target — to be built)

```
cmd/catnet-tui/
  └── main.go
internal/ui/
  ├── model.go        ← Bubble Tea model: Init(), Update(), View()
  ├── scan.go         ← integration with pkg/scan.Engine via channel
  └── styles.go       ← lipgloss style definitions
```

### Engine API to use

Use `pkg/scan.Engine.ScanStream` (channel-based API) exclusively.
**Never use `pkg/engine.StartScan`** (callback API) in this repository.

Event flow:
```
scan.Engine.ScanStream(ctx, ips, profile, eventChan)
  → goroutine drains eventChan
  → sends tea.Msg to Bubble Tea update loop
  → model.View() re-renders with new state
```

### Bubble Tea pattern for scan events

```go
// internal/ui/scan.go
type hostDiscoveredMsg results.HostResult
type scanProgressMsg float64
type scanDoneMsg struct{ err error }

func listenForEvents(ch <-chan events.Event) tea.Cmd {
    return func() tea.Msg {
        ev, ok := <-ch
        if !ok {
            return scanDoneMsg{}
        }
        switch ev.Type {
        case events.HostDiscovered:
            data := ev.Data.(events.HostDiscoveredData)
            return hostDiscoveredMsg(data.Host)
        case events.ScanProgress:
            data := ev.Data.(events.ProgressData)
            return scanProgressMsg(data.Ratio)
        case events.ScanCompleted:
            return scanDoneMsg{}
        }
        return nil
    }
}
```

---

## Hard rules — never violate

1. **No scanning logic in this repository.** All scanning happens in `catnet-io/engine`.
2. **Use `pkg/scan.Engine.ScanStream` only.** Never `pkg/engine.StartScan`.
3. **No CGO.**
4. **English only everywhere.** All code, comments, godoc, log messages, error strings, PR descriptions, PR review comments, commit messages, and documentation across the entire repository must be in English. Portuguese and non-English text are strictly forbidden.
5. **No local `replace` directives committed to `main`.** Use `scripts/dev-replace.sh on/off` to toggle during local development.
6. **Do not go public before MVP checklist is complete.**
7. **Immutable GitHub Action Pinning.** All GitHub Actions in workflow files must be pinned to full 40-character commit SHAs. Never use unpinned tags (e.g. `@v4`).
8. **No `squash and merge`.** Never use squash merges (`gh pr merge --squash` or GitHub UI squash) for PRs in this repository. All PR merges must preserve atomic commit history via merge commits (`gh pr merge --merge`) or rebase merges (`gh pr merge --rebase`) to maintain DevSecOps traceability, commit provenance, and auditability.

---

## Conventions

### Commit messages — Conventional Commits

```
feat(ui): add live host list updating from scan events
fix(ui): fix goroutine leak on Ctrl+C during scan
chore(deps): add charmbracelet/bubbletea v1.x
test(ui): add test for scan cancellation without leak
```

Scopes: `ui`, `scan`, `styles`, `deps`, `ci`, `docs`.

### Changelog — Keep a Changelog

Every PR that changes behavior must update `CHANGELOG.md` under `[Unreleased]`.
Sections: `Added`, `Changed`, `Fixed`, `Security`, `Deprecated`, `Removed`.
Breaking changes must start with `**BREAKING CHANGE**:`.

### Go style & Testing

- `gofmt` and `goimports` on all files.
- `golangci-lint` must pass (see `.golangci.yml` for enabled linters).
- All new public functions must have unit tests.
- Concurrency-sensitive code must be tested with `-race`.
- Prefer `context.Context` as first parameter in all public functions that do I/O or async tasks.

---

## CI requirements — all must pass before merge / going public

- `go build ./...`
- `go test -race ./...`
- `go vet ./...`
- `golangci-lint run` (via `golangci-lint.yml` workflow)
- `govulncheck ./...`

