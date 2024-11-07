package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/we-be/tritium/internal/config"
	"github.com/we-be/tritium/internal/server"
)

func main() {
	configFile := flag.String("config", ".env", "Path to configuration file")
	flag.Parse()

	cfg, err := config.NewConfigFromDotenv(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(cfg.RPCAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Printf("Server is running on %s\n", srv.GetAddress())
	fmt.Printf("Using memory store at %s\n", cfg.MemStoreAddr)
	if cfg.JoinAddr != "" {
		fmt.Printf("Joining cluster via %s\n", cfg.JoinAddr)
	} else {
		fmt.Println("Starting as cluster leader")
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	if err := srv.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
