package storage

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/we-be/tritium/internal/resp"
)

type connPool struct {
	conns chan net.Conn
	addr  string
}

func newConnPool(addr string, maxConn int) (*connPool, error) {
	pool := &connPool{
		conns: make(chan net.Conn, maxConn),
		addr:  addr,
	}

	// Initialize connections
	for i := 0; i < maxConn; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection %d: %w", i, err)
		}
		pool.conns <- conn
	}

	return pool, nil
}

type RespServer struct {
	primaryPool *connPool
	replicas    []*connPool
	mu          sync.RWMutex
}

func NewRespServer(addr string, maxConn int, replicaAddrs []string) (*RespServer, error) {
	primaryPool, err := newConnPool(addr, maxConn)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary pool: %w", err)
	}

	rs := &RespServer{
		primaryPool: primaryPool,
		replicas:    make([]*connPool, 0, len(replicaAddrs)),
	}

	// Initialize replica pools
	for _, replicaAddr := range replicaAddrs {
		replicaPool, err := newConnPool(replicaAddr, maxConn)
		if err != nil {
			fmt.Printf("[warning] failed to create replica pool for %s: %v\n", replicaAddr, err)
			continue
		}
		rs.replicas = append(rs.replicas, replicaPool)
	}

	return rs, nil
}

func (rs *RespServer) SetEx(key string, ttl int, value string) (int, error) {
	// Get connection from pool
	conn := <-rs.primaryPool.conns
	defer func() { rs.primaryPool.conns <- conn }()

	reader := resp.NewReader(conn)
	cmd := resp.NewCommand("SETEX", key, strconv.Itoa(ttl), value)

	// Write to primary
	nBytes, err := cmd.Execute(conn)
	if err != nil {
		return 0, fmt.Errorf("primary write failed: %w", err)
	}

	if !reader.IsOK() {
		return 0, fmt.Errorf("primary write not OK")
	}

	// Replicate to replicas asynchronously
	var wg sync.WaitGroup

	// Use RLock when accessing replicas slice
	rs.mu.RLock()
	replicaCount := len(rs.replicas)
	replicas := make([]*connPool, replicaCount)
	copy(replicas, rs.replicas)
	rs.mu.RUnlock()

	for _, replica := range replicas {
		wg.Add(1)
		go func(pool *connPool) {
			defer wg.Done()

			replicaConn := <-pool.conns
			defer func() { pool.conns <- replicaConn }()

			replicaReader := resp.NewReader(replicaConn)
			replicaCmd := resp.NewCommand("SETEX", key, strconv.Itoa(ttl), value)

			if _, err := replicaCmd.Execute(replicaConn); err != nil {
				fmt.Printf("[warning] replica write failed on %s: %v\n", pool.addr, err)
				return
			}

			if !replicaReader.IsOK() {
				fmt.Printf("[warning] replica write not OK on %s\n", pool.addr)
			}
		}(replica)
	}

	// Wait for replicas to complete
	wg.Wait()

	return nBytes, nil
}

func (rs *RespServer) Get(key string) (interface{}, error) {
	// Get connection from pool
	conn := <-rs.primaryPool.conns
	defer func() { rs.primaryPool.conns <- conn }()

	reader := resp.NewReader(conn)
	return resp.NewCommand("GET", key).ExecuteWithResponse(conn, reader)
}

func (rs *RespServer) Close() error {
	// Close primary connections
	for i := 0; i < cap(rs.primaryPool.conns); i++ {
		conn := <-rs.primaryPool.conns
		if err := conn.Close(); err != nil {
			fmt.Printf("[warning] error closing primary connection: %v\n", err)
		}
	}

	// Close replica connections
	for _, replica := range rs.replicas {
		for i := 0; i < cap(replica.conns); i++ {
			conn := <-replica.conns
			if err := conn.Close(); err != nil {
				fmt.Printf("[warning] error closing replica connection: %v\n", err)
			}
		}
	}

	return nil
}

// AddReplica safely adds a new replica to the server
func (rs *RespServer) AddReplica(addr string, maxConn int) error {
	replicaPool, err := newConnPool(addr, maxConn)
	if err != nil {
		return fmt.Errorf("failed to create replica pool: %w", err)
	}

	rs.mu.Lock()
	rs.replicas = append(rs.replicas, replicaPool)
	rs.mu.Unlock()

	return nil
}

// RemoveReplica safely removes a replica from the server
func (rs *RespServer) RemoveReplica(addr string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	for i, replica := range rs.replicas {
		if replica.addr == addr {
			// Close all connections in the pool
			for j := 0; j < cap(replica.conns); j++ {
				conn := <-replica.conns
				if err := conn.Close(); err != nil {
					fmt.Printf("[warning] error closing replica connection: %v\n", err)
				}
			}

			// Remove from slice
			rs.replicas = append(rs.replicas[:i], rs.replicas[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("replica %s not found", addr)
}
