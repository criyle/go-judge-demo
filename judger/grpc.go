package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/criyle/go-judge-client/pkg/diff"
	demopb "github.com/criyle/go-judge-demo/pb"
	"github.com/criyle/go-judge/pb"
	"github.com/google/shlex"
	"go.uber.org/zap"
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
		logger.Info("connect to demo")
		j.reportLoop()
		logger.Info("disconnected to demo")
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
			logger.Debug("request", zap.Any("request", req))
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
				logger.Debug("response", zap.Any("response", resp))
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

func judgeClientResponse(id string, t string, status string) *demopb.JudgeClientResponse {
	return demopb.JudgeClientResponse_builder{
		Id:     &id,
		Type:   &t,
		Status: &status,
	}.Build()
}

func (j *judger) judgeSingle(req *demopb.JudgeClientRequest) {
	sTime := time.Now()

	j.response <- judgeClientResponse(req.GetId(), "progress", "Compiling")

	// Compile
	args, err := shlex.Split(req.GetLanguage().GetCompileCmd())
	if err != nil {
		j.response <- judgeClientResponse(req.GetId(), "finished", fmt.Sprintf("Invalid CompileCmd %v", err))
		return
	}

	copyOutFiles := strings.Split(req.GetLanguage().GetExecutables(), " ")
	copyOut := make([]*pb.Request_CmdCopyOutFile, 0, len(copyOutFiles))
	for _, f := range copyOutFiles {
		copyOut = append(copyOut, pb.Request_CmdCopyOutFile_builder{Name: f}.Build())
	}

	compileReq := pb.Request_builder{
		Cmd: []*pb.Request_CmdType{pb.Request_CmdType_builder{
			Args: args,
			Env:  env,
			Files: []*pb.Request_File{
				pb.Request_File_builder{
					Memory: pb.Request_MemoryFile_builder{
						Content: []byte{},
					}.Build(),
				}.Build(),
				pb.Request_File_builder{
					Pipe: pb.Request_PipeCollector_builder{
						Name: "stdout",
						Max:  4096,
					}.Build(),
				}.Build(),
				pb.Request_File_builder{
					Pipe: pb.Request_PipeCollector_builder{
						Name: "stderr",
						Max:  4096,
					}.Build(),
				}.Build(),
			},
			CpuTimeLimit:   uint64(10 * time.Second),
			ClockTimeLimit: uint64(12 * time.Second),
			MemoryLimit:    memoryLimit,
			ProcLimit:      100,
			CopyIn: map[string]*pb.Request_File{
				req.GetLanguage().GetSourceFileName(): pb.Request_File_builder{
					Memory: pb.Request_MemoryFile_builder{
						Content: []byte(req.GetSource()),
					}.Build(),
				}.Build(),
			},
			CopyOut:       []*pb.Request_CmdCopyOutFile{pb.Request_CmdCopyOutFile_builder{Name: "stdout"}.Build(), pb.Request_CmdCopyOutFile_builder{Name: "stderr"}.Build()},
			CopyOutCached: copyOut,
		}.Build()},
	}.Build()
	compileRet, err := j.execClient.Exec(context.TODO(), compileReq)
	if err != nil {
		j.response <- judgeClientResponse(req.GetId(), "finished", fmt.Sprintf("Compile Error %v", err))
		return
	}
	if compileRet.GetError() != "" {
		j.response <- judgeClientResponse(req.GetId(), "finished", fmt.Sprintf("Compile Error %v", compileRet.GetError()))
		return
	}
	cRet := compileRet.GetResults()[0]
	var result []*demopb.Result
	rtTime := uint64(time.Duration(cRet.GetTime()).Round(time.Millisecond) / time.Millisecond)
	rtMemory := cRet.GetMemory() >> 10
	rtStdout := string(cRet.GetFiles()["stdout"])
	rtStderr := string(cRet.GetFiles()["stderr"])
	result = append(result, demopb.Result_builder{
		Time:   &rtTime,
		Memory: &rtMemory,
		Stdout: &rtStdout,
		Stderr: &rtStderr,
	}.Build())

	// remove exec file
	defer func() {
		for _, fid := range cRet.GetFileIDs() {
			j.execClient.FileDelete(context.TODO(), pb.FileID_builder{
				FileID: fid,
			}.Build())
		}
	}()

	if cRet.GetStatus() != pb.Response_Result_Accepted {
		rt := judgeClientResponse(req.GetId(), "finished", fmt.Sprintf("Compile %v %v", cRet.GetStatus().String(), compileRet.GetError()))
		rt.SetResults(result)
		j.response <- rt
		return
	}

	j.response <- judgeClientResponse(req.GetId(), "progress", "Compiled")

	var completed int32

	io := req.GetInputAnswer()
	runResult := make([]*demopb.Result, len(io))
	for i := range runResult {
		runResult[i] = new(demopb.Result)
	}
	runStatus := make([]pb.Response_Result_StatusType, len(io))
	var eg errgroup.Group
	for i, inputOutput := range io {
		i := i
		inputOutput := inputOutput
		eg.Go(func() (err error) {
			defer func() {
				if err != nil {
					runResult[i].SetLog(string(err.Error()))
					runStatus[i] = pb.Response_Result_JudgementFailed
				}
			}()

			args, err := shlex.Split(req.GetLanguage().GetRunCmd())
			if err != nil {
				return err
			}
			input := inputOutput.GetInput()
			ansContent := inputOutput.GetAnswer()
			// java, go, node needs more threads.. need a better way
			// may be add cpu bandwidth on cgroup..
			var procLimit uint64 = 1
			switch req.GetLanguage().GetName() {
			case "java":
				procLimit = 25
			case "go", "javascript", "typescript", "ruby", "csharp", "perl":
				procLimit = 12
			}
			copyin := make(map[string]*pb.Request_File)
			for k, v := range cRet.GetFileIDs() {
				copyin[k] = pb.Request_File_builder{
					Cached: pb.Request_CachedFile_builder{
						FileID: v,
					}.Build(),
				}.Build()
			}
			execReq := pb.Request_builder{
				Cmd: []*pb.Request_CmdType{pb.Request_CmdType_builder{
					Args: args,
					Env:  env,
					Files: []*pb.Request_File{
						pb.Request_File_builder{
							Memory: pb.Request_MemoryFile_builder{
								Content: []byte(input),
							}.Build(),
						}.Build(),
						pb.Request_File_builder{
							Pipe: pb.Request_PipeCollector_builder{
								Name: "stdout",
								Max:  4096,
							}.Build(),
						}.Build(),
						pb.Request_File_builder{
							Pipe: pb.Request_PipeCollector_builder{
								Name: "stderr",
								Max:  4096,
							}.Build(),
						}.Build(),
					},
					CpuTimeLimit:   uint64(3 * time.Second),
					ClockTimeLimit: uint64(6 * time.Second),
					MemoryLimit:    memoryLimit,
					StackLimit:     memoryLimit,
					ProcLimit:      procLimit,
					CopyIn:         copyin,
					CopyOut:        []*pb.Request_CmdCopyOutFile{pb.Request_CmdCopyOutFile_builder{Name: "stdout"}.Build(), pb.Request_CmdCopyOutFile_builder{Name: "stderr"}.Build()},
				}.Build()},
			}.Build()
			response, err := j.execClient.Exec(context.TODO(), execReq)
			if err != nil {
				return err
			}
			if response.GetError() != "" {
				return fmt.Errorf("case %d %v", i, response.GetError())
			}
			ret := response.GetResults()[0]
			err = diff.Compare(bytes.NewBufferString(ansContent), bytes.NewBuffer(ret.GetFiles()["stdout"]))
			if err != nil && ret.GetStatus() == pb.Response_Result_Accepted {
				ret.SetStatus(pb.Response_Result_WrongAnswer)
				runResult[i].SetLog(err.Error())
			}
			runResult[i].SetTime(ret.GetTime() / 1e6)
			runResult[i].SetMemory(ret.GetMemory() >> 10)
			runResult[i].SetStdin(input)
			runResult[i].SetStdout(string(ret.GetFiles()["stdout"]))
			runResult[i].SetStderr(string(ret.GetFiles()["stderr"]))
			runStatus[i] = ret.GetStatus()

			n := atomic.AddInt32(&completed, 1)
			j.response <- judgeClientResponse(req.GetId(), "progress", fmt.Sprintf("Judging (%d / %d)", n, len(io)))
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

	rt := judgeClientResponse(req.GetId(), "finished", status.String())
	rt.SetResults(result)
	j.response <- rt
}
