package storage

import (
	"fmt"
	"net/http"
)

func GetSenderPublicKey(r *http.Request) []byte {
	// Implement this function to retrieve the sender's public key based on your key management system
	// You might extract the sender's identity from the request headers or query parameters
	// and look up their public key in a database or key store
	// Return the public key as a byte slice
	// ...
	fmt.Println(r)
	return make([]byte, 0)
}
