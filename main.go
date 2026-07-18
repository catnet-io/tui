package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/catnet-io/engine/pkg/events"
	"github.com/catnet-io/engine/pkg/profile"
	"github.com/catnet-io/engine/pkg/results"
	"github.com/catnet-io/engine/pkg/scan"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateInput sessionState = iota
	stateScanning
	stateResults
)

// Msg types
type readEventMsg struct {
	event events.Event
	ok    bool
}

type model struct {
	state       sessionState
	textInput   textinput.Model
	progressBar progress.Model
	progress    float64
	devices     []results.HostResult
	selectedIdx int
	errorMsg    string
	targetRange string
	logMsgs     []string

	engine    *scan.Engine
	cancelFn  context.CancelFunc
	eventChan chan events.Event
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "192.168.1.1-254"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	pg := progress.New(progress.WithDefaultGradient())

	return model{
		state:       stateInput,
		textInput:   ti,
		progressBar: pg,
		engine:      scan.NewEngine(),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func readEvent(ch chan events.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		return readEventMsg{event: ev, ok: ok}
	}
}

func (m *model) startScan() tea.Cmd {
	m.eventChan = make(chan events.Event)
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFn = cancel

	// Start engine scan in background
	go func() {
		defer cancel()
		cfg := profile.DefaultProfile()
		cfg.Concurrency = 32
		cfg.TimeoutMs = 1000
		_ = m.engine.ScanStream(ctx, []string{m.targetRange}, cfg, m.eventChan)
		close(m.eventChan)
	}()

	return readEvent(m.eventChan)
}

func (m model) exportResults() error {
	data, err := json.MarshalIndent(m.devices, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("catnet_export.json", data, 0644)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.state == stateScanning && m.cancelFn != nil {
				m.cancelFn()
			}
			return m, tea.Quit

		case "q":
			if m.state == stateResults {
				m.state = stateInput
				m.devices = nil
				m.progress = 0
				m.logMsgs = nil
				m.textInput.Focus()
				return m, nil
			}
			if m.state == stateInput {
				return m, tea.Quit
			}

		case "enter":
			if m.state == stateInput {
				m.targetRange = m.textInput.Value()
				if m.targetRange == "" {
					m.targetRange = "192.168.1.1-254"
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
		}

	case readEventMsg:
		if !msg.ok {
			m.state = stateResults
			return m, nil
		}

		// Process event
		switch msg.event.Type {
		case events.ScanProgress:
			if data, ok := msg.event.Data.(events.ProgressData); ok {
				m.progress = data.Ratio
			}
		case events.HostDiscovered:
			if data, ok := msg.event.Data.(events.HostDiscoveredData); ok {
				m.devices = append(m.devices, data.Host)
				m.logMsgs = append(m.logMsgs, fmt.Sprintf("Host: %s (%s)", data.Host.IP, data.Host.Hostname))
			}
		case events.ScanCompleted:
			m.state = stateResults
		}

		// Read next event!
		return m, readEvent(m.eventChan)
	}

	if m.state == stateInput {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#66fcf1")).
			Bold(true).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#45a29e")).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#45a29e"))

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0b0c10")).
				Background(lipgloss.Color("#66fcf1")).
				Bold(true)

	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c5c6c7"))

	cyanText  = lipgloss.NewStyle().Foreground(lipgloss.Color("#66fcf1"))
	greyText  = lipgloss.NewStyle().Foreground(lipgloss.Color("#45a29e"))
	redText   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0055"))
	greenText = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff88"))
)

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l-3] + "..."
	}
	return s
}

func (m model) View() string {
	var s strings.Builder

	// Title Banner
	s.WriteString(titleStyle.Render("⚡ CATNET TERMINAL SCANNER ⚡"))
	s.WriteString("\n\n")

	switch m.state {
	case stateInput:
		s.WriteString(greyText.Render("Enter IP range or CIDR to scan (default: 192.168.1.1-254):"))
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
		s.WriteString(redText.Render("Press [ESC] to abort scan"))

	case stateResults:
		s.WriteString(greenText.Render(fmt.Sprintf("Scan completed for: %s", m.targetRange)))
		s.WriteString("\n\n")

		if len(m.devices) == 0 {
			s.WriteString(redText.Render("No active hosts discovered."))
			s.WriteString("\n\n")
		} else {
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

		if len(m.logMsgs) > 0 && m.logMsgs[len(m.logMsgs)-1] == "Exported to catnet_export.json" {
			s.WriteString(greenText.Render("✓ Results successfully exported to catnet_export.json"))
			s.WriteString("\n\n")
		}

		// Footer navigation
		s.WriteString(greyText.Render("Navigation: [↑/↓] select host • [e] export results • [q] new scan • [ESC] exit"))
	}

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
