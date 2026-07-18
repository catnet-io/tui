# AGENTS.md — catnet-io/tui

This file provides persistent context for AI coding agents working in `catnet-io/tui`.

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
4. CI passing (`go build`, `go test -race`, `go vet`)

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
4. **English only** in all Go source files.
5. **No local `replace` directives in `main` branch.**
6. **Do not go public before MVP checklist is complete.**

---

## Conventions

### Commit messages — Conventional Commits

```
feat(ui): add live host list updating from scan events
fix(ui): fix goroutine leak on Ctrl+C during scan
chore(deps): add charmbracelet/bubbletea v1.x
test(ui): add test for scan cancellation without leak
```

Scopes: `ui`, `scan`, `styles`, `deps`, `ci`.

### Changelog — Keep a Changelog

Start `CHANGELOG.md` from the first commit. Include `[Unreleased]` section.

---

## CI requirements — all must pass before going public

- `go build ./...`
- `go test -race ./...`
- `go vet ./...`
