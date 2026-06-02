package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mendsec/catnet-core/pkg/events"
	"github.com/mendsec/catnet-core/pkg/profile"
	"github.com/mendsec/catnet-core/pkg/results"
	"github.com/mendsec/catnet-core/pkg/scan"
	"github.com/mendsec/catnet-core/pkg/targets"
)

type scanFinishedMsg struct{ err error }

var (
	baseStyle = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#66fcf1")).
			Bold(true).
			MarginBottom(1)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))
	hostAliveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b"))
	hostDeadStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8be9fd")).Underline(true)
)

type Model struct {
	textInput textinput.Model
	spinner   spinner.Model
	progress  progress.Model

	engine    *scan.Engine
	eventChan chan events.Event
	cancel    context.CancelFunc

	isScanning bool
	err        error
	hosts      []results.HostResult
	ratio      float64
	total      int
	processed  int

	width  int
	height int
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter IP Range (e.g. 192.168.1.1-254)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	prog := progress.New(progress.WithDefaultGradient())

	return Model{
		textInput: ti,
		spinner:   s,
		progress:  prog,
		hosts:     []results.HostResult{},
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.isScanning {
				return m.startScan(m.textInput.Value())
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10

	case events.Event:
		switch msg.Type {
		case events.ScanStarted:
			m.total = msg.Data.(int)
		case events.HostDiscovered:
			data := msg.Data.(events.HostDiscoveredData)
			m.hosts = append(m.hosts, data.Host)
		case events.ScanProgress:
			data := msg.Data.(events.ProgressData)
			m.processed = data.Processed
			m.total = data.Total
			m.ratio = data.Ratio
		}
		// Continuar escutando eventos do canal
		cmds = append(cmds, waitForEvent(m.eventChan))

	case scanFinishedMsg:
		m.isScanning = false
		m.ratio = 1.0
		if msg.err != nil {
			m.err = msg.err
		}

	case spinner.TickMsg:
		if m.isScanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if !m.isScanning {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) startScan(input string) (tea.Model, tea.Cmd) {
	ips, err := targets.ParseRange(input)
	if err != nil {
		m.err = err
		return *m, nil
	}

	m.isScanning = true
	m.err = nil
	m.hosts = []results.HostResult{}
	m.ratio = 0
	m.processed = 0
	m.total = len(ips)

	m.engine = scan.NewEngine()
	prof := profile.DefaultProfile()
	prof.Concurrency = 100

	m.eventChan = make(chan events.Event, 100)
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	// Roda o engine em background
	go func() {
		_ = m.engine.ScanStream(ctx, ips, prof, m.eventChan)
		close(m.eventChan)
		// Envia mensagem final manual já que o close encerra a escuta
		time.Sleep(100 * time.Millisecond) // Pequeno atraso para a UI ler os ultimos
	}()

	return *m, tea.Batch(
		waitForEvent(m.eventChan),
		m.spinner.Tick,
	)
}

func waitForEvent(ch <-chan events.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return scanFinishedMsg{}
		}
		return ev
	}
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("CATNET TERMINAL UI"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	if !m.isScanning {
		b.WriteString(m.textInput.View())
		b.WriteString("\n(Press Enter to start, Esc to quit)\n")
	} else {
		b.WriteString(fmt.Sprintf("%s Scanning %d targets...\n", m.spinner.View(), m.total))
		b.WriteString(m.progress.ViewAs(m.ratio))
		b.WriteString(fmt.Sprintf(" %d/%d\n", m.processed, m.total))
	}

	b.WriteString("\n")
	b.WriteString(tableHeaderStyle.Render(fmt.Sprintf("%-16s %-30s %-8s %-20s", "IP", "HOSTNAME", "STATUS", "PORTS")))
	b.WriteString("\n")

	for _, h := range m.hosts {
		status := hostDeadStyle.Render("Dead")
		if h.Alive {
			status = hostAliveStyle.Render("Alive")
		}
		ports := fmt.Sprintf("%v", h.OpenPorts)
		if len(h.OpenPorts) == 0 {
			ports = "-"
		}
		
		row := fmt.Sprintf("%-16s %-30s %-8s %-20s\n", h.IP, h.Hostname, status, ports)
		b.WriteString(row)
	}

	if m.isScanning && len(m.hosts) == 0 {
		b.WriteString("\nWaiting for discoveries...\n")
	}

	return baseStyle.Render(b.String())
}
