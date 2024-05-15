package server

import (
	"log"
	"net/http"
)

func Start() {
	http.HandleFunc("/verify/", verifyBrowser)
	http.HandleFunc("/message", handleMessage)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
