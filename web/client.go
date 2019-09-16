package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ClientSubmit is the job submit model uploaded by client ws
type ClientSubmit struct {
	Lang   Language `json:"language"`
	Source string   `json:"source"`
}

type clientSubmitJob struct {
	data   ClientSubmit
	client *client
}

type client struct {
	hub  *clientHub
	conn *websocket.Conn
	send chan *websocket.PreparedMessage
}

// loop starts to broadcast to clients
func (c *client) loop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.hub.unregister <- c
		ticker.Stop()
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		select {
		case m, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			err := c.conn.WritePreparedMessage(m)
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

type clientHub struct {
	clients              map[*client]bool
	broadcast            chan interface{}
	register, unregister chan *client
}

func newClientHub() *clientHub {
	return &clientHub{
		clients:    make(map[*client]bool),
		broadcast:  make(chan interface{}, 64),
		register:   make(chan *client, 64),
		unregister: make(chan *client, 64),
	}
}

func (h *clientHub) Broadcast(m interface{}) {
	h.broadcast <- m
}

func (h *clientHub) handleWS(w http.ResponseWriter, r *http.Request) {
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
	h.register <- c
	go c.loop()
}

func (h *clientHub) loop() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true

		case c := <-h.unregister:
			if h.clients[c] {
				delete(h.clients, c)
				close(c.send)
			}

		case da := <-h.broadcast:
			// prepare message
			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(da)
			if err != nil {
				log.Println("JSON encode error", err)
				continue
			}
			pMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, buf.Bytes())
			if err != nil {
				log.Println("prepare message error", err)
				continue
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
		}
	}
}
