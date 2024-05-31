package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log"
	"net"

	"github.com/we-be/tritium/internal/storage"
)

const TOKEN_EXPIRY string = "172800" // two days in seconds

func RandLink(conn net.Conn, length int) string {
	// Generate an RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// Extract the public key
	publicKey := &privateKey.PublicKey

	// Encode the private key as PEM
	// privateKeyPEM := pem.EncodeToMemory(&pem.Block{
	//     Type:  "RSA PRIVATE KEY",
	//     Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	// })

	// Encode the public key as PEM
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	})

	// Generate a random byte slice of the specified length
	b := make([]byte, length)
	_, err = rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	randomString := base64.URLEncoding.EncodeToString(b)

	// Store the private key and public key in the user's browser (you'll need to implement this part)
	// For example, you can use cookies or local storage to store the keys

	// Store the server's public key in the database
	_, err = storage.NewCommand("SETEX", randomString+"_server_public_key", TOKEN_EXPIRY, string(publicKeyPEM)).Execute(conn)
	if err != nil {
		panic(err)
	}

	// Generate the secure link URL
	link := "http://localhost:8080/verify/" + randomString

	return link
}
