package server

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func RandLink(length int) string {
	// Generate a random byte slice of the specified length
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	randomString := base64.URLEncoding.EncodeToString(b)

	// Generate the secure link URL
	link := "http://localhost:8080/verify/" + randomString

	return link
}
