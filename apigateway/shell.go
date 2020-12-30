package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/criyle/go-judger-demo/pb"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type shellHandle struct {
	client pb.DemoBackendClient
	logger *zap.Logger
}

func (s *shellHandle) Register(r *gin.RouterGroup) {
	r.GET("/shell", s.wsShell)
}

func (s *shellHandle) wsShell(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithCancel(c)
	sc, err := s.client.Shell(ctx)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("shell error: %v", err)))
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
		cancel()
		return
	}
	sh := &shell{
		conn:   conn,
		sc:     sc,
		cancel: cancel,
		msg:    make(chan *pb.ShellOutput, 16),
		logger: s.logger,
	}
	go sh.readLoop()
	go sh.recvLoop()
	go sh.writeLoop()
}

type shell struct {
	conn   *websocket.Conn
	sc     pb.DemoBackend_ShellClient
	cancel context.CancelFunc
	msg    chan *pb.ShellOutput
	logger *zap.Logger
}

func (s *shell) readLoop() {
	defer s.cancel()
	defer s.conn.Close()
	defer s.logger.Sugar().Debug("wsread exit")

	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			break
		}
		err = s.sc.Send(&pb.ShellInput{
			Request: &pb.ShellInput_Input{
				Input: &pb.Input{
					Content: msg,
				},
			},
		})
		if err != nil {
			break
		}
	}
}

func (s *shell) writeLoop() {
	defer s.cancel()
	defer s.conn.Close()
	defer s.logger.Sugar().Debug("wswrite exit")

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case msg, ok := <-s.msg:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			err := s.conn.WriteMessage(websocket.TextMessage, msg.GetContent())
			if err != nil {
				return
			}
		}
	}
}

func (s *shell) recvLoop() {
	defer s.cancel()
	defer close(s.msg)
	defer s.logger.Sugar().Debug("recv exit")

	for {
		msg, err := s.sc.Recv()
		if err != nil {
			return
		}
		s.msg <- msg
	}
}
