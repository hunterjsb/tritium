package server

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func randLink(length int) string {
	// Generate a random 64-character base64 string
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	randomString := base64.URLEncoding.EncodeToString(b)

	// Generate the secure link URL
	secureLink := "https://tritium.example.com/verify/" + randomString

	return secureLink
}
