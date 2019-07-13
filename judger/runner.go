package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/criyle/go-judger/runconfig"
	"github.com/criyle/go-judger/runprogram"
)

const (
	outputLimit = 64        // 64k
	memoryLimit = 512 << 10 // 512m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
)

func runLoop(input chan job, output chan Model) {
	for {
		i, ok := <-input
		if !ok {
			break
		}
		workPath, _ := os.Getwd()
		p := path.Join(workPath, runDir, i.ID)
		os.MkdirAll(p, 0755)

		code := path.Join(p, "code.cc")
		exec := path.Join(p, "exec")
		stdin := path.Join(p, "stdin")
		stdout := path.Join(p, "stdout")
		stderr := path.Join(p, "stderr")
		ioutil.WriteFile(code, []byte(i.Code), 0755)
		ioutil.WriteFile(stdin, []byte("1 1"), 0755)

		// compile
		cargs := []string{"/usr/bin/g++", "-o", exec, code}
		u, err := run(cargs, workPath, "compiler", stdin, stdout, stderr, 10, false)
		if err != nil {
			log.Println("runLoop: ", err)
			output <- Model{
				ID: i.ID,
				Update: &Update{
					Status: "CJGF: " + err.Error(),
				},
			}
			continue
		}
		log.Println("runLoop compiled: ", u)
		if u.Status != "AC" {
			u.Status = "CJGF " + u.Status
			output <- Model{
				ID:     i.ID,
				Update: u,
			}
			continue
		}
		u.Status = "Compiled"
		output <- Model{
			ID:     i.ID,
			Update: u,
		}

		// run
		args := []string{exec}
		u, err = run(args, workPath, "", stdin, stdout, stderr, 1, false)
		if err != nil {
			log.Println("runLoop: ", err)
			output <- Model{
				ID: i.ID,
				Update: &Update{
					Status: "JGF: " + err.Error(),
				},
			}
			continue
		}
		log.Println("runLoop executed: ", u)
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

func run(args []string, workPath, pType, stdin, stdout, stderr string, timeLimit uint, showDetails bool) (*Update, error) {
	var (
		err       error
		startTime = time.Now().UnixNano()
	)
	h := runconfig.GetConf(pType, workPath, args, nil, nil, false, showDetails)
	files := make([]*os.File, 3)

	if files[0], err = os.OpenFile(stdin, os.O_RDONLY, 0755); err != nil {
		return nil, err
	}
	defer files[0].Close()
	if files[1], err = os.OpenFile(stdout, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
		return nil, err
	}
	defer files[1].Close()
	if files[2], err = os.OpenFile(stderr, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
		return nil, err
	}
	defer files[2].Close()

	fds := make([]uintptr, 3)
	for i, x := range files {
		fds[i] = x.Fd()
	}

	runner := &runprogram.RunProgram{
		Args:    h.Args,
		Env:     []string{pathEnv},
		WorkDir: workPath,
		RLimits: runprogram.RLimits{
			CPU:      timeLimit,
			CPUHard:  timeLimit + 2,
			FileSize: outputLimit,
			Data:     memoryLimit + 16<<10,
			Stack:    memoryLimit,
		},
		TraceLimit: runprogram.TraceLimit{
			TimeLimit:     uint64(timeLimit * 1e3),
			RealTimeLimit: uint64(timeLimit+2) * 1e3,
			MemoryLimit:   memoryLimit,
		},
		Files:          fds,
		SyscallAllowed: h.SyscallAllow,
		SyscallTraced:  h.SyscallTrace,
		ShowDetails:    showDetails,
		Handler:        h,
	}

	rt, err := runner.Start()
	status := "AC"
	if err != nil {
		status = err.Error()
	}
	if rt.ExitCode != 0 {
		status = "Exited: " + strconv.Itoa(rt.ExitCode)
	}
	return &Update{
		Status: status,
		Time:   uint64(rt.UserTime),
		Memory: uint64(rt.UserMem),
		Date:   uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		Stdin:  readfile(stdin),
		Stdout: readfile(stdout),
		Stderr: readfile(stderr),
		Log:    fmt.Sprintf("RealTime = %d ms", (time.Now().UnixNano()-startTime)/int64(time.Millisecond)),
	}, nil
}
