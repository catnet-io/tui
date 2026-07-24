# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Modular UI architecture in `internal/ui` (`model.go`, `scan.go`, `styles.go`).
- Main binary entrypoint at `cmd/catnet-tui/main.go`.
- Real-time engine event consumption via `ScanStream` channel (`listenForEvents`).
- Race-detector unit tests for UI model state transitions, event stream processing, and goroutine scan cancellation in `internal/ui/model_test.go`.

### Changed

- Refactored single-file root `main.go` into standard `cmd/catnet-tui` and `internal/ui` package structure.
