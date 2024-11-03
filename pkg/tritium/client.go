// pkg/tritium/client.go
package tritium

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/we-be/tritium/pkg/storage"
)

const PORT string = ":40585" // ADD YOUR PORT!!

// Client represents a Tritium RPC client
type Client struct {
	rpc *rpc.Client
}

// ClientOptions contains options for creating a new client
type ClientOptions struct {
	Address string
	Timeout time.Duration
}

// NewClient creates a new Tritium client
func NewClient(opts *ClientOptions) (*Client, error) {
	client, err := rpc.Dial("tcp", PORT)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tritium server: %w", err)
	}

	return &Client{
		rpc: client,
	}, nil
}

// Set stores a value with an optional TTL
func (c *Client) Set(key string, value []byte, ttl *int) error {
	args := &storage.SetArgs{
		Key:   key,
		Value: value,
		TTL:   ttl,
	}
	var reply storage.SetReply
	if err := c.rpc.Call("Store.Set", args, &reply); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	if reply.Error != "" {
		return fmt.Errorf("server error: %s", reply.Error)
	}
	return nil
}

// Get retrieves a value
func (c *Client) Get(key string) ([]byte, error) {
	args := &storage.GetArgs{
		Key: key,
	}
	var reply storage.GetReply
	if err := c.rpc.Call("Store.Get", args, &reply); err != nil {
		return nil, fmt.Errorf("failed to get value: %w", err)
	}
	if reply.Error != "" {
		return nil, fmt.Errorf("server error: %s", reply.Error)
	}
	return reply.Value, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.rpc.Close()
}
