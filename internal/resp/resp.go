package resp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

var (
	ErrInvalidResp   = errors.New("invalid RESP data type")
	ErrInvalidLength = errors.New("invalid length")
	ErrInvalidConn   = errors.New("invalid connection")
)

// RESP data types
type RespType byte

const (
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

type RespCommand []byte

// NewCommand creates a new RESP command with variadic string arguments
func NewCommand(args ...string) *RespCommand {
	// Pre-calculate total length to avoid multiple allocations
	totalLen := 1 + len(strconv.Itoa(len(args))) + 2 // *<len>\r\n
	for _, arg := range args {
		totalLen += 1 + len(strconv.Itoa(len(arg))) + 2 + len(arg) + 2 // $<len>\r\n<data>\r\n
	}

	// Initialize with capacity
	cmd := make(RespCommand, 0, totalLen)

	// Append array header
	cmd = append(cmd, []byte(fmt.Sprintf("*%d\r\n", len(args)))...)

	// Append each argument as a bulk string
	for _, arg := range args {
		cmd = append(cmd, []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))...)
	}

	return &cmd
}

// Execute writes the command to the connection and returns the number of bytes written
// It's important to remember to read the string off the buffer before Executing other commands
func (cmd *RespCommand) Execute(conn net.Conn) (int, error) {
	if conn == nil {
		return 0, ErrInvalidConn
	}

	// Write in a single call to avoid partial writes
	return conn.Write(*cmd)
}

// ExecuteWithResponse writes the command and reads the response
func (cmd *RespCommand) ExecuteWithResponse(conn net.Conn, reader *Reader) (interface{}, error) {
	if conn == nil {
		return nil, ErrInvalidConn
	}

	// Write command
	_, err := cmd.Execute(conn)
	if err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	// Read response
	if reader == nil {
		reader = NewReader(conn)
	}
	return reader.ReadValue()
}

// Example usage:
func Example() {
	// Create a command
	cmd := NewCommand("SET", "mykey", "myvalue")

	// Connect to Redis/Garnet
	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Execute with response
	resp, err := cmd.ExecuteWithResponse(conn, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: %v\n", resp)
}
