# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added terminal resize awareness (`tea.WindowSizeMsg`) storing dimensions and dynamically resizing UI components and viewports.
- Added ASCII/Lipgloss Network Topology Map in `internal/ui/topology.go` delegating graph construction to `catnet-io/engine/pkg/topology` (`BuildGraph`), featuring viewport scrolling, device role icons, edge classification, unit tests, and `[t]` view mode toggle key.
- Added official webpage and documentation link (https://catnet-io.github.io/tui/) to README.md and repository homepage URL.
- Added GitHub Pages documentation site under `docs/` matching catnet documentation layout.
- Added support for `TAP_GITHUB_TOKEN` secret in release workflow for Homebrew and Scoop publishing.

### Changed

- Refactored TUI key navigation so `q` and `esc` abort active scans back to target input instead of exiting the application.

### Removed

- Removed redundant static `docs/index.html` in favor of Jekyll-rendered `docs/index.md`.

### Fixed

- Fixed target range resolution in TUI scanner by utilizing `targets.ParseRange` from `catnet-io/engine` so `'auto'`, CIDR notation, and IP range strings expand into target IP slices before `ScanStream` execution.
- Filtered host discovery notifications so only active hosts (`Host.Alive == true`) are displayed in scan results.
- Explicitly configured `token` in `.goreleaser.yaml` for Homebrew tap and Scoop bucket repositories to use release workflow token secrets.
- Resolved data race in scan initialization goroutine (`startScan`) when accessing engine and target parameters.

## [0.1.0] - 2026-07-25

### Added

- Target auto-detection (`'auto'`) support in interactive TUI scanner model input.
- Modular UI architecture in `internal/ui` (`model.go`, `scan.go`, `styles.go`).
- Main binary entrypoint at `cmd/catnet-tui/main.go`.
- Real-time engine event consumption via `ScanStream` channel (`listenForEvents`).
- Race-detector unit tests for UI model state transitions, event stream processing, and goroutine scan cancellation in `internal/ui/model_test.go`.
- MIT License file (`LICENSE`).

### Changed

- Upgraded `github.com/catnet-io/engine` dependency to `v0.6.0`.
- Refactored single-file root `main.go` into standard `cmd/catnet-tui` and `internal/ui` package structure.

### Fixed

- Resolved `gosec` file permission error (G306) in `internal/ui/model.go`.
- Resolved `staticcheck` empty branch warning (SA9003) in `internal/ui/scan.go`.
- Restored `TimeoutMs` setting and improved error handling in scan stream execution.
- Updated commit signature validation rules in PR Rules Enforcer workflow for automated PRs.


