package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ClientSubmit is the job submit model uploaded by client ws
type ClientSubmit struct {
	Lang string `json:"language"`
	Code string `json:"code"`
}

type clientSubmitJob struct {
	data   ClientSubmit
	client *client
}

type client struct {
	hub  *hub
	conn *websocket.Conn
	send chan interface{}
}

func (c *client) readLoop() {
	defer func() {
		c.hub.unregisterClient <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var v ClientSubmit
	for {
		err := c.conn.ReadJSON(&v)
		if err != nil {
			log.Println("client ws read error: ", err)
			break
		}
		log.Println("client ws read: ", v)
		c.hub.clientUpload <- clientSubmitJob{v, c}
	}
}

func (c *client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case m, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			err := c.conn.WriteJSON(m)
			if err != nil {
				log.Println("client ws write: ", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleClientWS(h *hub, w http.ResponseWriter, r *http.Request) {
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
		send: make(chan interface{}, 64),
	}
	h.registerClient <- c
	go c.readLoop()
	go c.writeLoop()
}
