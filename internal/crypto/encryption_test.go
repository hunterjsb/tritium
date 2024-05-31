package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	bits := 2048
	privateKey, publicKey, err := GenerateKeyPair(bits)
	if err != nil {
		t.Fatalf("GenerateKeyPair returned an error: %v", err)
	}

	if privateKey == nil || publicKey == nil {
		t.Fatal("GenerateKeyPair returned nil keys")
	}

	if privateKey.PublicKey.N.Cmp(publicKey.N) != 0 || privateKey.PublicKey.E != publicKey.E {
		t.Fatal("Generated public key does not match the private key's public key")
	}
}

func TestEncodePrivateKey(t *testing.T) {
	privateKey, _, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair returned an error: %v", err)
	}

	privateKeyPEM := EncodePrivateKey(privateKey)
	if privateKeyPEM == nil {
		t.Fatal("EncodePrivateKey returned nil")
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		t.Fatal("Failed to decode PEM block containing the private key")
	}

	if block.Type != "RSA PRIVATE KEY" {
		t.Fatalf("Expected block type 'RSA PRIVATE KEY', got %s", block.Type)
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	if parsedKey.N.Cmp(privateKey.N) != 0 || parsedKey.E != privateKey.E {
		t.Fatal("Encoded private key does not match the original key")
	}
}

func TestEncodePublicKey(t *testing.T) {
	_, publicKey, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair returned an error: %v", err)
	}

	publicKeyPEM, err := EncodePublicKey(publicKey)
	if err != nil {
		t.Fatalf("EncodePublicKey returned an error: %v", err)
	}

	if publicKeyPEM == nil {
		t.Fatal("EncodePublicKey returned nil")
	}

	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		t.Fatal("Failed to decode PEM block containing the public key")
	}

	if block.Type != "RSA PUBLIC KEY" {
		t.Fatalf("Expected block type 'RSA PUBLIC KEY', got %s", block.Type)
	}

	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	rsaParsedKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("Parsed key is not an RSA public key")
	}

	if rsaParsedKey.N.Cmp(publicKey.N) != 0 || rsaParsedKey.E != publicKey.E {
		t.Fatal("Encoded public key does not match the original key")
	}
}
