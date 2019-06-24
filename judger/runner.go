package main

import (
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
		m, err := run(cargs, workPath, "compiler", stdin, stdout, stderr, 10, false)
		if err != nil {
			log.Println("runLoop: ", err)
			output <- Model{
				ID:     i.ID,
				Status: "CJGF",
			}
			continue
		}
		m.ID = i.ID
		log.Println("runLoop compiled: ", m)
		if m.Status != "AC" {
			m.Status = "JGF " + m.Status
			output <- *m
			continue
		}
		m.Status = "Compiled"
		output <- *m

		// run
		args := []string{exec}
		m, err = run(args, workPath, "", stdin, stdout, stderr, 1, false)
		if err != nil {
			log.Println("runLoop: ", err)
			output <- Model{
				ID:     i.ID,
				Status: "JGF",
			}
			continue
		}
		m.ID = i.ID
		log.Println("runLoop executed: ", m)
		output <- *m
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

func run(args []string, workPath, pType, stdin, stdout, stderr string, timeLimit uint, showDetails bool) (*Model, error) {
	var err error
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
		Env:     os.Environ(),
		WorkDir: workPath,
		RLimits: runprogram.RLimits{
			CPU:      timeLimit,
			CPUHard:  timeLimit + 2,
			FileSize: outputLimit,
			Data:     memoryLimit + 16<<10,
			Stack:    memoryLimit,
		},
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

	rt, err := runner.Start()
	status := "AC"
	if err != nil {
		status = err.Error()
	}
	if rt.ExitCode != 0 {
		status = "Exited: " + strconv.Itoa(rt.ExitCode)
	}
	return &Model{
		Status: status,
		Time:   uint64(rt.UserTime),
		Memory: uint64(rt.UserMem),
		Date:   uint64(time.Now().Unix()),
		Stdin:  readfile(stdin),
		Stdout: readfile(stdout),
		Stderr: readfile(stderr),
	}, nil
}
