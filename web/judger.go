package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	envJudgerToken = "JUDGER_TOKEN"
)

type job struct {
	ID     string   `json:"id"`
	Lang   Language `json:"language"`
	Source string   `json:"source"`
}

type judger struct {
	hub    *judgerHub
	conn   *websocket.Conn // connection
	submit chan job        // to submit the job
}

func (j *judger) readLoop() {
	defer func() {
		j.hub.unregister <- j
		j.conn.Close()
	}()

	j.conn.SetReadDeadline(time.Now().Add(pongWait))
	j.conn.SetPongHandler(func(string) error {
		j.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		v := JudgerUpdate{}
		err := j.conn.ReadJSON(&v)
		if err != nil {
			log.Println("judger ws: ", err)
			break
		}
		j.hub.update <- v
	}
}

func (j *judger) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		j.conn.Close()
	}()
	for {
		select {
		case m, ok := <-j.submit:
			j.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				j.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			err := j.conn.WriteJSON(m)
			if err != nil {
				log.Println("judger ws: ", err)
				return
			}

		case <-ticker.C:
			j.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := j.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type judgerHub struct {
	judgers              map[*judger]bool
	update               chan JudgerUpdate
	register, unregister chan *judger
	submit               chan Model
	*db
	*clientHub
}

func newJudgerHub(db *db, clientHub *clientHub) *judgerHub {
	return &judgerHub{
		judgers:    make(map[*judger]bool),
		register:   make(chan *judger, 64),
		unregister: make(chan *judger, 64),
		update:     make(chan JudgerUpdate, 64),
		submit:     make(chan Model, 64),
		db:         db,
		clientHub:  clientHub,
	}
}

func (h *judgerHub) Submit(m Model) {
	h.submit <- m
}

func (h *judgerHub) loop() {
	for {
		select {
		case j := <-h.register:
			h.judgers[j] = true

		case j := <-h.unregister:
			if h.judgers[j] {
				delete(h.judgers, j)
				close(j.submit)
			}

		case ju := <-h.update:
			ud, err := h.Update(&ju)
			if err != nil {
				log.Println("db update:", err)
				continue
			}
			h.clientHub.broadcast <- ud

		case m := <-h.submit:
			jo := job{
				ID:     m.ID.Hex(),
				Lang:   m.Lang,
				Source: m.Source,
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
				log.Println("no judger have logged in")
				continue
			}
			h.clientHub.Broadcast(m)
		}
	}
}

func (h *judgerHub) handleJudgerWS(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		log.Println("judger ws: auth failed", r)
		http.Error(w, "QAQ", http.StatusUnauthorized)
		return
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
	h.register <- j
	go j.readLoop()
	go j.writeLoop()
}

func checkToken(r *http.Request) bool {
	token := os.Getenv(envJudgerToken)
	if token == "" {
		return true
	}
	auth := r.Header["Authorization"]
	if auth == nil || len(auth) != 2 || auth[1] != token {
		return false
	}
	return true
}
