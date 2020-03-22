package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/criyle/go-judge/client"
	"github.com/criyle/go-judge/file"
	"github.com/criyle/go-judge/pkg/envexec"
	"github.com/criyle/go-judge/problem"
)

type task struct {
	j    job
	o    chan Model
	c    int32
	cmsg string
}

var _ client.Task = &task{}

func (t *task) Param() *client.JudgeTask {
	var buff bytes.Buffer
	json.NewEncoder(&buff).Encode(t.j.Lang)

	return &client.JudgeTask{
		Code: file.SourceCode{
			Code:     file.NewMemFile("code", []byte(t.j.Source)),
			Language: string(buff.Bytes()),
		},
		TimeLimit:   3 * time.Second / 2, // 1.5s
		MemoryLimit: 256 << 20,           // 256 M
	}
}

func (t *task) Parsed(c *problem.Config) {
	t.o <- Model{
		ID:     t.j.ID,
		Type:   "progress",
		Status: "Compiling",
	}
}

func (t *task) Compiled(c *client.ProgressCompiled) {
	log.Println(time.Now(), "compiled")
	t.cmsg = c.Message
	t.o <- Model{
		ID:     t.j.ID,
		Type:   "progress",
		Status: "Compiled",
	}
}

func (t *task) Progressed(*client.ProgressProgressed) {
	n := atomic.AddInt32(&t.c, 1)
	t.o <- Model{
		ID:     t.j.ID,
		Type:   "progress",
		Status: fmt.Sprintf("Judging (%d / %d)", n, noCase),
	}
}

func (t *task) Finished(rt *client.JudgeResult) {
	log.Println(time.Now(), "finished")
	var r []Result
	var status string
	if len(rt.SubTasks) > 0 {
		status = rt.SubTasks[0].Cases[0].ExecStatus.String()
		for _, ca := range rt.SubTasks[0].Cases {
			var ex string
			if ca.ExecStatus != envexec.StatusAccepted {
				status = ca.ExecStatus.String()
				ex = ca.ExecStatus.String()
			}
			r = append(r, Result{
				Time:   uint64(ca.Time.Round(time.Millisecond) / time.Millisecond),
				Memory: uint64(ca.Memory >> 10),
				Stdin:  string(ca.Input),
				Stdout: string(ca.UserOutput),
				Stderr: string(ca.UserError),
				Log:    string(ca.SPJOutput) + ca.Error + ex,
			})
		}
	} else {
		status = "compile failed"
		r = append(r, Result{
			Stderr: t.cmsg,
		})
	}
	t.o <- Model{
		ID:      t.j.ID,
		Type:    "finish",
		Status:  status,
		Results: r,
	}
}

type jClient struct {
	weburl        string
	retryInterval time.Duration
	c             chan client.Task
}

var _ client.Client = &jClient{}

func newClient(weburl string, retryInterval time.Duration) *jClient {
	c := &jClient{
		weburl:        weburl,
		retryInterval: retryInterval,
		c:             make(chan client.Task, 1),
	}
	go c.loop()
	return c
}

func (c *jClient) C() <-chan client.Task {
	return c.c
}

func (c *jClient) loop() {
	for {
		j, err := dialWS(c.weburl)
		if err != nil {
			log.Println("ws:", err)
			time.Sleep(c.retryInterval)
			continue
		}
		log.Println("ws connected")
		input := make(chan job, 64)
		output := make(chan Model, 64)

		// generate tasks
		go func() {
			for i := range input {
				c.c <- &task{
					j: i,
					o: output,
				}
			}
		}()

		judgerLoop(j, input, output)

		close(input)
		close(output)
	}
}

func judgerLoop(j *judgerWS, input chan job, output chan Model) {
	for {
		select {
		case <-j.disconnet:
			log.Println("ws disconneted")
			return

		case s := <-j.submit:
			log.Println("input: ", s)
			input <- s

		case o := <-output:
			//log.Println("output: ", o)
			j.update <- o
		}
	}
}
