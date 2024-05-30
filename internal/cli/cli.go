package cli

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/we-be/tritium/internal/server"
)

func Run() {
	lengthFlag := flag.Int("length", 32, "Length of the random link")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Please provide a command. Available commands:")
		fmt.Println("  generate-link: Generate a secure random link")
		os.Exit(1)
	}

	command := flag.Arg(0)
	switch command {
	case "generate-link":
		conn := connGarnet()
		link := server.RandLink(conn, *lengthFlag)
		fmt.Println(link)
		conn.Close()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func connGarnet() net.Conn {
	// Don't forget to close the connect
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	return conn
}
