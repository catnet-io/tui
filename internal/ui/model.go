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

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Filter    key.Binding
	ToggleMap key.Binding
	Export    key.Binding
	NewScan   key.Binding
	Help      key.Binding
	Quit      key.Binding
	Esc       key.Binding
	Enter     key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "select host"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "select host"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ToggleMap: key.NewBinding(
			key.WithKeys("t", "m"),
			key.WithHelp("t", "toggle map"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export"),
		),
		NewScan: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "new scan"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "exit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "start scan"),
		),
	}
}

func (k keyMap) ShortHelpInput() []key.Binding {
	return []key.Binding{k.Enter, k.Help, k.Esc}
}

func (k keyMap) ShortHelpScanning() []key.Binding {
	return []key.Binding{k.Esc, k.Help, k.Quit}
}

func (k keyMap) ShortHelpResults() []key.Binding {
	return []key.Binding{k.Up, k.Filter, k.ToggleMap, k.Export, k.NewScan, k.Help, k.Esc}
}

// Model represents the Bubble Tea model for the CatNet TUI.
type Model struct {
	state         sessionState
	viewMode      viewMode
	textInput     textinput.Model
	filterInput   textinput.Model
	isFiltering   bool
	showHelp      bool
	helpModel     help.Model
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

	fi := textinput.New()
	fi.Prompt = "/ "
	fi.Placeholder = "filter (IP, hostname, MAC)..."
	fi.CharLimit = 64
	fi.Width = 30

	pg := progress.New(progress.WithDefaultGradient())
	h := help.New()

	return Model{
		state:       stateInput,
		textInput:   ti,
		filterInput: fi,
		progressBar: pg,
		helpModel:   h,
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

func (m Model) filteredDevices() []results.HostResult {
	query := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	if query == "" {
		return m.devices
	}
	var res []results.HostResult
	for _, dev := range m.devices {
		if strings.Contains(strings.ToLower(dev.IP), query) ||
			strings.Contains(strings.ToLower(dev.Hostname), query) ||
			strings.Contains(strings.ToLower(dev.MAC), query) {
			res = append(res, dev)
		}
	}
	return res
}

func (m *Model) updateViewportContent() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	vpHeight := m.height - 14
	devices := m.filteredDevices()
	if m.state == stateResults {
		vpHeight = m.height - 7
		if m.viewMode == viewBoth {
			tableLines := len(devices) + 3
			if m.isFiltering || m.filterInput.Value() != "" {
				tableLines += 2
			}
			vpHeight = m.height - 7 - tableLines
		}
	}
	if vpHeight < 4 {
		vpHeight = 4
	}

	if !m.viewportReady {
		m.viewport = viewport.New(m.width, vpHeight)
		m.viewportReady = true
	} else {
		m.viewport.Width = m.width
		m.viewport.Height = vpHeight
	}

	if m.state == stateResults {
		mapContent := RenderNetworkMap(devices, m.targetRange, m.selectedIdx, m.width)
		m.viewport.SetContent(mapContent)
	}
}

func (m Model) renderHelpModal() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("⚡ KEYBOARD SHORTCUTS HELP ⚡"))
	b.WriteString("\n\n")

	switch m.state {
	case stateInput:
		b.WriteString(cyanText.Render("Target Input Mode:\n"))
		b.WriteString("  [ENTER]      Start scanning target range\n")
		b.WriteString("  [?]          Toggle help overlay\n")
		b.WriteString("  [ESC/CTRL+C] Quit application\n")
	case stateScanning:
		b.WriteString(cyanText.Render("Scanning Mode:\n"))
		b.WriteString("  [ESC / q]    Abort active scan\n")
		b.WriteString("  [?]          Toggle help overlay\n")
		b.WriteString("  [CTRL+C]     Quit application\n")
	case stateResults:
		b.WriteString(cyanText.Render("Scan Results Navigation:\n"))
		b.WriteString("  [↑/↓] or [k/j]  Select host in table\n")
		b.WriteString("  [/]             Filter hosts by IP, Hostname, or MAC\n")
		b.WriteString("  [t] or [m]      Toggle view mode (Table / Map / Both)\n")
		b.WriteString("  [e]             Export results to catnet_export.json\n")
		b.WriteString("  [q]             New scan\n")
		b.WriteString("  [?]             Toggle help overlay\n")
		b.WriteString("  [ESC]           Clear filter / Exit\n")
		b.WriteString("  [CTRL+C]        Quit application\n")
	}

	b.WriteString("\n")
	b.WriteString(greyText.Render("Press [ESC] or [?] to close help"))

	return helpModalStyle.Render(b.String()) + "\n"
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
		m.updateViewportContent()

	case tea.KeyMsg:
		if m.showHelp {
			if msg.String() == "esc" || msg.String() == "?" {
				m.showHelp = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "?":
			if !m.isFiltering {
				m.showHelp = true
				return m, nil
			}

		case "ctrl+c":
			if m.state == stateScanning && m.cancelFn != nil {
				m.cancelFn()
				m.cancelFn = nil
			}
			return m, tea.Quit

		case "esc":
			if m.isFiltering {
				m.isFiltering = false
				m.filterInput.SetValue("")
				m.filterInput.Blur()
				m.selectedIdx = 0
				m.updateViewportContent()
				return m, nil
			}
			if m.state == stateResults && m.filterInput.Value() != "" {
				m.filterInput.SetValue("")
				m.selectedIdx = 0
				m.updateViewportContent()
				return m, nil
			}
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

		case "/":
			if m.state == stateResults && !m.isFiltering {
				m.isFiltering = true
				m.filterInput.Focus()
				return m, textinput.Blink
			}

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
			if m.state == stateResults && !m.isFiltering {
				m.state = stateInput
				m.devices = nil
				m.progress = 0
				m.logMsgs = nil
				m.filterInput.SetValue("")
				m.isFiltering = false
				m.textInput.Focus()
				return m, nil
			}

		case "enter":
			if m.isFiltering {
				m.isFiltering = false
				m.filterInput.Blur()
				return m, nil
			}
			if m.state == stateInput {
				m.targetRange = strings.ToLower(strings.TrimSpace(m.textInput.Value()))
				if m.targetRange == "" {
					m.targetRange = "auto"
				}
				m.state = stateScanning
				m.progress = 0
				m.devices = nil
				m.errorMsg = ""
				m.logMsgs = []string{"Initializing scan..."}
				return m, m.startScan()
			}

		case "up", "k":
			if m.state == stateResults && !m.isFiltering {
				devs := m.filteredDevices()
				if len(devs) > 0 && m.selectedIdx > 0 {
					m.selectedIdx--
					m.updateViewportContent()
				}
			}

		case "down", "j":
			if m.state == stateResults && !m.isFiltering {
				devs := m.filteredDevices()
				if len(devs) > 0 && m.selectedIdx < len(devs)-1 {
					m.selectedIdx++
					m.updateViewportContent()
				}
			}

		case "e":
			if m.state == stateResults && !m.isFiltering && len(m.devices) > 0 {
				err := m.exportResults()
				if err != nil {
					m.logMsgs = append(m.logMsgs, fmt.Sprintf("Export error: %v", err))
				} else {
					m.logMsgs = append(m.logMsgs, "Exported to catnet_export.json")
				}
			}

		case "t", "m":
			if m.state == stateResults && !m.isFiltering {
				m.viewMode = (m.viewMode + 1) % 3
				m.updateViewportContent()
			}
		}

		if m.isFiltering {
			m.filterInput, cmd = m.filterInput.Update(msg)
			devs := m.filteredDevices()
			if m.selectedIdx >= len(devs) && len(devs) > 0 {
				m.selectedIdx = len(devs) - 1
			} else if len(devs) == 0 {
				m.selectedIdx = 0
			}
			m.updateViewportContent()
			return m, cmd
		}

	case hostDiscoveredMsg:
		if msg.Alive {
			m.devices = append(m.devices, results.HostResult(msg))
			m.logMsgs = append(m.logMsgs, fmt.Sprintf("Host: %s (%s)", msg.IP, msg.Hostname))
			m.updateViewportContent()
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
				m.logMsgs = append(m.logMsgs, fmt.Sprintf("Scan error: %v", msg.err))
			}
			m.updateViewportContent()
		}
		return m, nil
	}

	if m.state == stateInput {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	if m.state == stateResults && m.viewportReady {
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, tea.Batch(cmd, vpCmd)
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	// Title Banner
	s.WriteString(titleStyle.Render("⚡ CATNET TERMINAL SCANNER ⚡"))
	s.WriteString("\n\n")

	if m.showHelp {
		s.WriteString(m.renderHelpModal())
		return s.String()
	}

	switch m.state {
	case stateInput:
		s.WriteString(greyText.Render("Enter IP range, CIDR, or 'auto' to scan (default: auto):"))
		s.WriteString("\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")

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

	case stateResults:
		s.WriteString(greenText.Render(fmt.Sprintf("Scan completed for: %s", m.targetRange)))
		s.WriteString("\n\n")

		if m.errorMsg != "" {
			s.WriteString(redText.Render(fmt.Sprintf("Scan Error: %s", m.errorMsg)))
			s.WriteString("\n\n")
		}

		if m.isFiltering || m.filterInput.Value() != "" {
			s.WriteString(filterBoxStyle.Render(fmt.Sprintf("Filter: %s", m.filterInput.View())))
			s.WriteString("\n\n")
		}

		devices := m.filteredDevices()

		if len(m.devices) == 0 {
			s.WriteString(redText.Render("No active hosts discovered."))
			s.WriteString("\n\n")
		} else if len(devices) == 0 {
			s.WriteString(redText.Render(fmt.Sprintf("No active hosts matching filter '%s'.", m.filterInput.Value())))
			s.WriteString("\n\n")
		} else {
			if m.viewMode == viewBoth || m.viewMode == viewTable {
				showExtended := m.width >= 90
				if showExtended {
					s.WriteString(headerStyle.Render(fmt.Sprintf("  %-16s %-18s %-18s %-12s %-14s %-12s %-12s",
						"IP Address", "Hostname", "MAC", "OS", "Vendor", "DeviceType", "Ports")))
				} else {
					s.WriteString(headerStyle.Render(fmt.Sprintf("  %-16s %-25s %-20s %-15s",
						"IP Address", "Hostname", "MAC", "Ports")))
				}
				s.WriteString("\n")

				for i, dev := range devices {
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

					info := dev.ToDeviceInfo()
					osVal := info.OS
					if osVal == "" {
						osVal = "-"
					}
					vendorVal := info.Vendor
					if vendorVal == "" {
						vendorVal = "-"
					}
					devTypeVal := info.DeviceType
					if devTypeVal == "" {
						devTypeVal = "-"
					}

					var rowContent string
					if showExtended {
						rowContent = fmt.Sprintf("  %-16s %-18s %-18s %-12s %-14s %-12s %-12s",
							dev.IP, truncate(dev.Hostname, 16), dev.MAC, truncate(osVal, 10), truncate(vendorVal, 12), truncate(devTypeVal, 10), truncate(portsStr, 10))
					} else {
						rowContent = fmt.Sprintf("  %-16s %-25s %-20s %-15s",
							dev.IP, truncate(dev.Hostname, 23), dev.MAC, truncate(portsStr, 14))
					}

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
				if m.viewportReady && m.width > 0 && m.height > 0 {
					s.WriteString(m.viewport.View())
				} else {
					mapContent := RenderNetworkMap(devices, m.targetRange, m.selectedIdx, m.width)
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
	}

	km := defaultKeyMap()
	var bindings []key.Binding
	switch m.state {
	case stateInput:
		bindings = km.ShortHelpInput()
	case stateScanning:
		bindings = km.ShortHelpScanning()
	case stateResults:
		bindings = km.ShortHelpResults()
	}
	s.WriteString(m.helpModel.ShortHelpView(bindings))

	return s.String()
}
