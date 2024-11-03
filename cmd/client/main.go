package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"

	"github.com/we-be/tritium/pkg/storage"
)

func main() {
	var (
		addr    = flag.String("addr", "localhost:37381", "server address")
		command = flag.String("cmd", "get", "command (get/set)")
		key     = flag.String("key", "", "key to get/set")
		value   = flag.String("value", "", "value to set")
	)
	flag.Parse()

	if *key == "" {
		log.Fatal("key is required")
	}

	client, err := rpc.Dial("tcp", *addr)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()

	switch *command {
	case "set":
		if *value == "" {
			log.Fatal("value is required for set command")
		}
		args := &storage.SetArgs{
			Key:   *key,
			Value: []byte(*value),
		}
		var reply storage.SetReply
		err := client.Call("Store.Set", args, &reply)
		if err != nil {
			log.Fatal("Store.Set error:", err)
		}
		if reply.Error != "" {
			fmt.Printf("Error: %s\n", reply.Error)
		} else {
			fmt.Printf("Successfully set %s = %s\n", *key, *value)
		}

	case "get":
		args := &storage.GetArgs{
			Key: *key,
		}
		var reply storage.GetReply
		err := client.Call("Store.Get", args, &reply)
		if err != nil {
			log.Fatal("Store.Get error:", err)
		}
		if reply.Error != "" {
			fmt.Printf("Error: %s\n", reply.Error)
		} else {
			fmt.Printf("%s = %s\n", *key, string(reply.Value))
		}

	default:
		fmt.Printf("Unknown command: %s\n", *command)
		os.Exit(1)
	}
}
