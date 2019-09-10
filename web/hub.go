package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
)

type hub struct {
	// Registed connections
	clients map[*client]bool
	judgers map[*judger]bool

	// Datebase
	db *db

	// Operations
	registerClient, unregisterClient chan *client
	registerJudger, unregisterJudger chan *judger

	// Transmissions
	judgerUpdate    chan JudgerUpdate
	clientUpload    chan clientSubmitJob
	clientBroadcast chan interface{}
}

func (h *hub) loop() {
	for {
		select {
		case c := <-h.registerClient:
			h.clients[c] = true

		case c := <-h.unregisterClient:
			if h.clients[c] {
				delete(h.clients, c)
				close(c.send)
			}

		case j := <-h.registerJudger:
			h.judgers[j] = true

		case j := <-h.unregisterJudger:
			if h.judgers[j] {
				delete(h.judgers, j)
				close(j.submit)
			}

		case ju := <-h.judgerUpdate:
			h.db.update <- ju

		case da := <-h.clientBroadcast:
			// prepare message
			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(da)
			if err != nil {
				log.Println("JSON encode error", err)
			}
			pMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, buf.Bytes())
			if err != nil {
				log.Println("prepare message error", err)
			}

			// broadcast to clients
			for c := range h.clients {
				select {
				case c.send <- pMsg:
				default:
					delete(h.clients, c)
					close(c.send)
				}
			}

		case ju := <-h.db.updateDone:
			h.clientBroadcast <- ju

		case cs := <-h.clientUpload:
			h.db.insert <- cs.data

		case m := <-h.db.insertDone:
			// TODO: load balancer
			jo := job{
				ID:   m.ID.Hex(),
				Lang: m.Lang,
				Code: m.Code,
			}

			// submit job to judger
			submited := false
			for j := range h.judgers {
				select {
				case j.submit <- jo:
					submited = true
				default:
					delete(h.judgers, j)
					close(j.submit)
				}
			}
			if !submited {
				log.Println("not judger have logged in")
			}
			h.clientBroadcast <- m
		}
	}
}

func newHub(db *db) *hub {
	return &hub{
		clients:          make(map[*client]bool),
		judgers:          make(map[*judger]bool),
		db:               db,
		registerClient:   make(chan *client, 64),
		unregisterClient: make(chan *client, 64),
		registerJudger:   make(chan *judger, 64),
		unregisterJudger: make(chan *judger, 64),
		judgerUpdate:     make(chan JudgerUpdate, 64),
		clientUpload:     make(chan clientSubmitJob, 64),
		clientBroadcast:  make(chan interface{}, 64),
	}
}

func (h *hub) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("client ws:", err)
		http.Error(w, "TAT", http.StatusUpgradeRequired)
		return
	}
	log.Println("client ws:", "new connection", r)
	c := &client{
		hub:  h,
		conn: conn,
		send: make(chan *websocket.PreparedMessage, 64),
	}
	h.registerClient <- c
	go c.readLoop()
	go c.writeLoop()
}

func (h *hub) handleJudgerWS(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv(envJudgerToken)
	// no token == dev env
	if token != "" {
		if r.Header["Authorization"] == nil || r.Header["Authorization"][1] != token {
			log.Println("judger ws: auth failed", r)
			http.Error(w, "QAQ", http.StatusUnauthorized)
			return
		}
	}
	log.Println("judger ws: logged in", r)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	j := &judger{
		hub:    h,
		conn:   conn,
		submit: make(chan job),
	}
	h.registerJudger <- j
	go j.readLoop()
	go j.writeLoop()
}
