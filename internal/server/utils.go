package server

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net"

	"github.com/we-be/tritium/internal/storage"
)

const TOKEN_EXPIRY string = "172800" // two days in seconds

func RandLink(conn net.Conn, length int) string {
	// Generate a random byte slice of the specified length
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	randomString := base64.URLEncoding.EncodeToString(b)
	_, err = storage.NewCommand("SETEX", randomString, TOKEN_EXPIRY, "TEMPxSIG").Execute(conn)
	if err != nil {
		panic(err)
	}

	// Generate the secure link URL
	link := "http://localhost:8080/verify/" + randomString

	return link
}
