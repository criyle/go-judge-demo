package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/criyle/go-judger/rlimit"
	"github.com/criyle/go-judger/runconfig"
	"github.com/criyle/go-judger/runprogram"
	"github.com/criyle/go-judger/rununshared"
	"github.com/criyle/go-judger/tracer"
)

const (
	outputLimit = 64        // 64k
	memoryLimit = 512 << 10 // 512m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
)

func runLoop(input chan job, output chan Model, namespace bool) {
	prefix := "ptrace: "
	if namespace {
		prefix = "namespace: "
	}
	subdir := "p"
	if namespace {
		subdir = "n"
	}

	for {
		i, ok := <-input
		if !ok {
			break
		}
		workPath, _ := os.Getwd()
		p := path.Join(workPath, runDir, i.ID, subdir)
		os.MkdirAll(p, 0755)

		code := path.Join(p, "code.cc")
		stdin := path.Join(p, "stdin")
		ioutil.WriteFile(code, []byte(i.Code), 0755)
		ioutil.WriteFile(stdin, []byte("1 1"), 0755)

		// compile
		cargs := []string{"/usr/bin/g++", "-lm", "-o", "exec", "code.cc"}
		u, err := run(cargs, p, "compiler", "stdin", "stdout", "stderr", 10, namespace, namespace)
		if err != nil {
			log.Println("runLoop: ", err)
			output <- Model{
				ID: i.ID,
				Update: &Update{
					Status: prefix + "CJGF: " + err.Error(),
				},
			}
			continue
		}
		log.Println("runLoop compiled: ", u)
		if u.Status != "AC" {
			u.Status = prefix + "CJGF " + u.Status
			output <- Model{
				ID:     i.ID,
				Update: u,
			}
			continue
		}
		u.Status = prefix + "Compiled"
		output <- Model{
			ID:     i.ID,
			Update: u,
		}

		// run
		args := []string{"exec"}
		u, err = run(args, p, "compiler", "stdin", "stdout", "stderr", 1, false, namespace)
		if err != nil {
			log.Println("runLoop: ", err)
			output <- Model{
				ID: i.ID,
				Update: &Update{
					Status: prefix + "JGF: " + err.Error(),
				},
			}
			continue
		}
		log.Println("runLoop executed: ", u)
		u.Status = prefix + u.Status
		output <- Model{
			ID:     i.ID,
			Update: u,
		}
	}
}

func readfile(filename string) string {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	if len(b) > 1024 {
		return string(b[:1024])
	}
	return string(b)
}

// Runner can be ptraced runner or namespaced runner
type Runner interface {
	Start() (tracer.TraceResult, error)
}

func run(args []string, workPath, pType, stdin, stdout, stderr string, timeLimit uint64, showDetails, namespace bool) (*Update, error) {
	var (
		err       error
		startTime = time.Now().UnixNano()
		runner    Runner
	)
	h := runconfig.GetConf(pType, workPath, args, nil, nil, false, showDetails)
	files := make([]*os.File, 3)

	fin := path.Join(workPath, "stdin")
	fout := path.Join(workPath, "stdout")
	ferr := path.Join(workPath, "stderr")

	if files[0], err = os.OpenFile(fin, os.O_RDONLY, 0755); err != nil {
		return nil, err
	}
	defer files[0].Close()
	if files[1], err = os.OpenFile(fout, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
		return nil, err
	}
	defer files[1].Close()
	if files[2], err = os.OpenFile(ferr, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
		return nil, err
	}
	defer files[2].Close()

	fds := make([]uintptr, 3)
	for i, x := range files {
		fds[i] = x.Fd()
	}

	rlims := rlimit.RLimits{
		CPU:      timeLimit,
		CPUHard:  timeLimit + 2,
		FileSize: outputLimit,
		Data:     memoryLimit + 16<<10,
		Stack:    memoryLimit,
	}

	if namespace {
		h.SyscallAllow = append(h.SyscallAllow, h.SyscallTrace...)
		root, err := ioutil.TempDir("", "ns")
		if err != nil {
			return nil, err
		}
		runner = &rununshared.RunUnshared{
			Args:    h.Args,
			Env:     []string{pathEnv},
			WorkDir: "/w",
			Files:   fds,
			RLimits: rlims,
			ResLimits: tracer.ResLimit{
				TimeLimit:     timeLimit * 1e3,
				RealTimeLimit: uint64(timeLimit+2) * 1e3,
				MemoryLimit:   memoryLimit << 10,
			},
			SyscallAllowed: h.SyscallAllow,
			Root:           root,
			Mounts: rununshared.GetDefaultMounts(root, []rununshared.AddBind{
				{
					Source: workPath,
					Target: "w",
				},
			}),
			ShowDetails: true,
		}
	} else {
		runner = &runprogram.RunProgram{
			Args:    h.Args,
			Env:     []string{pathEnv},
			WorkDir: workPath,
			RLimits: rlims,
			TraceLimit: runprogram.TraceLimit{
				TimeLimit:     timeLimit * 1e3,
				RealTimeLimit: (timeLimit + 2) * 1e3,
				MemoryLimit:   memoryLimit,
			},
			Files:          fds,
			SyscallAllowed: h.SyscallAllow,
			SyscallTraced:  h.SyscallTrace,
			ShowDetails:    showDetails,
			Handler:        h,
		}
	}

	rt, err := runner.Start()
	status := "AC"
	if err != nil {
		status = err.Error()
	}
	if rt.ExitCode != 0 {
		status = "Exited: " + strconv.Itoa(rt.ExitCode)
	}
	l := fmt.Sprintf("SetUpTime = %d ms\nRunningTime = %d ms\nRealTime = %d ms",
		rt.TraceStat.SetUpTime/int64(time.Millisecond), rt.TraceStat.RunningTime/int64(time.Millisecond),
		(time.Now().UnixNano()-startTime)/int64(time.Millisecond))
	return &Update{
		Status: status,
		Time:   uint64(rt.UserTime),
		Memory: uint64(rt.UserMem),
		Date:   uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		Stdin:  readfile(fin),
		Stdout: readfile(fout),
		Stderr: readfile(ferr),
		Log:    l,
	}, nil
}
