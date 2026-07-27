package ui

import (
	"strings"
	"testing"

	"github.com/catnet-io/engine/pkg/results"
	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderNetworkMap(t *testing.T) {
	t.Run("Empty Devices", func(t *testing.T) {
		out := RenderNetworkMap(nil, "auto", 0)
		if out != "" {
			t.Errorf("expected empty string for zero devices, got %q", out)
		}
	})

	t.Run("Single Device", func(t *testing.T) {
		devs := []results.HostResult{
			{IP: "192.168.1.1", Hostname: "router.local", Alive: true},
		}
		out := RenderNetworkMap(devs, "192.168.1.0/24", 0)
		if !strings.Contains(out, "NETWORK TOPOLOGY MAP") {
			t.Errorf("expected section header in output, got: %s", out)
		}
		if !strings.Contains(out, "192.168.1.1") {
			t.Errorf("expected IP in output, got: %s", out)
		}
	})

	t.Run("Multiple Devices Role & Link Rendering", func(t *testing.T) {
		devs := []results.HostResult{
			{IP: "192.168.1.1", Hostname: "router.home", Alive: true},
			{IP: "192.168.1.15", Hostname: "workstation-pc", OpenPorts: []int{22}, Alive: true},
			{IP: "192.168.1.42", Hostname: "hp-printer", OpenPorts: []int{9100}, Alive: true},
			{IP: "192.168.1.105", Hostname: "android-phone", Alive: true},
		}
		out := RenderNetworkMap(devs, "auto", 1)
		if !strings.Contains(out, "192.168.1.1") {
			t.Errorf("expected gateway IP in output, got: %s", out)
		}
		if !strings.Contains(out, "192.168.1.15") {
			t.Errorf("expected host IP in output, got: %s", out)
		}
	})

	t.Run("No Distance Or Hops Words In UI", func(t *testing.T) {
		devs := []results.HostResult{
			{IP: "10.0.0.1", Hostname: "gateway", Alive: true},
			{IP: "10.0.0.2", Hostname: "host", Alive: true},
		}
		out := RenderNetworkMap(devs, "10.0.0.0/24", 0)
		lowerOut := strings.ToLower(out)
		if strings.Contains(lowerOut, "hop") || strings.Contains(lowerOut, "distância") || strings.Contains(lowerOut, "distance") {
			t.Errorf("found forbidden hop/distance wording in output: %s", out)
		}
	})
}

func TestViewModeToggleKeyT(t *testing.T) {
	m := InitialModel()
	m.state = stateResults
	m.devices = []results.HostResult{{IP: "192.168.1.1", Alive: true}}

	if m.viewMode != viewBoth {
		t.Errorf("expected initial viewMode viewBoth, got %v", m.viewMode)
	}

	// Pressing 't' toggles view mode
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updatedModel.(Model)
	if m.viewMode != viewTable {
		t.Errorf("expected viewMode viewTable after 't', got %v", m.viewMode)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updatedModel.(Model)
	if m.viewMode != viewMap {
		t.Errorf("expected viewMode viewMap after second 't', got %v", m.viewMode)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updatedModel.(Model)
	if m.viewMode != viewBoth {
		t.Errorf("expected viewMode viewBoth after third 't', got %v", m.viewMode)
	}
}

func TestWindowSizeMsgHandling(t *testing.T) {
	m := InitialModel()
	m.state = stateResults
	m.devices = []results.HostResult{{IP: "192.168.1.1", Alive: true}}

	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updatedModel.(Model)

	if m.width != 100 || m.height != 40 {
		t.Errorf("expected model dimensions (100, 40), got (%d, %d)", m.width, m.height)
	}
	if !m.viewportReady {
		t.Error("expected viewport to be initialized on WindowSizeMsg")
	}

	viewStr := m.View()
	if !strings.Contains(viewStr, "NETWORK TOPOLOGY MAP") {
		t.Errorf("expected View output to render topology map within viewport, got: %s", viewStr)
	}
}

func TestSelectionConsistencyWithSortedNodes(t *testing.T) {
	// devs[0] is regular host 192.168.1.50, devs[1] is router/gateway 192.168.1.1
	devs := []results.HostResult{
		{IP: "192.168.1.50", Hostname: "workstation", Alive: true},
		{IP: "192.168.1.1", Hostname: "router.gateway", Alive: true},
	}

	// Select selectedIdx = 0 (192.168.1.50)
	out := RenderNetworkMap(devs, "192.168.1.0/24", 0)

	// Verify that workstation (192.168.1.50) is rendered as selected
	lines := strings.Split(out, "\n")
	var hostLine, gatewayLine string
	for _, line := range lines {
		if strings.Contains(line, "192.168.1.50") {
			hostLine = line
		}
		if strings.Contains(line, "192.168.1.1") {
			gatewayLine = line
		}
	}

	if hostLine == "" || gatewayLine == "" {
		t.Fatalf("expected both host and gateway lines in rendered output: %s", out)
	}

	// Gateway should come first in sorted nodes list
	gatewayIdx := strings.Index(out, "192.168.1.1")
	hostIdx := strings.Index(out, "192.168.1.50")
	if gatewayIdx >= hostIdx {
		t.Errorf("expected Gateway (192.168.1.1) to be sorted before Host (192.168.1.50)")
	}

	// Highlight should be on 192.168.1.50 (devs[0])
	selectedStyleSnippet := selectedRowStyle.Render("192.168.1.50")
	if !strings.Contains(out, selectedStyleSnippet) {
		t.Errorf("expected selected host 192.168.1.50 to be rendered with selectedRowStyle")
	}
}
