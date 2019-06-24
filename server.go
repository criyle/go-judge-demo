package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func main() {
	// creates hub
	hub := newHub()
	go hub.loop()

	// websocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleClientWS(hub, w, r)
	})
	http.HandleFunc("/jws", func(w http.ResponseWriter, r *http.Request) {
		handleJudgerWS(hub, w, r)
	})

	// REST api
	http.HandleFunc("/api/submission", func(w http.ResponseWriter, r *http.Request) {
		apiSubmission(hub, w, r)
	})
	http.HandleFunc("/api", handleAPI)

	// static files
	http.Handle("/", http.FileServer(http.Dir("app/dist")))

	// local test env
	addr := ":5000"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	// start serve
	log.Fatal(http.ListenAndServe(addr, nil))
}
