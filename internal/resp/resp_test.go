package resp

import (
	"fmt"
	"net"
	"testing"
	"time"
)

const (
	serverAddr = "localhost:6379"
	timeout    = 2 * time.Second
)

// setupConnection establishes a connection with timeout
func setupConnection(t *testing.T) net.Conn {
	conn, err := net.DialTimeout("tcp", serverAddr, timeout)
	if err != nil {
		t.Fatalf("Failed to connect to %s: %v", serverAddr, err)
	}
	return conn
}

func TestIntegration_BasicOperations(t *testing.T) {
	conn := setupConnection(t)
	defer conn.Close()

	reader := NewReader(conn)

	tests := []struct {
		name     string
		cmd      *RespCommand
		validate func(interface{}, error) error
	}{
		{
			name: "PING",
			cmd:  NewCommand("PING"),
			validate: func(resp interface{}, err error) error {
				if err != nil {
					return fmt.Errorf("PING failed: %v", err)
				}
				str, ok := resp.(string)
				if !ok {
					return fmt.Errorf("expected string response, got %T", resp)
				}
				if str != "PONG" {
					return fmt.Errorf("expected 'PONG', got '%s'", str)
				}
				return nil
			},
		},
		{
			name: "SET key",
			cmd:  NewCommand("SET", "test_key", "test_value"),
			validate: func(resp interface{}, err error) error {
				if err != nil {
					return fmt.Errorf("SET failed: %v", err)
				}
				str, ok := resp.(string)
				if !ok {
					return fmt.Errorf("expected string response, got %T", resp)
				}
				if str != "OK" {
					return fmt.Errorf("expected 'OK', got '%s'", str)
				}
				return nil
			},
		},
		{
			name: "GET key",
			cmd:  NewCommand("GET", "test_key"),
			validate: func(resp interface{}, err error) error {
				if err != nil {
					return fmt.Errorf("GET failed: %v", err)
				}
				bytes, ok := resp.([]byte)
				if !ok {
					return fmt.Errorf("expected []byte response, got %T", resp)
				}
				if string(bytes) != "test_value" {
					return fmt.Errorf("expected 'test_value', got '%s'", string(bytes))
				}
				return nil
			},
		},
		{
			name: "DEL key",
			cmd:  NewCommand("DEL", "test_key"),
			validate: func(resp interface{}, err error) error {
				if err != nil {
					return fmt.Errorf("DEL failed: %v", err)
				}
				num, ok := resp.(int64)
				if !ok {
					return fmt.Errorf("expected int64 response, got %T", resp)
				}
				if num != 1 {
					return fmt.Errorf("expected 1 key deleted, got %d", num)
				}
				return nil
			},
		},
		{
			name: "GET deleted key",
			cmd:  NewCommand("GET", "test_key"),
			validate: func(resp interface{}, err error) error {
				if err != nil {
					return fmt.Errorf("GET failed: %v", err)
				}
				bytes, ok := resp.([]byte)
				if !ok {
					return fmt.Errorf("expected []byte response, got %T", resp)
				}
				if len(bytes) != 0 {
					return fmt.Errorf("expected empty []byte for deleted key, got %v", bytes)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.cmd.ExecuteWithResponse(conn, reader)
			if err := tt.validate(resp, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestIntegration_ArrayResponses(t *testing.T) {
	conn := setupConnection(t)
	defer conn.Close()

	reader := NewReader(conn)

	// Setup multiple keys
	keys := []string{"key1", "key2", "key3"}
	values := []string{"value1", "value2", "value3"}

	// Clean up any existing keys first
	cleanCmd := append([]string{"DEL"}, keys...)
	_, err := NewCommand(cleanCmd...).ExecuteWithResponse(conn, reader)
	if err != nil {
		t.Fatalf("Failed to clean up keys: %v", err)
	}

	// Set multiple keys
	for i := range keys {
		_, err := NewCommand("SET", keys[i], values[i]).ExecuteWithResponse(conn, reader)
		if err != nil {
			t.Fatalf("Failed to set key %s: %v", keys[i], err)
		}
	}

	// Test MGET
	t.Run("MGET multiple keys", func(t *testing.T) {
		cmd := append([]string{"MGET"}, keys...)
		resp, err := NewCommand(cmd...).ExecuteWithResponse(conn, reader)
		if err != nil {
			t.Fatalf("MGET failed: %v", err)
		}

		arr, ok := resp.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", resp)
		}

		if len(arr) != len(values) {
			t.Fatalf("expected %d values, got %d", len(values), len(arr))
		}

		for i, val := range arr {
			bytes, ok := val.([]byte)
			if !ok {
				t.Errorf("value %d: expected []byte, got %T", i, val)
				continue
			}
			if string(bytes) != values[i] {
				t.Errorf("value %d: expected %q, got %q", i, values[i], string(bytes))
			}
		}
	})

	// Cleanup
	_, err = NewCommand(cleanCmd...).ExecuteWithResponse(conn, reader)
	if err != nil {
		t.Errorf("Failed to clean up keys: %v", err)
	}
}

func TestIntegration_LargeValues(t *testing.T) {
	conn := setupConnection(t)
	defer conn.Close()

	reader := NewReader(conn)

	// Create a large value (1MB)
	largeValue := make([]byte, 1024*1024)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	// Test with large value
	t.Run("Large value SET/GET", func(t *testing.T) {
		// Set large value
		_, err := NewCommand("SET", "large_key", string(largeValue)).ExecuteWithResponse(conn, reader)
		if err != nil {
			t.Fatalf("Failed to set large value: %v", err)
		}

		// Get large value
		resp, err := NewCommand("GET", "large_key").ExecuteWithResponse(conn, reader)
		if err != nil {
			t.Fatalf("Failed to get large value: %v", err)
		}

		bytes, ok := resp.([]byte)
		if !ok {
			t.Fatalf("expected []byte response, got %T", resp)
		}

		if len(bytes) != len(largeValue) {
			t.Errorf("expected %d bytes, got %d bytes", len(largeValue), len(bytes))
		}

		// Clean up
		_, err = NewCommand("DEL", "large_key").ExecuteWithResponse(conn, reader)
		if err != nil {
			t.Errorf("Failed to clean up large key: %v", err)
		}
	})
}

func TestIntegration_BulkStringReply(t *testing.T) {
	conn := setupConnection(t)
	defer conn.Close()

	reader := NewReader(conn)

	testBytes := []byte{0x00, 0x01, 0x02, 0xFF}

	NewCommand("SET", "binary", string(testBytes)).Execute(conn)
	if !reader.IsOK() {
		t.Errorf("SET binary was not OK")
	}

	NewCommand("GET", "binary").Execute(conn)
	res, err := reader.ReadBulk()
	if err != nil {
		t.Errorf("error while reading bulk string %s", err)
	}

	if string(res) != string(testBytes) {
		t.Errorf("response %x does not match expected %x", res, testBytes)
	}

	NewCommand("DEL", "binary").Execute(conn)
	numDel, err := reader.ReadInt()
	if err != nil {
		t.Errorf("could not delete bc %s", err)
	}

	if numDel != 1 {
		t.Errorf("expected to delete 1, actually %d", numDel)
	}
}
