package main

import (
	"context"
	"net/http"
	"time"

	"github.com/criyle/go-judge-demo/pb"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

type judgeUpdater struct {
	client pb.DemoBackendClient
	logger *zap.Logger

	broadcast  chan *pb.JudgeUpdate
	observers  map[*observer]bool
	register   chan *observer
	unregister chan *observer
}

func newJudgeUpdater(client pb.DemoBackendClient, logger *zap.Logger) *judgeUpdater {
	ju := &judgeUpdater{
		client:     client,
		logger:     logger,
		broadcast:  make(chan *pb.JudgeUpdate, 64),
		observers:  make(map[*observer]bool),
		register:   make(chan *observer, 64),
		unregister: make(chan *observer, 64),
	}
	go ju.getUpdateLoop()
	go ju.broadcastLoop()
	return ju
}

func (j *judgeUpdater) Register(r *gin.RouterGroup) {
	r.GET("/judge", j.wsJudge)
}

func (j *judgeUpdater) getUpdateLoop() {
	for {
		j.logger.Debug("connected to updater")
		j.getUpdateSingle()
		j.logger.Debug("disconnected to updater")
		time.Sleep(5 * time.Second)
	}
}

func (j *judgeUpdater) getUpdateSingle() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := j.client.Updates(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	for {
		updates, err := c.Recv()
		j.logger.Debug("update", zap.Any("updates", updates))
		if err != nil {
			return err
		}
		j.broadcast <- updates
	}
}

func (j *judgeUpdater) broadcastLoop() {
	for {
		select {
		case c := <-j.register:
			j.observers[c] = true

		case c := <-j.unregister:
			delete(j.observers, c)

		case msg := <-j.broadcast:
			buf, err := protojson.Marshal(msg)
			if err != nil {
				j.logger.Debug("encode fail", zap.Error(err))
				continue
			}
			pMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, buf)
			if err != nil {
				j.logger.Debug("prepare fail", zap.Error(err))
				continue
			}
			j.logger.Debug("#observer", zap.Int("count", len(j.observers)))
			for o := range j.observers {
				select {
				case o.send <- pMsg:
				default:
					j.logger.Debug("too slow")
					// client too slow, stop it
					delete(j.observers, o)
					close(o.send)
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
)

func (j *judgeUpdater) wsJudge(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	o := &observer{
		ju:   j,
		conn: conn,
		send: make(chan *websocket.PreparedMessage, 64),
	}
	j.register <- o
	go o.loop()
	go o.pongLoop()
}

type observer struct {
	ju   *judgeUpdater
	conn *websocket.Conn
	send chan *websocket.PreparedMessage
}

func (c *observer) pongLoop() {
	defer func() {
		c.ju.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// loop starts to broadcast to clients
func (c *observer) loop() {
	defer c.conn.Close()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

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
				// TODO: log err
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
