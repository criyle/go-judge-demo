package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func apiSubmission(hub *hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "TAT", http.StatusBadRequest)
		return
	}
	id := ""
	if q := r.URL.Query(); q["id"] != nil && len(q["id"]) > 0 {
		id = q["id"][0]
	}
	m, err := hub.db.Query(id)
	if err != nil {
		log.Println("api submission:", err)
		http.Error(w, "QAQ", http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(m)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "QAQ", http.StatusNotImplemented)
}
