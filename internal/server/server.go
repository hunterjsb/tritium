// internal/server/server.go
package server

import (
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
}

type ServerStats struct {
	ActiveConnections int64
	BytesTransferred  int64
}

// NewServer creates a new Tritium server
func NewServer(config config.Config) (*Server, error) {
	store, err := storage.NewRespServer(config.MemStoreAddr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		store: store,
		rpc:   rpc.NewServer(),
	}

	// Register RPC methods
	if err := srv.rpc.RegisterName("Store", srv); err != nil {
		return nil, err
	}

	return srv, nil
}

// Start starts the RPC server
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	// Accept connections
	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Handle error or server shutdown
			return
		}
		atomic.AddInt64(&s.stats.ActiveConnections, 1)
		go s.serveConn(conn)
	}
}

func (s *Server) serveConn(conn net.Conn) {
	defer func() {
		conn.Close()
		atomic.AddInt64(&s.stats.ActiveConnections, -1)
	}()
	s.rpc.ServeConn(conn)
}

// In server.go
func (s *Server) Set(args *storage.SetArgs, reply *storage.SetReply) error {
	value := string(args.Value)
	_, err := s.store.SetEx(args.Key, DefaultTTL, value)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	atomic.AddInt64(&s.stats.BytesTransferred, int64(len(args.Value)))
	return nil
}

func (s *Server) Get(args *storage.GetArgs, reply *storage.GetReply) error {
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
	case string:
		if v == "" {
			reply.Error = "key not found"
			return nil
		}
		reply.Value = []byte(v)
	case nil:
		reply.Error = "key not found"
	default:
		reply.Error = "unexpected value type"
	}

	return nil
}

func (s *Server) GetAddress() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}
