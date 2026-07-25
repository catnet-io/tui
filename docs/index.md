---
layout: default
title: catnet-tui — Interactive Terminal UI Scanner
nav_order: 1
description: Keyboard-first interactive Terminal UI for network scanning. Built in Go with Bubble Tea. Powered by catnet-io/engine.
---

> **catnet-tui** — Interactive Terminal UI Scanner. Built in Go with Bubble Tea. Pure consumer of `catnet-io/engine`.

[![Release](https://img.shields.io/github/v/release/catnet-io/tui)](https://github.com/catnet-io/tui/releases)
[![CI](https://github.com/catnet-io/tui/actions/workflows/ci.yml/badge.svg)](https://github.com/catnet-io/tui/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## What catnet-tui does

| 🎯 Interactive Target Input | ⚡ Real-Time Scan Streaming | 🖥️ Live Host Discovery |
|-----------------------------|-----------------------------|------------------------|
| Target auto-detection (`auto`) | Progress bar updating in real time | Live updating host list |
| Clean input interface       | Graceful scan cancellation (`q`/`Esc`) | Rich Lipgloss styling |

## Install in 30 seconds

**macOS / Linux (Homebrew):**
```bash
brew install catnet-io/tap/catnet-tui
```

**Windows (Scoop):**
```powershell
scoop bucket add catnet https://github.com/catnet-io/scoop-bucket
scoop install catnet-tui
```

**Linux / macOS (Binary Download):**
```bash
curl -sSL https://github.com/catnet-io/tui/releases/latest/download/catnet-tui_Linux_x86_64.tar.gz | tar xz
sudo mv catnet-tui /usr/local/bin/
catnet-tui version
```

**Using Go:**
```bash
go install github.com/catnet-io/tui/cmd/catnet-tui@latest
```

## Interactive Workflow & Keybindings

Launch the Terminal UI:
```bash
catnet-tui
```

### Keybindings & Controls

| Key / Action | Description |
|:---:|:---|
| `Enter` | Submit target (e.g. `192.168.1.0/24` or `auto`) and start scanning |
| `q` / `Esc` | Abort active scan and return to target input |
| `Ctrl+C` | Exit application safely without goroutine leaks |

---

## Part of the CatNet Ecosystem

CatNet is a complete network scanning suite designed for terminal users, automation scripts, and graphical desktops.

| | Repository | Role | Description |
|---|---|---|---|
| ⚙️ | [catnet-io/engine](https://github.com/catnet-io/engine) | Shared scanning engine | High-performance, asynchronous scanning library in Go. |
| 💻 | [catnet-io/catnet](https://github.com/catnet-io/catnet) | Scriptable CLI | CLI client optimized for terminal pipelining. |
| 🖥️ | [catnet-io/app](https://github.com/catnet-io/app) | Desktop GUI | Cross-platform desktop application (Wails + React) with local SQLite history and scan comparison diffing. |
| 📟 | [catnet-io/tui](https://github.com/catnet-io/tui) | **Terminal UI** | Keyboard-centric interactive Terminal UI built with Bubble Tea. |

---

- [Full documentation on the Wiki](https://github.com/catnet-io/tui/wiki)
- [GitHub Repository](https://github.com/catnet-io/tui)
- [Report an Issue](https://github.com/catnet-io/tui/issues/new)
