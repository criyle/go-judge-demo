package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	envJudgerToken = "JUDGER_TOKEN"
)

type job struct {
	ID   string `json:"id"`
	Lang string `json:"language"`
	Code string `json:"code"`
}

type judger struct {
	hub    *hub
	conn   *websocket.Conn // connection
	submit chan job        // to submit the job
}

func (j *judger) readLoop() {
	defer func() {
		j.hub.unregisterJudger <- j
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
		j.hub.judgerUpdate <- v
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
