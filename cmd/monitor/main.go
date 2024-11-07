package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/we-be/tritium/internal/resp"
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
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

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

func printHeader() {
	header := `
████████╗██████╗ ██╗████████╗██╗██╗   ██╗███╗   ███╗
╚══██╔══╝██╔══██╗██║╚══██╔══╝██║██║   ██║████╗ ████║
   ██║   ██████╔╝██║   ██║   ██║██║   ██║██╔████╔██║
   ██║   ██╔══██╗██║   ██║   ██║██║   ██║██║╚██╔╝██║
   ██║   ██║  ██║██║   ██║   ██║╚██████╔╝██║ ╚═╝ ██║
   ╚═╝   ╚═╝  ╚═╝╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝     ╚═╝
`
	fmt.Printf("%s%s%s", BrightCyan, header, Reset)
	fmt.Printf("%s%s%sReplication Monitor%s\n",
		BgBlue, BrightWhite, Bold, Reset)
}

func printInstanceStatus(instance string, addr string, info map[string]string, err error) {
	if err != nil {
		fmt.Printf("\n%s%s%s %s (%s)%s\n",
			Bold, Red, "✗", instance, addr, Reset)
		fmt.Printf("  %sError: %v%s\n",
			Red, err, Reset)
		return
	}

	role := info["role"]
	switch role {
	case "master":
		connectedSlaves := info["connected_slaves"]
		fmt.Printf("\n%s%s%s %s (%s)%s\n",
			Bold, Green, "✓", instance, addr, Reset)
		fmt.Printf("  %s%sRole:%s %sMaster%s\n",
			Dim, White, Reset, BrightGreen, Reset)
		fmt.Printf("  %s%sConnected Replicas:%s %s%s%s\n",
			Dim, White, Reset, BrightYellow, connectedSlaves, Reset)

		// Show clients
		clients := info["connected_clients"]
		fmt.Printf("  %s%sConnected Clients:%s %s%s%s\n",
			Dim, White, Reset, BrightMagenta, clients, Reset)

	case "slave":
		masterLinkStatus := info["master_link_status"]
		syncDelay := info["master_last_io_seconds_ago"]
		masterHost := info["master_host"]
		masterPort := info["master_port"]

		statusColor := Green
		statusSymbol := "✓"
		if masterLinkStatus != "up" {
			statusColor = Red
			statusSymbol = "✗"
		}

		syncStatusColor := BrightGreen
		if masterLinkStatus != "up" {
			syncStatusColor = BrightRed
		}

		fmt.Printf("\n%s%s%s %s (%s)%s\n",
			Bold, statusColor, statusSymbol, instance, addr, Reset)
		fmt.Printf("  %s%sRole:%s %sReplica%s\n",
			Dim, White, Reset, BrightBlue, Reset)
		fmt.Printf("  %s%sMaster:%s %s%s:%s%s\n",
			Dim, White, Reset, BrightYellow, masterHost, masterPort, Reset)
		fmt.Printf("  %s%sSync Status:%s %s%s%s\n",
			Dim, White, Reset,
			syncStatusColor,
			masterLinkStatus, Reset)
		fmt.Printf("  %s%sSync Delay:%s %s%ss%s\n",
			Dim, White, Reset, BrightCyan, syncDelay, Reset)
	}
}

func main() {
	instances := []struct {
		name string
		addr string
	}{
		{"Primary Node", "localhost:6379"},
		{"Replica Node 1", "localhost:6380"},
		{"Replica Node 2", "localhost:6381"},
	}

	for {
		clearScreen()
		printHeader()

		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("\n%s%s%sStatus at %s%s\n",
			Dim, Italic, White, timestamp, Reset)

		for _, instance := range instances {
			info, err := getRedisInfo(instance.addr)
			printInstanceStatus(instance.name, instance.addr, info, err)
		}

		// Print footer
		fmt.Printf("\n%s%s%sPress Ctrl+C to exit • Refreshing every 2s%s\n",
			Dim, Italic, White, Reset)

		time.Sleep(2 * time.Second)
	}
}
