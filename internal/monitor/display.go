package monitor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/we-be/tritium/internal/server"
)

func (m *Monitor) PrintHeader() {
	header := `
████████╗██████╗ ██╗████████╗██╗██╗   ██╗███╗   ███╗
╚══██╔══╝██╔══██╗██║╚══██╔══╝██║██║   ██║████╗ ████║
   ██║   ██████╔╝██║   ██║   ██║██║   ██║██╔████╔██║
   ██║   ██╔══██╗██║   ██║   ██║██║   ██║██║╚██╔╝██║
   ██║   ██║  ██║██║   ██║   ██║╚██████╔╝██║ ╚═╝ ██║
   ╚═╝   ╚═╝  ╚═╝╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝     ╚═╝
`
	fmt.Printf("%s%s%s", BrightMagenta, header, Reset)
	fmt.Printf("%s%s%sCluster Monitor%s\n",
		BgBlue, BrightWhite, Bold, Reset)
}

func (m *Monitor) PrintSummary(nodes map[string]*server.NodeInfo, respInfo map[string]map[string]string) {
	fmt.Printf("\n%s%s%sCluster Summary%s\n",
		BgBlue, BrightWhite, Bold, Reset)

	fmt.Printf("\n  %s\n", strings.Repeat("─", 50))

	var nodesList []*server.NodeInfo
	for _, node := range nodes {
		nodesList = append(nodesList, node)
	}
	sort.Slice(nodesList, func(i, j int) bool {
		return nodesList[i].RPCAddr < nodesList[j].RPCAddr
	})

	for _, node := range nodesList {
		var stateColor, stateSymbol string
		switch node.State {
		case server.NodeStateHealthy:
			stateColor = BrightGreen
			stateSymbol = "✓"
		case server.NodeStateDegraded:
			stateColor = BrightYellow
			stateSymbol = "!"
		default:
			stateColor = BrightRed
			stateSymbol = "✗"
		}

		respNodes := m.respNodes[node.RespAddr]
		allHealthy := true
		for _, rn := range respNodes {
			info := respInfo[rn.addr]
			if info == nil || (info["role"] == "slave" && info["master_link_status"] != "up") {
				allHealthy = false
				break
			}
		}

		storeStatus := "✓"
		if !allHealthy {
			storeStatus = "!"
		}

		fmt.Printf("  %s%s%s Node %s [%s%s%s Store] %s%s\n",
			stateColor, stateSymbol, Reset,
			node.RPCAddr,
			map[bool]string{true: BrightGreen, false: BrightYellow}[allHealthy],
			storeStatus,
			Reset,
			map[bool]string{true: "Leader", false: "Follower"}[node.IsLeader],
			Reset)
	}

	fmt.Printf("  %s\n", strings.Repeat("─", 50))
}

func (m *Monitor) PrintDetailed(nodes map[string]*server.NodeInfo, respInfo map[string]map[string]string) {
	fmt.Printf("\n%s%s%sCluster Status%s\n",
		BgBlue, BrightWhite, Bold, Reset)

	var nodesList []*server.NodeInfo
	for _, node := range nodes {
		nodesList = append(nodesList, node)
	}
	sort.Slice(nodesList, func(i, j int) bool {
		return nodesList[i].RPCAddr < nodesList[j].RPCAddr
	})

	for _, node := range nodesList {
		m.printNode(node, respInfo)
	}
}
