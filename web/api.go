package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type api struct {
	*db
}

func (a *api) apiSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "TAT", http.StatusBadRequest)
		return
	}
	id := ""
	if q := r.URL.Query(); q["id"] != nil && len(q["id"]) > 0 {
		id = q["id"][0]
	}
	m, err := a.db.Query(id)
	if err != nil {
		log.Println("api submission:", err)
		http.Error(w, "QAQ", http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(m)
}

func (a *api) handleAPI(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "QAQ", http.StatusNotImplemented)
}

func (a *api) serveMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/submission", a.apiSubmission)
	mux.HandleFunc("/", a.handleAPI)
	return mux
}
