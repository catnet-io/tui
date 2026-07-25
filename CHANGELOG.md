# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Refactored TUI key navigation so `q` and `esc` abort active scans back to target input instead of exiting the application.

### Fixed

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


