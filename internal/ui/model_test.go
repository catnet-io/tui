package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/catnet-io/engine/pkg/events"
	"github.com/catnet-io/engine/pkg/results"
	"github.com/catnet-io/engine/pkg/scan"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	m := InitialModel()
	if m.state != stateInput {
		t.Errorf("expected initial state stateInput, got %v", m.state)
	}
	if cmd := m.Init(); cmd == nil {
		t.Error("expected non-nil Init command")
	}
	viewStr := m.View()
	if !strings.Contains(viewStr, "CATNET TERMINAL SCANNER") {
		t.Errorf("expected View to contain title banner, got %q", viewStr)
	}
}

func TestModelUpdateNavigationAndEvents(t *testing.T) {
	m := InitialModel()

	// 1. Enter key in stateInput transitions state to stateScanning
	m.textInput.SetValue("127.0.0.1")
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(Model)
	if m.state != stateScanning {
		t.Fatalf("expected stateScanning, got %v", m.state)
	}
	if cmd == nil {
		t.Error("expected non-nil command when scan starts")
	}

	// 2. Receive HostDiscoveredMsg
	host := results.HostResult{
		IP:        "127.0.0.1",
		Hostname:  "localhost",
		MAC:       "00:11:22:33:44:55",
		OpenPorts: []int{80, 443},
	}
	updatedModel, _ = m.Update(hostDiscoveredMsg(host))
	m = updatedModel.(Model)
	if len(m.devices) != 1 || m.devices[0].IP != "127.0.0.1" {
		t.Errorf("expected 1 discovered host, got %v", m.devices)
	}

	// 3. Receive ScanProgressMsg
	updatedModel, _ = m.Update(scanProgressMsg(0.5))
	m = updatedModel.(Model)
	if m.progress != 0.5 {
		t.Errorf("expected progress 0.5, got %f", m.progress)
	}

	// 4. Receive ScanDoneMsg
	updatedModel, _ = m.Update(scanDoneMsg{})
	m = updatedModel.(Model)
	if m.state != stateResults {
		t.Fatalf("expected stateResults, got %v", m.state)
	}

	// 5. Test results view and table selection (Down/Up)
	m.devices = append(m.devices, results.HostResult{IP: "192.168.1.2", Hostname: "host2"})
	if m.selectedIdx != 0 {
		t.Errorf("expected initial selectedIdx 0, got %d", m.selectedIdx)
	}

	// Move down
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1 after Down, got %d", m.selectedIdx)
	}

	// Move up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 after Up, got %d", m.selectedIdx)
	}

	// 6. Test Export ('e')
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updatedModel.(Model)
	defer os.Remove("catnet_export.json")
	if _, err := os.Stat("catnet_export.json"); os.IsNotExist(err) {
		t.Error("expected catnet_export.json file to be created")
	}

	// 7. Test Reset ('q') back to stateInput
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updatedModel.(Model)
	if m.state != stateInput {
		t.Errorf("expected stateInput after 'q' in results state, got %v", m.state)
	}
}

func TestScanCancellation(t *testing.T) {
	engine := scan.NewEngine()
	m := NewModel(engine)
	m.state = stateScanning

	canceled := false
	m.cancelFn = func() {
		canceled = true
	}

	// Press Esc while scanning
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !canceled {
		t.Error("expected cancelFn to be invoked when Esc is pressed in scanning state")
	}
	if cmd() != tea.Quit() {
		t.Error("expected tea.Quit command on Esc press")
	}
}

func TestListenForEvents(t *testing.T) {
	ch := make(chan events.Event, 3)

	ch <- events.Event{
		Type: events.HostDiscovered,
		Data: events.HostDiscoveredData{
			Host: results.HostResult{IP: "10.0.0.1"},
		},
	}
	ch <- events.Event{
		Type: events.ScanProgress,
		Data: events.ProgressData{Ratio: 0.75},
	}
	ch <- events.Event{
		Type: events.ScanCompleted,
	}

	cmd := listenForEvents(ch)

	msg1 := cmd()
	if hostMsg, ok := msg1.(hostDiscoveredMsg); !ok || hostMsg.IP != "10.0.0.1" {
		t.Errorf("unexpected msg1: %#v", msg1)
	}

	msg2 := listenForEvents(ch)()
	if progMsg, ok := msg2.(scanProgressMsg); !ok || float64(progMsg) != 0.75 {
		t.Errorf("unexpected msg2: %#v", msg2)
	}

	msg3 := listenForEvents(ch)()
	if _, ok := msg3.(scanDoneMsg); !ok {
		t.Errorf("unexpected msg3: %#v", msg3)
	}

	close(ch)
	msg4 := listenForEvents(ch)()
	if _, ok := msg4.(scanDoneMsg); !ok {
		t.Errorf("unexpected msg4 when channel closed: %#v", msg4)
	}
}

func TestTruncate(t *testing.T) {
	if res := truncate("hello world", 5); res != "he..." {
		t.Errorf("expected 'he...', got %q", res)
	}
	if res := truncate("hello", 10); res != "hello" {
		t.Errorf("expected 'hello', got %q", res)
	}
	if res := truncate("abc", 2); res != "ab" {
		t.Errorf("expected 'ab', got %q", res)
	}
}

func TestScanCancellationWithoutLeak(t *testing.T) {
	engine := scan.NewEngine()
	m := NewModel(engine)
	m.targetRange = "127.0.0.1"
	m.state = stateScanning

	cmd := m.startScan()
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd from startScan")
	}

	// Cancel scan immediately
	if m.cancelFn != nil {
		m.cancelFn()
	}

	// Give background goroutine time to complete defer cancel() and channel close
	time.Sleep(100 * time.Millisecond)

	// Draining eventChan should yield scanDoneMsg cleanly without blocking
	msg := listenForEvents(m.eventChan)()
	if _, ok := msg.(scanDoneMsg); !ok {
		t.Errorf("expected scanDoneMsg upon scan cancellation, got %#v", msg)
	}
}
