package ui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/catnet-io/engine/pkg/events"
	"github.com/catnet-io/engine/pkg/results"
	"github.com/catnet-io/engine/pkg/scan"
	"github.com/catnet-io/engine/pkg/targets"
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
		Alive:     true,
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
	if info, err := os.Stat("catnet_export.json"); os.IsNotExist(err) {
		t.Error("expected catnet_export.json file to be created")
	} else if err == nil {
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("expected 0600 file permissions for export file, got %o", perm)
		}
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

	// Press Esc while scanning aborts scan to stateInput
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !canceled {
		t.Error("expected cancelFn to be invoked when Esc is pressed in scanning state")
	}
	m = updatedModel.(Model)
	if m.state != stateInput {
		t.Errorf("expected transition to stateInput on Esc during scan, got %v", m.state)
	}
	if cmd != nil {
		t.Error("expected nil cmd (no quit) on Esc press during scanning")
	}

	// Press Ctrl+C while scanning cancels scan and quits
	m.state = stateScanning
	canceled = false
	m.cancelFn = func() {
		canceled = true
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !canceled {
		t.Error("expected cancelFn to be invoked when Ctrl+C is pressed in scanning state")
	}
	if cmd == nil || cmd() != tea.Quit() {
		t.Error("expected tea.Quit command on Ctrl+C press")
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

func TestStartScanEventPropagation(t *testing.T) {
	engine := scan.NewEngine()
	m := NewModel(engine)
	m.targetRange = "127.0.0.1"
	m.state = stateScanning

	cmd := m.startScan()
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd from startScan")
	}

	msg := listenForEvents(m.eventChan)()
	if msg == nil {
		t.Error("expected non-nil message from scan event listener")
	}
}

func TestDefaultAutoTarget(t *testing.T) {
	m := InitialModel()
	if m.textInput.Value() != "" {
		t.Errorf("expected empty initial input value, got %q", m.textInput.Value())
	}
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(Model)
	if m.targetRange != "auto" {
		t.Errorf("expected targetRange to default to 'auto', got %q", m.targetRange)
	}

	m2 := InitialModel()
	m2.textInput.SetValue("   ")
	updatedModel2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 = updatedModel2.(Model)
	if m2.targetRange != "auto" {
		t.Errorf("expected whitespace-only input to default to 'auto', got %q", m2.targetRange)
	}

	m3 := InitialModel()
	m3.textInput.SetValue("  192.168.1.0/24  ")
	updatedModel3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 = updatedModel3.(Model)
	if m3.targetRange != "192.168.1.0/24" {
		t.Errorf("expected trimmed targetRange '192.168.1.0/24', got %q", m3.targetRange)
	}
}

func TestTargetRangeResolution(t *testing.T) {
	parsedAuto, err := targets.ParseRange("auto")
	if err != nil || len(parsedAuto) == 0 {
		t.Errorf("expected targets.ParseRange('auto') to resolve IPs, got err: %v, count: %d", err, len(parsedAuto))
	}

	parsedCIDR, err := targets.ParseRange("192.168.1.0/28")
	if err != nil || len(parsedCIDR) != 14 {
		t.Errorf("expected targets.ParseRange('192.168.1.0/28') to resolve 14 IPs, got err: %v, count: %d", err, len(parsedCIDR))
	}
}

func TestInputQHandling(t *testing.T) {
	m := InitialModel()
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updatedModel.(Model)
	if m.state != stateInput {
		t.Errorf("expected to stay in stateInput when typing 'q', got %v", m.state)
	}
	if cmd != nil && cmd() == tea.Quit() {
		t.Error("typing 'q' in stateInput should not quit application")
	}
}

func TestExportFeedbackRendering(t *testing.T) {
	m := InitialModel()
	m.state = stateResults
	m.devices = []results.HostResult{{IP: "192.168.1.1"}}

	// Success export log
	m.logMsgs = []string{"Exported to catnet_export.json"}
	viewStr := m.View()
	if !strings.Contains(viewStr, "Exported to catnet_export.json") {
		t.Errorf("expected View to render export success log, got %q", viewStr)
	}

	// Error export log
	m.logMsgs = []string{"Export error: permission denied"}
	viewStr = m.View()
	if !strings.Contains(viewStr, "Export error: permission denied") {
		t.Errorf("expected View to render export error log, got %q", viewStr)
	}
}

func TestAutoPlaceholderAndView(t *testing.T) {
	engine := scan.NewEngine()
	m := NewModel(engine)
	if !strings.Contains(m.textInput.Placeholder, "auto") {
		t.Errorf("expected textInput placeholder to contain 'auto', got %q", m.textInput.Placeholder)
	}
	viewStr := m.View()
	if !strings.Contains(viewStr, "(default: auto)") {
		t.Errorf("expected stateInput View to contain prompt with default 'auto', got %q", viewStr)
	}
}

func TestScanAbortWithQAndEsc(t *testing.T) {
	for _, key := range []string{"q", "esc"} {
		t.Run("abort with "+key, func(t *testing.T) {
			engine := scan.NewEngine()
			m := NewModel(engine)
			m.targetRange = "127.0.0.1"
			m.state = stateScanning
			m.startScan()

			if m.cancelFn == nil {
				t.Fatal("expected non-nil cancelFn during scan")
			}

			updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if key == "esc" {
				updatedModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			}
			m = updatedModel.(Model)

			if m.state != stateInput {
				t.Errorf("expected state to return to stateInput after abort, got %v", m.state)
			}
			if m.cancelFn != nil {
				t.Error("expected cancelFn to be nil after abort")
			}
			if cmd != nil && cmd() == tea.Quit() {
				t.Errorf("key %q during scan should abort to stateInput, not quit app", key)
			}

			// Verify late scanDoneMsg does not change state back to stateResults
			updatedModel, _ = m.Update(scanDoneMsg{})
			m = updatedModel.(Model)
			if m.state != stateInput {
				t.Errorf("expected state to remain stateInput on late scanDoneMsg, got %v", m.state)
			}

			// Drain eventChan so goroutine terminates cleanly
			for range m.eventChan {
			}
		})
	}
}

func TestScanErrorPropagation(t *testing.T) {
	m := InitialModel()
	m.state = stateScanning

	err := errors.New("network interface down")

	updatedModel, _ := m.Update(scanDoneMsg{err: err})
	m = updatedModel.(Model)

	if m.state != stateResults {
		t.Fatalf("expected transition to stateResults, got %v", m.state)
	}
	if m.errorMsg != "network interface down" {
		t.Errorf("expected errorMsg 'network interface down', got %q", m.errorMsg)
	}

	viewStr := m.View()
	if !strings.Contains(viewStr, "Scan Error: network interface down") {
		t.Errorf("expected View to contain error banner, got: %s", viewStr)
	}
}

func TestViewportMessageDelegation(t *testing.T) {
	m := InitialModel()
	m.state = stateResults
	m.devices = []results.HostResult{{IP: "192.168.1.1", Alive: true}}

	// Initialize window size & viewport
	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updatedModel.(Model)

	if !m.viewportReady {
		t.Fatal("expected viewportReady to be true after WindowSizeMsg")
	}

	// Send PgDn key to model in stateResults to test delegation to viewport
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updatedModel.(Model)

	// Ensure view renders cleanly after viewport update
	viewStr := m.View()
	if !strings.Contains(viewStr, "NETWORK TOPOLOGY MAP") {
		t.Errorf("expected View to render topology map after viewport scrolling message, got: %s", viewStr)
	}
}

func TestResultColumnsOSVendorDeviceType(t *testing.T) {
	m := InitialModel()
	m.state = stateResults
	m.devices = []results.HostResult{
		{
			IP:        "192.168.1.50",
			Hostname:  "printer.local",
			MAC:       "AA:BB:CC:DD:EE:FF",
			Alive:     true,
			OpenPorts: []int{9100},
		},
	}

	// Wide terminal: OS, Vendor, DeviceType headers should be displayed
	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = updatedModel.(Model)
	viewStr := m.View()
	if !strings.Contains(viewStr, "OS") || !strings.Contains(viewStr, "Vendor") || !strings.Contains(viewStr, "DeviceType") {
		t.Errorf("expected wide view to render OS, Vendor, DeviceType headers, got:\n%s", viewStr)
	}

	// Narrow terminal: OS, Vendor, DeviceType headers should be hidden
	updatedModel, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = updatedModel.(Model)
	viewStr = m.View()
	if strings.Contains(viewStr, "DeviceType") {
		t.Errorf("expected narrow view to hide DeviceType column, got:\n%s", viewStr)
	}
}

func TestTextFilter(t *testing.T) {
	m := InitialModel()
	m.state = stateResults
	m.devices = []results.HostResult{
		{IP: "192.168.1.1", Hostname: "router.home", MAC: "00:11:22:33:44:55", Alive: true},
		{IP: "192.168.1.20", Hostname: "desktop-pc", MAC: "AA:BB:CC:11:22:33", Alive: true},
		{IP: "10.0.0.5", Hostname: "server-node", MAC: "99:88:77:66:55:44", Alive: true},
	}

	// Activate filter with '/'
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updatedModel.(Model)
	if !m.isFiltering {
		t.Fatal("expected isFiltering to be true after '/' keypress")
	}

	// Type 'router' to filter
	for _, r := range "router" {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updatedModel.(Model)
	}

	filtered := m.filteredDevices()
	if len(filtered) != 1 || filtered[0].IP != "192.168.1.1" {
		t.Errorf("expected 1 host matching 'router', got %v", filtered)
	}

	// Case-insensitive check: clear and filter by 'DESKTOP'
	m.filterInput.SetValue("DESKTOP")
	filtered = m.filteredDevices()
	if len(filtered) != 1 || filtered[0].IP != "192.168.1.20" {
		t.Errorf("expected 1 host matching 'DESKTOP', got %v", filtered)
	}

	// Filter by MAC substring
	m.filterInput.SetValue("99:88")
	filtered = m.filteredDevices()
	if len(filtered) != 1 || filtered[0].IP != "10.0.0.5" {
		t.Errorf("expected 1 host matching MAC '99:88', got %v", filtered)
	}

	// Filter with no matches
	m.filterInput.SetValue("nonexistent")
	filtered = m.filteredDevices()
	if len(filtered) != 0 {
		t.Errorf("expected 0 hosts matching 'nonexistent', got %d", len(filtered))
	}
	viewStr := m.View()
	if !strings.Contains(viewStr, "No active hosts matching filter 'nonexistent'.") {
		t.Errorf("expected empty filter message in View, got:\n%s", viewStr)
	}

	// Esc clears filter
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(Model)
	if m.isFiltering {
		t.Error("expected isFiltering to be false after Esc press")
	}
	if len(m.filteredDevices()) != 3 {
		t.Errorf("expected full list (3 hosts) after clearing filter, got %d", len(m.filteredDevices()))
	}
}

func TestHelpOverlayToggle(t *testing.T) {
	m := InitialModel()
	m.state = stateResults

	// Press '?' to open help modal
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updatedModel.(Model)
	if !m.showHelp {
		t.Fatal("expected showHelp to be true after '?' keypress")
	}
	viewStr := m.View()
	if !strings.Contains(viewStr, "KEYBOARD SHORTCUTS HELP") {
		t.Errorf("expected View to contain help overlay header, got:\n%s", viewStr)
	}

	// Press 'esc' to close help modal
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(Model)
	if m.showHelp {
		t.Fatal("expected showHelp to be false after 'esc' keypress")
	}
}
