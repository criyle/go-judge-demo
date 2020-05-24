package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // for pprof
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/criyle/go-judge-client/pkg/diff"
	"github.com/criyle/go-judge/pb"
	"github.com/flynn/go-shlex"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	envWebURL        = "WEB_URL"
	envExecServerURL = "EXEC_SERVER_ADDR"
)

const (
	outputLimit = 64 << 10  // 64k
	memoryLimit = 256 << 20 // 256m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/bin:/usr/bin:/bin"
	noCase      = 6
	parallism   = 2
)

var (
	taskHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "judger_task_execute_time_seconds",
		Help:    "Time for whole processed case",
		Buckets: prometheus.ExponentialBuckets(time.Millisecond.Seconds(), 1.4, 30), // 1 ms -> 10s
	}, []string{"status"})

	taskSummry = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "judger_task_execute_time",
		Help:       "Summary for the whole process case time",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"status"})
)

func init() {
	prometheus.MustRegister(taskHist, taskSummry)
}

var env = []string{
	pathEnv,
	"HOME=/tmp",
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// collect metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	execServer := "localhost:5051"
	if e := os.Getenv(envExecServerURL); e != "" {
		execServer = e
	}

	conn, err := grpc.Dial(execServer, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	)
	if err != nil {
		log.Fatalln("client", err)
	}
	client := pb.NewExecutorClient(conn)

	input := make(chan job, 64)
	output := make(chan Model, 64)

	go judgeLoop(client, input, output)
	go clientLoop(input, output)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("interrupted")
}

func clientLoop(input chan<- job, output <-chan Model) {
	for {
		j, err := dialWS(os.Getenv(envWebURL))
		if err != nil {
			log.Println("ws:", err)
			time.Sleep(time.Second * 3)
			continue
		}
		connLoop(j, input, output)
	}
}

func connLoop(j *judgerWS, input chan<- job, output <-chan Model) {
	for {
		select {
		case <-j.disconnet:
			log.Println("ws: disconnected")
			return

		case s := <-j.submit:
			log.Println("ws: input", s)
			input <- s

		case o := <-output:
			j.update <- o
		}
	}
}

func judgeSingle(client pb.ExecutorClient, in job, output chan<- Model) {
	sTime := time.Now()

	output <- Model{
		ID:     in.ID,
		Type:   "progress",
		Status: "Compiling",
	}

	// Compile
	args, err := shlex.Split(in.Lang.CompileCmd)
	if err != nil {
		output <- Model{ID: in.ID, Type: "finished", Status: fmt.Sprintf("Invalid CompileCmd %v", err)}
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
				in.Lang.SourceFileName: {
					File: &pb.Request_File_Memory{
						Memory: &pb.Request_MemoryFile{
							Content: []byte(in.Source),
						},
					},
				},
			},
			CopyOut:       []string{"stdout", "stderr"},
			CopyOutCached: strings.Split(in.Lang.Executables, " "),
		}},
	}
	compileRet, err := client.Exec(context.TODO(), compileReq)
	if err != nil {
		output <- Model{ID: in.ID, Type: "finished", Status: fmt.Sprintf("Compile Error %v", err)}
		return
	}
	if compileRet.Error != "" {
		output <- Model{ID: in.ID, Type: "finished", Status: fmt.Sprintf("Compile Error %v", compileRet.Error)}
		return
	}
	cRet := compileRet.Results[0]
	var result []Result
	result = append(result, Result{
		Time:   cRet.Time / 1e6,
		Memory: cRet.Memory >> 10,
		Stdout: string(cRet.Files["stdout"]),
		Stderr: string(cRet.Files["stderr"]),
	})

	// remove exec file
	defer func() {
		log.Println("file delete", cRet.FileIDs)
		for _, fid := range cRet.FileIDs {
			client.FileDelete(context.TODO(), &pb.FileID{
				FileID: fid,
			})
		}
	}()

	if cRet.Status != pb.Response_Result_Accepted {
		output <- Model{
			ID:      in.ID,
			Type:    "finished",
			Status:  fmt.Sprintf("Compile %v", compileRet.Error),
			Results: result,
		}
		return
	}

	output <- Model{
		ID:     in.ID,
		Type:   "progress",
		Status: "Compiled",
	}

	var completed int32

	runResult := make([]Result, noCase)
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

			args, err := shlex.Split(in.Lang.RunCmd)
			if err != nil {
				return err
			}
			input := strconv.Itoa(i) + " " + strconv.Itoa(i)
			ansContent := strconv.Itoa(i + i)
			// java, go, node needs more threads.. need a better way
			// may be add cpu bandwidth on cgroup..
			var procLimit uint64 = 1
			switch in.Lang.Name {
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
			response, err := client.Exec(context.TODO(), execReq)
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
			output <- Model{
				ID:     in.ID,
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

	output <- Model{
		ID:      in.ID,
		Type:    "finished",
		Status:  status.String(),
		Results: result,
	}
}

func judgeLoop(client pb.ExecutorClient, input <-chan job, output chan<- Model) {
	for {
		// received
		in := <-input
		judgeSingle(client, in, output)
	}
}
