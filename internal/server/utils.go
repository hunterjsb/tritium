package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
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
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

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

	// Store the keys in the database
	_, privErr := storage.NewCommand("SETEX", fmt.Sprintf("server:%s:privkey", randomString), TOKEN_EXPIRY, string(privateKeyPEM)).Execute(conn)
	_, pubErr := storage.NewCommand("SETEX", fmt.Sprintf("server:%s:pubkey", randomString), TOKEN_EXPIRY, string(publicKeyPEM)).Execute(conn)
	if privErr != nil || pubErr != nil {
		panic("error setting key in storage")
	}

	// Generate the secure link URL
	link := "http://localhost:8080/verify/" + randomString

	return link
}
