package server

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"
)

func RandLink(length int) string {
	// Generate a random byte slice of the specified length
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	// Add additional entropy using the current timestamp
	timestamp := time.Now().UnixNano()
	timestampBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		timestampBytes[i] = byte(timestamp >> (i * 8))
	}
	b = append(b, timestampBytes...)

	// Encode the byte slice using base64 RawURLEncoding
	randomString := base64.RawURLEncoding.EncodeToString(b)

	// Generate the secure link URL
	link := "https://tritium.example.com/verify/" + randomString

	return link
}
