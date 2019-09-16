package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type api struct {
	*db
	*judgerHub
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

func (a *api) apiSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "TAT", http.StatusBadRequest)
		return
	}
	var cs ClientSubmit
	if err := json.NewDecoder(r.Body).Decode(&cs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m, err := a.Add(&cs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.Submit(*m)
}

func (a *api) api(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "QAQ", http.StatusNotImplemented)
}

func (a *api) serveMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/submission", a.apiSubmission)
	mux.HandleFunc("/submit", a.apiSubmit)
	mux.HandleFunc("/", a.api)
	return mux
}
