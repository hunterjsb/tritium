package server

import (
	"net/rpc"
	"testing"

	"github.com/we-be/tritium/internal/config"
	"github.com/we-be/tritium/pkg/storage"
)

const testAddr = ":0" // Let OS assign a random port

func TestServerConnectivity(t *testing.T) {
	// Create server
	cfg := config.Config{
		MemStoreAddr:   "localhost:6379",
		MaxConnections: 10,
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := srv.Start(testAddr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	// defer srv.Close()

	// Get the actual address that was assigned
	addr := srv.listener.Addr().String()
	t.Logf("Server listening on %s", addr)

	// Connect RPC client
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer client.Close()

	// Simple ping test
	args := &storage.GetArgs{Key: "test-key"}
	reply := &storage.GetReply{}

	err = client.Call("Store.Get", args, reply)
	if err != nil {
		t.Fatalf("RPC call failed: %v", err)
	}

	// We expect a "key not found" error for a non-existent key
	if reply.Error != "key not found" {
		t.Errorf("Expected 'key not found' error, got: %v", reply.Error)
	}
}

func TestServerReadWrite(t *testing.T) {
	// Create server
	cfg := config.Config{
		MemStoreAddr:   "localhost:6379",
		MaxConnections: 10,
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := srv.Start(testAddr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	// defer srv.Close()

	// Get the actual address that was assigned
	addr := srv.listener.Addr().String()

	// Connect RPC client
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer client.Close()

	// Test data
	testKey := "test-key-1"
	testValue := []byte("Hello, Tritium!")

	// Test Set operation
	setArgs := &storage.SetArgs{
		Key:   testKey,
		Value: testValue,
	}
	setReply := &storage.SetReply{}

	err = client.Call("Store.Set", setArgs, setReply)
	if err != nil {
		t.Fatalf("Set RPC call failed: %v", err)
	}
	if setReply.Error != "" {
		t.Fatalf("Set operation failed: %v", setReply.Error)
	}

	// Test Get operation - should retrieve the value we just set
	getArgs := &storage.GetArgs{Key: testKey}
	getReply := &storage.GetReply{}

	err = client.Call("Store.Get", getArgs, getReply)
	if err != nil {
		t.Fatalf("Get RPC call failed: %v", err)
	}
	if getReply.Error != "" {
		t.Fatalf("Get operation failed: %v", getReply.Error)
	}

	// Verify the value matches what we set
	if string(getReply.Value) != string(testValue) {
		t.Errorf("Expected value %q, got %q", string(testValue), string(getReply.Value))
	}

	// Test Get with non-existent key
	getNonExistentArgs := &storage.GetArgs{Key: "non-existent-key"}
	getNonExistentReply := &storage.GetReply{}

	err = client.Call("Store.Get", getNonExistentArgs, getNonExistentReply)
	if err != nil {
		t.Fatalf("Get RPC call failed: %v", err)
	}
	if getNonExistentReply.Error != "key not found" {
		t.Errorf("Expected 'key not found' error, got: %v", getNonExistentReply.Error)
	}

	// Verify server stats
	if srv.stats.ActiveConnections != 1 {
		t.Errorf("Expected 1 active connection, got %d", srv.stats.ActiveConnections)
	}
	expectedBytes := int64(len(testValue))
	if srv.stats.BytesTransferred != expectedBytes {
		t.Errorf("Expected %d bytes transferred, got %d", expectedBytes, srv.stats.BytesTransferred)
	}
}
