package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/criyle/go-judge-client/pkg/diff"
	"github.com/criyle/go-judge/pb"
	demopb "github.com/criyle/go-judger-demo/pb"
	"github.com/flynn/go-shlex"
	"golang.org/x/sync/errgroup"
)

type judger struct {
	execClient pb.ExecutorClient
	demoClient demopb.DemoBackendClient

	request  chan *demopb.JudgeClientRequest
	response chan *demopb.JudgeClientResponse
}

func newJudger(execClient pb.ExecutorClient, demoClient demopb.DemoBackendClient) *judger {
	return &judger{
		execClient: execClient,
		demoClient: demoClient,

		request:  make(chan *demopb.JudgeClientRequest, 64),
		response: make(chan *demopb.JudgeClientResponse, 64),
	}
}

func (j *judger) Start() {
	go j.demoLoop()
	go j.judgeLoop()
}

func (j *judger) demoLoop() {
	for {
		logger.Sugar().Info("connect to demo")
		j.reportLoop()
		logger.Sugar().Info("disconnected to demo")
		time.Sleep(5 * time.Second)
	}
}

func (j *judger) reportLoop() error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	r, err := j.demoClient.Judge(ctx)
	if err != nil {
		return err
	}
	// read loop
	go func() {
		for {
			req, err := r.Recv()
			logger.Sugar().Debug("request:", req)
			if err != nil {
				cancel()
				return
			}
			j.request <- req
		}
	}()

	// write loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case resp := <-j.response:
				logger.Sugar().Debug("response:", resp)
				if err := r.Send(resp); err != nil {
					cancel()
					return
				}
			}
		}
	}()
	<-ctx.Done()
	return nil
}

func (j *judger) judgeLoop() {
	for {
		req := <-j.request
		j.judgeSingle(req)
	}
}

func (j *judger) judgeSingle(req *demopb.JudgeClientRequest) {
	sTime := time.Now()

	j.response <- &demopb.JudgeClientResponse{
		Id:     req.Id,
		Type:   "progress",
		Status: "Compiling",
	}

	// Compile
	args, err := shlex.Split(req.Language.CompileCmd)
	if err != nil {
		j.response <- &demopb.JudgeClientResponse{Id: req.Id, Type: "finished", Status: fmt.Sprintf("Invalid CompileCmd %v", err)}
		return
	}
	compileReq := &pb.Request{
		Cmd: []*pb.Request_CmdType{{
			Args: args,
			Env:  env,
			Files: []*pb.Request_File{
				{
					File: &pb.Request_File_Memory{
						Memory: &pb.Request_MemoryFile{
							Content: []byte{},
						},
					},
				},
				{
					File: &pb.Request_File_Pipe{
						Pipe: &pb.Request_PipeCollector{
							Name: "stdout",
							Max:  1024,
						},
					},
				},
				{
					File: &pb.Request_File_Pipe{
						Pipe: &pb.Request_PipeCollector{
							Name: "stderr",
							Max:  1024,
						},
					},
				},
			},
			CPULimit:     uint64(10 * time.Second),
			RealCPULimit: uint64(12 * time.Second),
			MemoryLimit:  memoryLimit,
			ProcLimit:    50,
			CopyIn: map[string]*pb.Request_File{
				req.Language.SourceFileName: {
					File: &pb.Request_File_Memory{
						Memory: &pb.Request_MemoryFile{
							Content: []byte(req.Source),
						},
					},
				},
			},
			CopyOut:       []string{"stdout", "stderr"},
			CopyOutCached: strings.Split(req.Language.Executables, " "),
		}},
	}
	compileRet, err := j.execClient.Exec(context.TODO(), compileReq)
	if err != nil {
		j.response <- &demopb.JudgeClientResponse{Id: req.Id, Type: "finished", Status: fmt.Sprintf("Compile Error %v", err)}
		return
	}
	if compileRet.Error != "" {
		j.response <- &demopb.JudgeClientResponse{Id: req.Id, Type: "finished", Status: fmt.Sprintf("Compile Error %v", compileRet.Error)}
		return
	}
	cRet := compileRet.Results[0]
	var result []*demopb.Result
	result = append(result, &demopb.Result{
		Time:   uint64(time.Duration(cRet.Time).Round(time.Millisecond) / time.Millisecond),
		Memory: cRet.Memory >> 10,
		Stdout: string(cRet.Files["stdout"]),
		Stderr: string(cRet.Files["stderr"]),
	})

	// remove exec file
	defer func() {
		for _, fid := range cRet.FileIDs {
			j.execClient.FileDelete(context.TODO(), &pb.FileID{
				FileID: fid,
			})
		}
	}()

	if cRet.Status != pb.Response_Result_Accepted {
		j.response <- &demopb.JudgeClientResponse{
			Id:      req.Id,
			Type:    "finished",
			Status:  fmt.Sprintf("Compile %v", compileRet.Error),
			Results: result,
		}
		return
	}

	j.response <- &demopb.JudgeClientResponse{
		Id:     req.Id,
		Type:   "progress",
		Status: "Compiled",
	}

	var completed int32

	runResult := make([]*demopb.Result, noCase)
	for i := range runResult {
		runResult[i] = new(demopb.Result)
	}
	runStatus := make([]pb.Response_Result_StatusType, noCase)
	var eg errgroup.Group
	for i := 0; i < noCase; i++ {
		i := i
		eg.Go(func() (err error) {
			defer func() {
				if err != nil {
					runResult[i].Log = string(err.Error())
					runStatus[i] = pb.Response_Result_JudgementFailed
				}
			}()

			args, err := shlex.Split(req.Language.RunCmd)
			if err != nil {
				return err
			}
			input := strconv.Itoa(i) + " " + strconv.Itoa(i)
			ansContent := strconv.Itoa(i + i)
			// java, go, node needs more threads.. need a better way
			// may be add cpu bandwidth on cgroup..
			var procLimit uint64 = 1
			switch req.Language.Name {
			case "java":
				procLimit = 25
			case "go", "javascript", "typescript", "ruby", "csharp", "perl":
				procLimit = 12
			}
			copyin := make(map[string]*pb.Request_File)
			for k, v := range cRet.FileIDs {
				copyin[k] = &pb.Request_File{
					File: &pb.Request_File_Cached{
						Cached: &pb.Request_CachedFile{
							FileID: v,
						},
					},
				}
			}
			execReq := &pb.Request{
				Cmd: []*pb.Request_CmdType{{
					Args: args,
					Env:  env,
					Files: []*pb.Request_File{
						{
							File: &pb.Request_File_Memory{
								Memory: &pb.Request_MemoryFile{
									Content: []byte(input),
								},
							},
						},
						{
							File: &pb.Request_File_Pipe{
								Pipe: &pb.Request_PipeCollector{
									Name: "stdout",
									Max:  1024,
								},
							},
						},
						{
							File: &pb.Request_File_Pipe{
								Pipe: &pb.Request_PipeCollector{
									Name: "stderr",
									Max:  1024,
								},
							},
						},
					},
					CPULimit:     uint64(3 * time.Second),
					RealCPULimit: uint64(3 * time.Second),
					MemoryLimit:  memoryLimit,
					ProcLimit:    procLimit,
					CopyIn:       copyin,
					CopyOut:      []string{"stdout", "stderr"},
				}},
			}
			response, err := j.execClient.Exec(context.TODO(), execReq)
			if err != nil {
				return err
			}
			if response.Error != "" {
				return fmt.Errorf("case %d %v", i, response.Error)
			}
			ret := response.Results[0]
			err = diff.Compare(bytes.NewBufferString(ansContent), bytes.NewBuffer(ret.Files["stdout"]))
			if err != nil && ret.Status == pb.Response_Result_Accepted {
				ret.Status = pb.Response_Result_WrongAnswer
				runResult[i].Log = err.Error()
			}
			runResult[i].Time = ret.Time / 1e6
			runResult[i].Memory = ret.Memory >> 10
			runResult[i].Stdin = input
			runResult[i].Stdout = string(ret.Files["stdout"])
			runResult[i].Stderr = string(ret.Files["stderr"])
			runStatus[i] = ret.Status

			n := atomic.AddInt32(&completed, 1)
			j.response <- &demopb.JudgeClientResponse{
				Id:     req.Id,
				Type:   "progress",
				Status: fmt.Sprintf("Judging (%d / %d)", n, noCase),
			}
			return nil
		})
	}
	status := pb.Response_Result_Accepted
	err = eg.Wait()
	if err != nil {
		status = pb.Response_Result_JudgementFailed
	}
	for _, r := range runStatus {
		if r > status {
			status = r
		}
	}
	result = append(result, runResult...)

	t := time.Since(sTime)
	taskHist.WithLabelValues(status.String()).Observe(t.Seconds())
	taskSummry.WithLabelValues(status.String()).Observe(t.Seconds())

	j.response <- &demopb.JudgeClientResponse{
		Id:      req.Id,
		Type:    "finished",
		Status:  status.String(),
		Results: result,
	}
}
