package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/criyle/go-judge/pb"
	"github.com/gorilla/websocket"
)

const (
	cpuLimit     = 10 * time.Second
	sessionLimit = 10 * time.Minute
	procLimit    = 50
	memoryLimit  = 256 << 20 // 256m
)

var (
	args = []string{"/bin/bash"}
	env  = []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"HOME=/w",
		"TERM=xterm",
	}
)

type shellCoon struct {
	db   *db
	conn *websocket.Conn

	read  chan []byte
	write chan []byte

	input  []byte
	output []byte
}

func (sc *shellCoon) readLoop() {
	defer sc.conn.Close()
	defer close(sc.read)

	sc.conn.SetReadDeadline(time.Now().Add(pongWait))
	sc.conn.SetPongHandler(func(string) error {
		sc.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		typ, reader, err := sc.conn.NextReader()
		if err != nil {
			log.Println("shell read", err)
			return
		}
		_ = typ
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Println("shell read ", err)
			return
		}
		sc.input = append(sc.input, b...)
		sc.read <- b
	}
}

func (sc *shellCoon) writeLoop() {
	defer sc.conn.Close()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case b, ok := <-sc.write:
			sc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				sc.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			sc.output = append(sc.output, b...)

			writer, err := sc.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("shell write", err)
				return
			}
			writer.Write(b)
			writer.Close()

		case <-ticker.C:
			sc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := sc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}

	}
}

type shell struct {
	db     *db
	client pb.ExecutorClient
}

func (s *shell) handleShellWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("handleShell", err)
		return
	}
	defer conn.Close()

	sc := &shellCoon{
		db:    s.db,
		conn:  conn,
		read:  make(chan []byte, 64),
		write: make(chan []byte, 64),
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sc.readLoop()
	}()

	go func() {
		defer wg.Done()
		sc.writeLoop()
	}()

	err = serveShell(sc, s.client)
	if err != nil {
		sc.write <- []byte("\n\rError: " + err.Error())
		log.Println("handleShellWS: ", err)
	}
	close(sc.write)
	wg.Wait()

	log.Println(r, "input\n", string(sc.input), "\noutput\n", string(sc.output))
	s.db.Store(&ShellStore{
		Stdin:  string(sc.input),
		Stdout: string(sc.output),
	})
}

func serveShell(sc *shellCoon, client pb.ExecutorClient) error {
	// create shell
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	stream, err := client.ExecStream(ctx)
	if err != nil {
		return err
	}
	err = stream.Send(&pb.StreamRequest{
		Request: &pb.StreamRequest_ExecRequest{
			ExecRequest: shellReq,
		},
	})
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)

	// pump stdin
	go func() {
		for {
			b, ok := <-sc.read
			if !ok {
				return
			}
			err := stream.Send(&pb.StreamRequest{
				Request: &pb.StreamRequest_ExecInput{
					ExecInput: &pb.StreamRequest_Input{
						Name:    "i",
						Content: b,
					},
				},
			})
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	// pump stdout
	for {
		sr, err := stream.Recv()
		if err != nil {
			return err
		}
		switch sr := sr.Response.(type) {
		case *pb.StreamResponse_ExecOutput:
			sc.write <- sr.ExecOutput.GetContent()

		case *pb.StreamResponse_ExecResponse:
			sc.write <- []byte(sr.ExecResponse.String())
			select {
			case err := <-errCh:
				return err
			default:
				return nil
			}
		}
	}
}

var (
	shellReq = &pb.Request{
		Cmd: []*pb.Request_CmdType{{
			Args: args,
			Env:  env,
			Files: []*pb.Request_File{
				{
					File: &pb.Request_File_StreamIn{
						StreamIn: &pb.Request_StreamInput{
							Name: "i",
						},
					},
				},
				{
					File: &pb.Request_File_StreamOut{
						StreamOut: &pb.Request_StreamOutput{
							Name: "o",
						},
					},
				},
				{
					File: &pb.Request_File_StreamOut{
						StreamOut: &pb.Request_StreamOutput{
							Name: "o",
						},
					},
				},
			},
			CPULimit:     uint64(cpuLimit),
			RealCPULimit: uint64(sessionLimit),
			MemoryLimit:  memoryLimit,
			ProcLimit:    procLimit,
			Tty:          true,
		}},
	}
)
