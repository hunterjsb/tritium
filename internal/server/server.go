package server

import (
	"log"
	"net/http"
)

func Start() {
	http.HandleFunc("/set-cookie/", verifyBrowser)
	http.HandleFunc("/message", handleMessage)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
