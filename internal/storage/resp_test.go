package storage

import (
	"fmt"
	"net"
	"testing"
)

func TestGarnet(t *testing.T) {
	// Connect to the Garnet server
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create a new RESP reader
	reader := NewReader(conn)

	// Send a PING command to check the connection
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		panic(err)
	}

	// Read the response using the RESP reader
	pong, err := reader.ReadValue()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to Garnet server:", pong)

	// Send a SET command to set a key-value pair
	_, err = conn.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"))
	if err != nil {
		panic(err)
	}

	// Read the response using the RESP reader
	setResponse, err := reader.ReadValue()
	if err != nil {
		panic(err)
	}
	fmt.Println("SET response:", setResponse)

	// Send a GET command to retrieve the value of a key
	_, err = conn.Write([]byte("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"))
	if err != nil {
		panic(err)
	}

	// Read the response using the RESP reader
	getResponse, err := reader.ReadValue()
	if err != nil {
		panic(err)
	}
	fmt.Println("GET response:", string(getResponse.([]byte)))
}
