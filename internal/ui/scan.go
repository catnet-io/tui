package ui

import (
	"context"

	"github.com/catnet-io/engine/pkg/events"
	"github.com/catnet-io/engine/pkg/profile"
	"github.com/catnet-io/engine/pkg/results"
	tea "github.com/charmbracelet/bubbletea"
)

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
			if data, ok := ev.Data.(events.HostDiscoveredData); ok {
				return hostDiscoveredMsg(data.Host)
			}
		case events.ScanProgress:
			if data, ok := ev.Data.(events.ProgressData); ok {
				return scanProgressMsg(data.Ratio)
			}
		case events.ScanCompleted:
			return scanDoneMsg{}
		}
		return listenForEvents(ch)()
	}
}

func (m *Model) startScan() tea.Cmd {
	m.eventChan = make(chan events.Event)
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFn = cancel

	go func() {
		defer cancel()
		cfg := profile.DefaultProfile()
		cfg.Concurrency = 32
		cfg.TimeoutMs = 1000
		if err := m.engine.ScanStream(ctx, []string{m.targetRange}, cfg, m.eventChan); err != nil && ctx.Err() == nil {
			_ = err
		}
		close(m.eventChan)
	}()

	return listenForEvents(m.eventChan)
}
