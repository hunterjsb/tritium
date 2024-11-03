package tritium_test

import (
	"fmt"
	"log"
	"time"

	"github.com/we-be/tritium/pkg/tritium"
)

func Example() {
	// Create a new client
	client, err := tritium.NewClient(&tritium.ClientOptions{
		Address: "localhost:40585",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Set a value
	err = client.Set("hello", []byte("world"), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get the value back
	value, err := client.Get("hello")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Value: %s\n", string(value))
	// Output: Value: world
}
