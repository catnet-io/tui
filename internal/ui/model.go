package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/catnet-io/engine/pkg/events"
	"github.com/catnet-io/engine/pkg/results"
	"github.com/catnet-io/engine/pkg/scan"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	stateInput sessionState = iota
	stateScanning
	stateResults
)

type viewMode int

const (
	viewBoth viewMode = iota
	viewTable
	viewMap
)

// Model represents the Bubble Tea model for the CatNet TUI.
type Model struct {
	state         sessionState
	viewMode      viewMode
	textInput     textinput.Model
	progressBar   progress.Model
	viewport      viewport.Model
	viewportReady bool
	width         int
	height        int
	progress      float64
	devices       []results.HostResult
	selectedIdx   int
	errorMsg      string
	targetRange   string
	logMsgs       []string

	engine    *scan.Engine
	cancelFn  context.CancelFunc
	eventChan chan events.Event
}

// InitialModel returns a new Model initialized for starting the TUI.
func InitialModel() Model {
	return NewModel(scan.NewEngine())
}

// NewModel creates a new Model with the provided scan Engine instance.
func NewModel(engine *scan.Engine) Model {
	ti := textinput.New()
	ti.Placeholder = "auto (or e.g. 192.168.1.1-254)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	pg := progress.New(progress.WithDefaultGradient())

	return Model{
		state:       stateInput,
		textInput:   ti,
		progressBar: pg,
		engine:      engine,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) exportResults() error {
	data, err := json.MarshalIndent(m.devices, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("catnet_export.json", data, 0600)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if msg.Width > 20 {
			m.progressBar.Width = msg.Width - 20
		}
		vpHeight := m.height - 14
		if vpHeight < 5 {
			vpHeight = 5
		}
		if !m.viewportReady {
			m.viewport = viewport.New(m.width, vpHeight)
			m.viewportReady = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = vpHeight
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.state == stateScanning && m.cancelFn != nil {
				m.cancelFn()
				m.cancelFn = nil
			}
			return m, tea.Quit

		case "esc":
			if m.state == stateScanning {
				if m.cancelFn != nil {
					m.cancelFn()
					m.cancelFn = nil
				}
				m.state = stateInput
				m.textInput.Focus()
				return m, nil
			}
			return m, tea.Quit

		case "q":
			if m.state == stateScanning {
				if m.cancelFn != nil {
					m.cancelFn()
					m.cancelFn = nil
				}
				m.state = stateInput
				m.textInput.Focus()
				return m, nil
			}
			if m.state == stateResults {
				m.state = stateInput
				m.devices = nil
				m.progress = 0
				m.logMsgs = nil
				m.textInput.Focus()
				return m, nil
			}

		case "enter":
			if m.state == stateInput {
				m.targetRange = strings.ToLower(strings.TrimSpace(m.textInput.Value()))
				if m.targetRange == "" {
					m.targetRange = "auto"
				}
				m.state = stateScanning
				m.progress = 0
				m.devices = nil
				m.logMsgs = []string{"Initializing scan..."}
				return m, m.startScan()
			}

		case "up", "k":
			if m.state == stateResults && len(m.devices) > 0 {
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			}

		case "down", "j":
			if m.state == stateResults && len(m.devices) > 0 {
				if m.selectedIdx < len(m.devices)-1 {
					m.selectedIdx++
				}
			}

		case "e":
			if m.state == stateResults && len(m.devices) > 0 {
				err := m.exportResults()
				if err != nil {
					m.logMsgs = append(m.logMsgs, fmt.Sprintf("Export error: %v", err))
				} else {
					m.logMsgs = append(m.logMsgs, "Exported to catnet_export.json")
				}
			}

		case "t", "m":
			if m.state == stateResults {
				m.viewMode = (m.viewMode + 1) % 3
			}
		}

	case hostDiscoveredMsg:
		if msg.Alive {
			m.devices = append(m.devices, results.HostResult(msg))
			m.logMsgs = append(m.logMsgs, fmt.Sprintf("Host: %s (%s)", msg.IP, msg.Hostname))
		}
		return m, listenForEvents(m.eventChan)

	case scanProgressMsg:
		m.progress = float64(msg)
		return m, listenForEvents(m.eventChan)

	case scanDoneMsg:
		if m.state == stateScanning {
			m.state = stateResults
			if msg.err != nil {
				m.errorMsg = msg.err.Error()
			}
		}
		return m, nil
	}

	if m.state == stateInput {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	// Title Banner
	s.WriteString(titleStyle.Render("⚡ CATNET TERMINAL SCANNER ⚡"))
	s.WriteString("\n\n")

	switch m.state {
	case stateInput:
		s.WriteString(greyText.Render("Enter IP range, CIDR, or 'auto' to scan (default: auto):"))
		s.WriteString("\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString(greyText.Render("Press [ENTER] to start scan • [ESC] to quit"))

	case stateScanning:
		s.WriteString(cyanText.Render(fmt.Sprintf("Scanning: %s", m.targetRange)))
		s.WriteString("\n\n")

		// Progress bar
		s.WriteString(m.progressBar.ViewAs(m.progress))
		s.WriteString(fmt.Sprintf("  %.0f%%", m.progress*100))
		s.WriteString("\n\n")

		// Display last 5 logs
		s.WriteString(greyText.Render("Live Logs:"))
		s.WriteString("\n")
		logStart := len(m.logMsgs) - 5
		if logStart < 0 {
			logStart = 0
		}
		for i := logStart; i < len(m.logMsgs); i++ {
			s.WriteString(fmt.Sprintf("  %s\n", m.logMsgs[i]))
		}
		s.WriteString("\n")
		s.WriteString(redText.Render("Press [ESC] or [q] to abort scan"))

	case stateResults:
		s.WriteString(greenText.Render(fmt.Sprintf("Scan completed for: %s", m.targetRange)))
		s.WriteString("\n\n")

		if len(m.devices) == 0 {
			s.WriteString(redText.Render("No active hosts discovered."))
			s.WriteString("\n\n")
		} else {
			if m.viewMode == viewBoth || m.viewMode == viewTable {
				// Render Table header
				s.WriteString(headerStyle.Render(fmt.Sprintf("  %-16s %-25s %-20s %-15s", "IP Address", "Hostname", "MAC", "Ports")))
				s.WriteString("\n")

				// Render Table rows
				for i, dev := range m.devices {
					portsStr := ""
					if len(dev.OpenPorts) > 0 {
						var ports []string
						for _, p := range dev.OpenPorts {
							ports = append(ports, fmt.Sprintf("%d", p))
						}
						portsStr = strings.Join(ports, ",")
					} else {
						portsStr = "None"
					}

					rowContent := fmt.Sprintf("  %-16s %-25s %-20s %-15s", dev.IP, truncate(dev.Hostname, 23), dev.MAC, truncate(portsStr, 14))
					if i == m.selectedIdx {
						s.WriteString(selectedRowStyle.Render(rowContent))
					} else {
						s.WriteString(normalRowStyle.Render(rowContent))
					}
					s.WriteString("\n")
				}
				s.WriteString("\n")
			}

			if m.viewMode == viewBoth || m.viewMode == viewMap {
				mapContent := RenderNetworkMap(m.devices, m.targetRange, m.selectedIdx, m.width)
				if m.width > 0 && m.height > 0 {
					vpHeight := m.height - 7
					if m.viewMode == viewBoth {
						tableLines := len(m.devices) + 3
						vpHeight = m.height - 7 - tableLines
					}
					if vpHeight < 4 {
						vpHeight = 4
					}
					m.viewport.Width = m.width
					m.viewport.Height = vpHeight
					m.viewport.SetContent(mapContent)
					s.WriteString(m.viewport.View())
				} else {
					s.WriteString(mapContent)
				}
				s.WriteString("\n")
			}
		}

		if len(m.logMsgs) > 0 {
			lastLog := m.logMsgs[len(m.logMsgs)-1]
			if strings.HasPrefix(lastLog, "Exported to") {
				s.WriteString(greenText.Render(fmt.Sprintf("✓ %s", lastLog)))
				s.WriteString("\n\n")
			} else if strings.HasPrefix(lastLog, "Export error:") {
				s.WriteString(redText.Render(fmt.Sprintf("✗ %s", lastLog)))
				s.WriteString("\n\n")
			}
		}

		// Footer navigation
		footerText := "Navigation: [↑/↓] select host • [t] toggle map • [e] export results • [q] new scan • [ESC] exit"
		if m.width > 0 && m.width < 90 {
			footerText = "Nav: [↑/↓] select • [t] map • [e] export • [q] new • [ESC] exit"
		}
		s.WriteString(greyText.Render(footerText))
	}

	return s.String()
}
