package main

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

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Regular colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Backgrounds
	BgBlue = "\033[44m"
)

type Monitor struct {
	rpcNodes        []string
	nodeRespMapping map[string][]struct {
		name string
		addr string
	}
}

func NewMonitor() *Monitor {
	m := &Monitor{
		rpcNodes: []string{
			"localhost:8080",
			"localhost:8081",
			"localhost:8082",
		},
		nodeRespMapping: map[string][]struct {
			name string
			addr string
		}{
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
	return m
}

func clearScreen() {
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

func (m *Monitor) printHeader() {
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

func (m *Monitor) printNodeStatus(node *server.NodeInfo, respInfo map[string]map[string]string) {
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

	// Node header with status
	fmt.Printf("\n%s%s%s Node %s%s\n",
		Bold, stateColor, stateSymbol, node.ID, Reset)

	// Node details section
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

	lastSeenDur := time.Since(node.LastSeen)
	fmt.Printf("  %s%sLast Seen:%s %s%s ago%s\n",
		Dim, White, Reset,
		BrightBlue, formatDuration(lastSeenDur), Reset)

	// Print associated RESP nodes if any
	if respNodes, ok := m.nodeRespMapping[node.RespAddr]; ok {
		fmt.Printf("  %s\n", strings.Repeat("─", 50))
		m.printRESPNodes(respNodes, respInfo)
	}

	fmt.Printf("  %s\n", strings.Repeat("─", 50))
}

func (m *Monitor) printRESPNodes(nodes []struct{ name, addr string }, info map[string]map[string]string) {
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

		if role == "slave" {
			if nodeInfo["master_link_status"] != "up" {
				stateColor = BrightRed
				stateSymbol = "✗"
			}
		}

		fmt.Printf("  %s%s%s %s%s\n",
			stateColor, stateSymbol, Reset, node.name,
			Reset)

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

func (m *Monitor) gatherNodeInfo() (map[string]*server.NodeInfo, error) {
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

func (m *Monitor) gatherRESPInfo() map[string]map[string]string {
	info := make(map[string]map[string]string)
	for _, nodes := range m.nodeRespMapping {
		for _, node := range nodes {
			if nodeInfo, err := getRedisInfo(node.addr); err == nil {
				info[node.addr] = nodeInfo
			}
		}
	}
	return info
}

func main() {
	monitor := NewMonitor()

	for {
		clearScreen()
		monitor.printHeader()

		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("\n%s%s%sStatus at %s%s\n",
			Dim, Italic, White, timestamp, Reset)

		nodes, err := monitor.gatherNodeInfo()
		respInfo := monitor.gatherRESPInfo()

		if err != nil {
			fmt.Printf("\n%s%s✗ No RPC nodes found%s\n", Bold, Red, Reset)
		} else {
			fmt.Printf("\n%s%s%sCluster Status%s\n",
				BgBlue, BrightWhite, Bold, Reset)

			for _, node := range nodes {
				monitor.printNodeStatus(node, respInfo)
			}
		}

		fmt.Printf("\n%s%s%sPress Ctrl+C to exit • Refreshing every 2s%s\n",
			Dim, Italic, White, Reset)

		time.Sleep(2 * time.Second)
	}
}
