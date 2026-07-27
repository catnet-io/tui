package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/catnet-io/engine/pkg/results"
	"github.com/catnet-io/engine/pkg/topology"
)

// RenderNetworkMap builds an ASCII/Unicode visual tree graph representing network topology
// by delegating graph construction and analysis to catnet-io/engine/pkg/topology.
func RenderNetworkMap(devices []results.HostResult, targetRange string, selectedIdx int, width ...int) string {
	if len(devices) == 0 {
		return ""
	}

	devInfos := make([]results.DeviceInfo, len(devices))
	for i, dev := range devices {
		devInfos[i] = dev.ToDeviceInfo()
	}

	report := &results.ScanReport{
		Devices: devInfos,
	}

	graph := topology.BuildGraph(report)
	if graph == nil || len(graph.Nodes) == 0 {
		return ""
	}

	w := 70
	if len(width) > 0 && width[0] > 0 {
		w = width[0]
	}

	var b strings.Builder
	headerBar := "── NETWORK TOPOLOGY MAP "
	if w > len(headerBar) {
		headerBar += strings.Repeat("─", w-len(headerBar))
	}
	b.WriteString(cyanText.Render(headerBar))
	b.WriteString("\n\n")

	subnet := targetRange
	if subnet == "" || subnet == "auto" {
		if graph.Gateway != "" {
			subnet = fmt.Sprintf("Subnet (Gateway: %s)", graph.Gateway)
		} else {
			subnet = "Local Subnet"
		}
	}
	b.WriteString(greyText.Render(fmt.Sprintf("📍 Network Target / Subnet: %s", subnet)))
	b.WriteString("\n")

	edgeMap := make(map[string]topology.TopologyEdge)
	for _, edge := range graph.Edges {
		edgeMap[edge.Target] = edge
	}

	nodes := make([]topology.TopologyNode, len(graph.Nodes))
	copy(nodes, graph.Nodes)

	sort.Slice(nodes, func(i, j int) bool {
		isIGateway := nodes[i].Role == topology.RoleGateway || nodes[i].ID == graph.Gateway
		isJGateway := nodes[j].Role == topology.RoleGateway || nodes[j].ID == graph.Gateway
		if isIGateway != isJGateway {
			return isIGateway
		}
		return nodes[i].ID < nodes[j].ID
	})

	for i, node := range nodes {
		branch := " ├── "
		if i == len(nodes)-1 {
			branch = " └── "
		}

		icon, roleLabel := getRoleIconAndLabel(node)
		labelStr := node.Label
		if labelStr == "" {
			labelStr = "unknown"
		}

		linkType := "Gateway"
		if edge, ok := edgeMap[node.ID]; ok {
			if edge.Weight < 1.0 {
				linkType = "Shared service"
			}
		} else if node.Role != topology.RoleGateway && node.ID != graph.Gateway {
			linkType = "Subnet"
		}

		linkInfo := greyText.Render(fmt.Sprintf("(Role: %-10s • Link: %s)", roleLabel, linkType))
		hostInfo := fmt.Sprintf("%s %-15s [%s] %s",
			icon, node.ID, truncate(labelStr, 20), linkInfo)

		if i == selectedIdx {
			b.WriteString(greyText.Render(branch) + selectedRowStyle.Render(hostInfo) + "\n")
		} else {
			b.WriteString(greyText.Render(branch) + normalRowStyle.Render(hostInfo) + "\n")
		}
	}

	return b.String()
}

func getRoleIconAndLabel(node topology.TopologyNode) (string, string) {
	icon := "🖥️"
	roleLabel := "Host"

	switch node.Role {
	case topology.RoleGateway:
		icon = "🌐"
		roleLabel = "Gateway"
	case topology.RoleServer:
		icon = "💻"
		roleLabel = "Server"
	case topology.RoleHost:
		icon = "🖥️"
		roleLabel = "Host"
	case topology.RoleUnknown:
		icon = "❓"
		roleLabel = "Unknown"
	}

	devType := strings.ToLower(node.DeviceType)
	switch {
	case strings.Contains(devType, "printer"):
		icon = "🖨️"
	case strings.Contains(devType, "mobile") || strings.Contains(devType, "phone"):
		icon = "📱"
	case strings.Contains(devType, "router") || strings.Contains(devType, "gateway"):
		icon = "🌐"
	case strings.Contains(devType, "server"):
		icon = "💻"
	}

	return icon, roleLabel
}
