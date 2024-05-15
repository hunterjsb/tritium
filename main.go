package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	signature := r.Header.Get("X-TRITIUM-SIGNATURE")
	if signature == "" {
		http.Error(w, "Missing signature header", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	fmt.Println(body)

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/message", handleMessage)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
