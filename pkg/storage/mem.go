package storage

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/we-be/tritium/internal/resp"
)

type RespServer struct {
	conn   net.Conn
	reader *resp.Reader
	mu     sync.RWMutex
}

func NewRespServer(addr string) (*RespServer, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &RespServer{
		conn:   conn,
		reader: resp.NewReader(conn),
		mu:     sync.RWMutex{},
	}, nil
}

// SetEx writes the command to the connection and returns the number of bytes written
func (rs *RespServer) SetEx(key string, ttl int, value string) (int, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	nBytes, err := resp.NewCommand("SETEX", key, strconv.Itoa(ttl), value).Execute(rs.conn)
	isOK := rs.reader.IsOK()
	if err != nil || !isOK {
		return 0, fmt.Errorf("error setting key %s=%s", key, value)
	}
	return nBytes, nil
}

func (rs *RespServer) Get(key string) (interface{}, error) {
	return resp.NewCommand("GET", key).ExecuteWithResponse(rs.conn, rs.reader)
}
