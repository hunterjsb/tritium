package server

import (
	"fmt"
	"net"
	"net/rpc"
	"sync/atomic"

	"github.com/we-be/tritium/internal/config"
	"github.com/we-be/tritium/pkg/storage"
)

const DefaultTTL int = 17600

type Server struct {
	store    *storage.RespServer
	listener net.Listener
	rpc      *rpc.Server
	stats    ServerStats
	stopCh   chan struct{}
}

type ServerStats struct {
	ActiveConnections int64
	BytesTransferred  int64
}

// NewServer creates a new Tritium server
func NewServer(config config.Config) (*Server, error) {
	store, err := storage.NewRespServer(
		config.MemStoreAddr,
		config.MaxConnections,
		config.ReplicaAddrs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RESP server: %w", err)
	}

	srv := &Server{
		store:  store,
		rpc:    rpc.NewServer(),
		stopCh: make(chan struct{}),
	}

	// Register RPC methods
	if err := srv.rpc.RegisterName("Store", srv); err != nil {
		return nil, fmt.Errorf("failed to register RPC methods: %w", err)
	}

	return srv, nil
}

// Start starts the RPC server
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	s.listener = listener

	// Accept connections
	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		select {
		case <-s.stopCh:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					continue
				}
				// Non-temporary error or server shutdown
				return
			}
			atomic.AddInt64(&s.stats.ActiveConnections, 1)
			go s.serveConn(conn)
		}
	}
}

func (s *Server) serveConn(conn net.Conn) {
	defer func() {
		conn.Close()
		atomic.AddInt64(&s.stats.ActiveConnections, -1)
	}()
	s.rpc.ServeConn(conn)
}

// Set handles the Set RPC call
func (s *Server) Set(args *storage.SetArgs, reply *storage.SetReply) error {
	if args == nil {
		reply.Error = "invalid arguments"
		return nil
	}

	// Use default TTL if none provided
	ttl := DefaultTTL
	if args.TTL != nil {
		ttl = *args.TTL
	}

	value := string(args.Value)
	_, err := s.store.SetEx(args.Key, ttl, value)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	atomic.AddInt64(&s.stats.BytesTransferred, int64(len(args.Value)))
	return nil
}

// Get handles the Get RPC call
func (s *Server) Get(args *storage.GetArgs, reply *storage.GetReply) error {
	if args == nil {
		reply.Error = "invalid arguments"
		return nil
	}

	value, err := s.store.Get(args.Key)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			reply.Error = "key not found"
			return nil
		}
		reply.Value = v
		atomic.AddInt64(&s.stats.BytesTransferred, int64(len(v)))
	case string:
		if v == "" {
			reply.Error = "key not found"
			return nil
		}
		reply.Value = []byte(v)
		atomic.AddInt64(&s.stats.BytesTransferred, int64(len(v)))
	case nil:
		reply.Error = "key not found"
	default:
		reply.Error = "unexpected value type"
	}

	return nil
}

// Stats returns current server statistics
func (s *Server) Stats() ServerStats {
	return ServerStats{
		ActiveConnections: atomic.LoadInt64(&s.stats.ActiveConnections),
		BytesTransferred:  atomic.LoadInt64(&s.stats.BytesTransferred),
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	// Signal acceptLoop to stop
	close(s.stopCh)

	// Close listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	// Close RESP store
	if err := s.store.Close(); err != nil {
		return fmt.Errorf("failed to close store: %w", err)
	}

	return nil
}

func (s *Server) GetAddress() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}
