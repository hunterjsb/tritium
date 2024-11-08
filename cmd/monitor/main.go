package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/we-be/tritium/internal/monitor"
)

func main() {
	summary := flag.Bool("summary", false, "Show concise summary view")
	flag.Parse()

	m := monitor.New()
	ticker := time.NewTicker(2 * time.Second)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			monitor.ClearScreen()
			m.PrintHeader()

			timestamp := time.Now().Format("15:04:05")
			fmt.Printf("\n%s%s%sStatus at %s%s\n",
				monitor.Dim, monitor.Italic, monitor.White, timestamp, monitor.Reset)

			nodes, err := m.GatherNodeInfo()
			respInfo := m.GatherRespInfo()

			if err != nil {
				fmt.Printf("\n%s%s✗ No RPC nodes found%s\n",
					monitor.Bold, monitor.Red, monitor.Reset)
			} else if *summary {
				m.PrintSummary(nodes, respInfo)
			} else {
				m.PrintDetailed(nodes, respInfo)
			}

			fmt.Printf("\n%s%s%sPress Ctrl+C to exit • Refreshing every 2s%s\n",
				monitor.Dim, monitor.Italic, monitor.White, monitor.Reset)

		case <-sigChan:
			fmt.Println("\nShutting down...")
			return
		}
	}
}
