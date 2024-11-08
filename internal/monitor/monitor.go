package monitor

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/we-be/tritium/internal/resp"
	"github.com/we-be/tritium/internal/server"
)

type Monitor struct {
	rpcNodes  []string
	respNodes map[string][]respNode
}

type respNode struct {
	name string
	addr string
}

func New() *Monitor {
	return &Monitor{
		rpcNodes: []string{
			"localhost:8080",
			"localhost:8081",
			"localhost:8082",
		},
		respNodes: map[string][]respNode{
			"localhost:6379": {
				{"Primary Node 1", "localhost:6379"},
				{"Replica Node 1.1", "localhost:6380"},
			},
			"localhost:6381": {
				{"Primary Node 2", "localhost:6381"},
				{"Replica Node 2.1", "localhost:6382"},
			},
			"localhost:6383": {
				{"Primary Node 3", "localhost:6383"},
				{"Replica Node 3.1", "localhost:6384"},
			},
		},
	}
}

func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func getRedisInfo(addr string) (map[string]string, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reader := resp.NewReader(conn)
	result, err := resp.NewCommand("INFO", "replication").ExecuteWithResponse(conn, reader)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)
	switch v := result.(type) {
	case []byte:
		lines := strings.Split(string(v), "\n")
		for _, line := range lines {
			if strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					info[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}
	return info, nil
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "just now"
	} else if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

// GatherNodeInfo collects information about all RPC nodes in the cluster
func (m *Monitor) GatherNodeInfo() (map[string]*server.NodeInfo, error) {
	var nodes map[string]*server.NodeInfo

	// Try each RPC node until we find one that responds
	for _, addr := range m.rpcNodes {
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			continue
		}
		defer client.Close()

		if err := client.Call("Store.GetClusterNodes", struct{}{}, &nodes); err == nil {
			return nodes, nil
		}
	}

	return nil, fmt.Errorf("no RPC nodes responded")
}

// GatherRespInfo collects information about all RESP storage nodes
func (m *Monitor) GatherRespInfo() map[string]map[string]string {
	info := make(map[string]map[string]string)
	// Gather all RESP nodes info
	for _, nodes := range m.respNodes {
		for _, node := range nodes {
			if nodeInfo, err := getRedisInfo(node.addr); err == nil {
				info[node.addr] = nodeInfo
			}
		}
	}
	return info
}

// printNode displays detailed information about a single node
func (m *Monitor) printNode(node *server.NodeInfo, respInfo map[string]map[string]string) {
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

	fmt.Printf("\n%s%s%s Node %s%s\n",
		Bold, stateColor, stateSymbol, node.ID, Reset)

	fmt.Printf("  %s\n", strings.Repeat("─", 50))

	fmt.Printf("  %s%sRole:%s %s%s%s\n",
		Dim, White, Reset,
		BrightCyan,
		map[bool]string{true: "Leader", false: "Follower"}[node.IsLeader],
		Reset)

	fmt.Printf("  %s%sRPC Address:%s %s%s%s\n",
		Dim, White, Reset,
		BrightYellow, node.RPCAddr, Reset)

	fmt.Printf("  %s%sRESP Store:%s %s%s%s\n",
		Dim, White, Reset,
		BrightYellow, node.RespAddr, Reset)

	fmt.Printf("  %s%sConnections:%s %s%d%s\n",
		Dim, White, Reset,
		BrightMagenta, node.Stats.ActiveConnections, Reset)

	fmt.Printf("  %s%sThroughput:%s %s%.2f MB/s%s\n",
		Dim, White, Reset,
		BrightGreen,
		float64(node.Stats.BytesTransferred)/(1024*1024),
		Reset)

	fmt.Printf("  %s%sLast Seen:%s %s%s ago%s\n",
		Dim, White, Reset,
		BrightBlue, formatDuration(time.Since(node.LastSeen)), Reset)

	if respNodes, ok := m.respNodes[node.RespAddr]; ok {
		fmt.Printf("  %s\n", strings.Repeat("─", 50))
		m.printRespNodes(respNodes, respInfo)
	}

	fmt.Printf("  %s\n", strings.Repeat("─", 50))
}

// printRespNodes displays information about RESP storage nodes
func (m *Monitor) printRespNodes(nodes []respNode, info map[string]map[string]string) {
	for _, node := range nodes {
		nodeInfo := info[node.addr]
		if nodeInfo == nil {
			fmt.Printf("  %s%s✗ %s (%s)%s\n",
				Bold, Red, node.name, node.addr, Reset)
			continue
		}

		role := nodeInfo["role"]
		stateColor := BrightGreen
		stateSymbol := "✓"

		if role == "slave" && nodeInfo["master_link_status"] != "up" {
			stateColor = BrightRed
			stateSymbol = "✗"
		}

		fmt.Printf("  %s%s%s %s%s\n",
			stateColor, stateSymbol, Reset, node.name, Reset)

		if role == "master" {
			fmt.Printf("    %s%sRole:%s Master, Replicas: %s%s%s\n",
				Dim, White, Reset,
				BrightYellow, nodeInfo["connected_slaves"], Reset)
		} else {
			fmt.Printf("    %s%sRole:%s Replica, Status: %s%s%s\n",
				Dim, White, Reset,
				map[string]string{
					"up":   BrightGreen,
					"down": BrightRed,
				}[nodeInfo["master_link_status"]],
				nodeInfo["master_link_status"],
				Reset)
		}
	}
}
