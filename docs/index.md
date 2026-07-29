---
layout: default
title: catnet-tui — Interactive Terminal UI Scanner
nav_order: 1
description: Keyboard-first interactive Terminal UI for network scanning. Built in Go with Bubble Tea. Powered by catnet-io/engine.
---

<!-- 
  MAINTENANCE NOTE:
  When re-recording demo/demo.tape in the future, remember to copy the new demo.gif to docs/assets/demo.gif:
  cp demo/demo.gif docs/assets/demo.gif
-->

<section class="hero-section">
  <div class="hero-text">
    <h1 class="hero-title">catnet-tui</h1>
    <p class="hero-subtitle">
      Keyboard-first, interactive Terminal UI for network scanning. Built in Go with Bubble Tea, powered by <code>catnet-io/engine</code>.
    </p>
    <div class="hero-badges">
      <a href="https://github.com/catnet-io/tui/releases" target="_blank" rel="noopener">
        <img src="https://img.shields.io/github/v/release/catnet-io/tui" alt="Release">
      </a>
      <a href="https://github.com/catnet-io/tui/actions/workflows/ci.yml" target="_blank" rel="noopener">
        <img src="https://github.com/catnet-io/tui/actions/workflows/ci.yml/badge.svg" alt="CI">
      </a>
      <a href="https://opensource.org/licenses/MIT" target="_blank" rel="noopener">
        <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT">
      </a>
    </div>
    <div class="hero-actions">
      <a href="#install" class="btn btn-primary">&gt;_ Install Now</a>
      <a href="https://github.com/catnet-io/tui" target="_blank" rel="noopener" class="btn btn-secondary">GitHub Repository</a>
    </div>
  </div>
  <div class="hero-preview">
    <div class="terminal-window">
      <div class="terminal-header">
        <div class="terminal-dots">
          <span class="terminal-dot dot-red"></span>
          <span class="terminal-dot dot-yellow"></span>
          <span class="terminal-dot dot-green"></span>
        </div>
        <div class="terminal-title">catnet-tui — 80x24</div>
      </div>
      <div class="terminal-body">
        <img src="{{ '/assets/demo.gif' | relative_url }}" alt="catnet-tui live terminal scan demonstration">
      </div>
    </div>
  </div>
</section>

<section class="features-grid">
  <div class="feature-card">
    <div class="feature-icon">🎯</div>
    <h3 class="feature-title">Interactive Target Input</h3>
    <p class="feature-desc">Automatic target detection (<code>auto</code>) or explicit CIDR input with clean, keyboard-centered input forms.</p>
  </div>
  <div class="feature-card">
    <div class="feature-icon">⚡</div>
    <h3 class="feature-title">Real-Time Streaming</h3>
    <p class="feature-desc">Live scan progress updating in real time with instant abort capability (<code>q</code> / <code>Esc</code>).</p>
  </div>
  <div class="feature-card">
    <div class="feature-icon">🖥️</div>
    <h3 class="feature-title">Live Host Discovery</h3>
    <p class="feature-desc">Dynamic updating host list with rich Lipgloss terminal styling and clear visual status indicators.</p>
  </div>
</section>

<section id="why-tui">
  <h2>Why TUI?</h2>
  <p>CatNet provides CLI, Desktop, and TUI interfaces. Here is how <code>catnet-tui</code> fits into your workflow:</p>
  
  <div class="why-tui-grid">
    <div class="qa-card">
      <div class="qa-question">❓ Why not just use the CLI (<code>catnet-io/catnet</code>)?</div>
      <div class="qa-answer">
        The CLI is optimized for shell scripts, automated pipelines, and stdout redirection. <code>catnet-tui</code> provides an interactive, visually rich interface directly in your terminal — ideal for live monitoring, rapid target entry, and interactive scan control without writing complex CLI flags.
      </div>
    </div>
    
    <div class="qa-card">
      <div class="qa-question">❓ Does this replace the Desktop App (<code>catnet-io/app</code>)?</div>
      <div class="qa-answer">
        No. The Desktop App (Wails + React) is designed for graphical environments, offering persistent SQLite scan history and visual comparison diffing. <code>catnet-tui</code> brings a fast, lightweight terminal UI for SSH sessions, headless servers, and terminal purists.
      </div>
    </div>
    
    <div class="qa-card">
      <div class="qa-question">❓ How does it connect to the scanning engine?</div>
      <div class="qa-answer">
        <code>catnet-tui</code> contains zero scanning logic. It acts as a pure presentation layer consuming live event channels from <code>catnet-io/engine</code> (via <code>ScanStream</code>), ensuring high performance and identical scan results across all CatNet clients.
      </div>
    </div>
  </div>
</section>

<section id="install">
  <h2>Install in 30 Seconds</h2>
  <p>Choose your preferred installation method below:</p>

  <div class="tab-container">
    <div class="tab-nav">
      <button class="tab-button active" data-tab="tab-brew">Homebrew (macOS/Linux)</button>
      <button class="tab-button" data-tab="tab-scoop">Scoop (Windows)</button>
      <button class="tab-button" data-tab="tab-binary">Binary Download</button>
      <button class="tab-button" data-tab="tab-go">Go Install</button>
    </div>

    <!-- Homebrew Panel -->
    <div id="tab-brew" class="tab-panel active">
      <div class="code-block-wrapper">
        <button class="copy-btn" data-copy-target="code-brew">Copy</button>
        <pre><code id="code-brew">brew install catnet-io/tap/catnet-tui</code></pre>
      </div>
    </div>

    <!-- Scoop Panel -->
    <div id="tab-scoop" class="tab-panel">
      <div class="code-block-wrapper">
        <button class="copy-btn" data-copy-target="code-scoop">Copy</button>
        <pre><code id="code-scoop">scoop bucket add catnet https://github.com/catnet-io/scoop-bucket
scoop install catnet-tui</code></pre>
      </div>
    </div>

    <!-- Binary Download Panel -->
    <div id="tab-binary" class="tab-panel">
      <div class="code-block-wrapper">
        <button class="copy-btn" data-copy-target="code-binary">Copy</button>
        <pre><code id="code-binary">curl -sSL https://github.com/catnet-io/tui/releases/latest/download/catnet-tui_Linux_x86_64.tar.gz | tar xz
sudo mv catnet-tui /usr/local/bin/
catnet-tui version</code></pre>
      </div>
    </div>

    <!-- Go Install Panel -->
    <div id="tab-go" class="tab-panel">
      <div class="code-block-wrapper">
        <button class="copy-btn" data-copy-target="code-go">Copy</button>
        <pre><code id="code-go">go install github.com/catnet-io/tui/cmd/catnet-tui@latest</code></pre>
      </div>
    </div>
  </div>
</section>

<section id="workflow">
  <h2>Interactive Workflow & Keybindings</h2>
  <p>Launch the Terminal UI by running:</p>

  <div class="code-block-wrapper" style="margin-bottom: 1.5rem;">
    <button class="copy-btn" data-copy-target="code-launch">Copy</button>
    <pre><code id="code-launch">catnet-tui</code></pre>
  </div>

  <h3>Keybindings & Controls</h3>
  <table>
    <thead>
      <tr>
        <th style="text-align: center; width: 140px;">Key / Action</th>
        <th>Description</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td style="text-align: center;"><kbd>Enter</kbd></td>
        <td>Submit target (e.g. <code>192.168.1.0/24</code> or <code>auto</code>) and start scanning</td>
      </tr>
      <tr>
        <td style="text-align: center;"><kbd>q</kbd> / <kbd>Esc</kbd></td>
        <td>Abort active scan and return to target input</td>
      </tr>
      <tr>
        <td style="text-align: center;"><kbd>Ctrl+C</kbd></td>
        <td>Exit application safely without goroutine leaks</td>
      </tr>
    </tbody>
  </table>
</section>

<section id="ecosystem">
  <h2>Part of the CatNet Ecosystem</h2>
  <p>CatNet is a complete network scanning suite designed for terminal users, automation scripts, and graphical desktops.</p>

  <table>
    <thead>
      <tr>
        <th style="width: 40px;"></th>
        <th>Repository</th>
        <th>Role</th>
        <th>Description</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>⚙️</td>
        <td><a href="https://github.com/catnet-io/engine" target="_blank" rel="noopener">catnet-io/engine</a></td>
        <td>Shared scanning engine</td>
        <td>High-performance, asynchronous scanning library in Go.</td>
      </tr>
      <tr>
        <td>💻</td>
        <td><a href="https://github.com/catnet-io/catnet" target="_blank" rel="noopener">catnet-io/catnet</a></td>
        <td>Scriptable CLI</td>
        <td>CLI client optimized for terminal pipelining.</td>
      </tr>
      <tr>
        <td>🖥️</td>
        <td><a href="https://github.com/catnet-io/app" target="_blank" rel="noopener">catnet-io/app</a></td>
        <td>Desktop GUI</td>
        <td>Cross-platform desktop application (Wails + React) with local SQLite history and scan comparison diffing.</td>
      </tr>
      <tr>
        <td>📟</td>
        <td><a href="https://github.com/catnet-io/tui" target="_blank" rel="noopener">catnet-io/tui</a></td>
        <td><strong>Terminal UI</strong></td>
        <td>Keyboard-centric interactive Terminal UI built with Bubble Tea.</td>
      </tr>
    </tbody>
  </table>
</section>

<hr>

<section id="links" style="display: flex; gap: 1.5rem; flex-wrap: wrap;">
  <a href="https://github.com/catnet-io/tui/wiki" target="_blank" rel="noopener">📚 Full documentation on the Wiki</a>
  <a href="https://github.com/catnet-io/tui" target="_blank" rel="noopener">💻 GitHub Repository</a>
  <a href="https://github.com/catnet-io/tui/issues/new" target="_blank" rel="noopener">🐛 Report an Issue</a>
</section>
