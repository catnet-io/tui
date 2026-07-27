package ui

import (
	"context"

	"github.com/catnet-io/engine/pkg/events"
	"github.com/catnet-io/engine/pkg/profile"
	"github.com/catnet-io/engine/pkg/results"
	"github.com/catnet-io/engine/pkg/targets"
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
			if err, ok := ev.Data.(error); ok && err != nil {
				return scanDoneMsg{err: err}
			}
			return scanDoneMsg{}
		}
		return listenForEvents(ch)()
	}
}

func (m *Model) startScan() tea.Cmd {
	m.eventChan = make(chan events.Event)
	eventChan := m.eventChan
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFn = cancel
	targetRange := m.targetRange
	eng := m.engine

	go func() {
		defer cancel()
		cfg := profile.DefaultProfile()
		cfg.Concurrency = 32
		cfg.TimeoutMs = 1000

		targetList, err := targets.ParseRange(targetRange)
		if err != nil || len(targetList) == 0 {
			targetList = []string{targetRange}
		}

		var scanErr error
		if eng != nil {
			if err := eng.ScanStream(ctx, targetList, cfg, eventChan); err != nil && ctx.Err() == nil {
				scanErr = err
			}
		}
		if scanErr != nil {
			eventChan <- events.Event{
				Type: events.ScanCompleted,
				Data: scanErr,
			}
		}
		close(eventChan)
	}()

	return listenForEvents(eventChan)
}

