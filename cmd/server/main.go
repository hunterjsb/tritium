package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/we-be/tritium/internal/config"
	"github.com/we-be/tritium/internal/server"
)

func main() {
	cfg, err := config.NewConfigFromDotenv(".env")
	if err != nil {
		panic(err)
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}

	if err := srv.Start(":0"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Print the actual address
	fmt.Printf("Server is running on %s\n", srv.GetAddress())

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}
