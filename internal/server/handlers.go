package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/we-be/tritium/internal/storage"
)

func verifyBrowser(w http.ResponseWriter, r *http.Request) {
	// Extract the base64 encoded code from the URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	code := parts[2]
	fmt.Println("RECEIVED CODE", code)

	signature := "TEMPxSIG"

	// Set the cookie with the signature value
	cookie := &http.Cookie{
		Name:     "tritium_signature",
		Value:    signature,
		Secure:   false, // TODO prod set to true ofc
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("content-type", "text/plain")
	w.Write([]byte("SIGSET: " + signature))
	// w.Write([]byte("browser verified"))
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var signature string

	// Check if the signature is present in the X-TRITIUM-SIGNATURE header
	signature = r.Header.Get("X-TRITIUM-SIGNATURE")

	if signature == "" {
		// If the signature is not in the header, check the cookies
		cookie, err := r.Cookie("tritium_signature")
		if err != nil {
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}
		signature = cookie.Value
	}

	// Decode the base64-encoded signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		http.Error(w, "Invalid signature format", http.StatusBadRequest)
		return
	}

	// Obtain the sender's public key (you'll need to implement this based on your key management system)
	senderPublicKey := storage.GetSenderPublicKey(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// Verify the signature
	if !ed25519.Verify(senderPublicKey, body, signatureBytes) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	fmt.Println(body)

	w.WriteHeader(http.StatusOK)
}
