package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func main() {
	// connect to database
	db := getDB()
	go db.loop()

	// creates hub
	hub := newHub(db)
	go hub.loop()

	// websocket
	http.HandleFunc("/ws", hub.handleWS)
	http.HandleFunc("/jws", hub.handleJudgerWS)

	// REST api
	a := api{db: db}
	http.Handle("/api/", http.StripPrefix("/api", a.serveMux()))

	// static files for spa
	s := spaFS{FileSystem: http.Dir("dist")}
	http.Handle("/", http.FileServer(s))

	// local test env
	addr := ":5000"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	// start serve
	log.Fatal(http.ListenAndServe(addr, nil))
}

type spaFS struct {
	http.FileSystem
}

func (fs spaFS) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil && os.IsNotExist(err) {
		return fs.FileSystem.Open("index.html")
	}
	return f, err
}
