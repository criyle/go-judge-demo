package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/criyle/go-judger/cgroup"
	"github.com/criyle/go-judger/deamon"
	"github.com/criyle/go-judger/runconfig"
	"github.com/criyle/go-judger/runprogram"
	"github.com/criyle/go-judger/rununshared"
	"github.com/criyle/go-judger/types/rlimit"
	"github.com/criyle/go-judger/types/specs"
)

const (
	outputLimit = 64        // 64k
	memoryLimit = 256 << 10 // 256m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/bin:/usr/bin:/bin"
)

func runLoop(input chan job, output chan Model, namespace, ud bool) {
	prefix := "ptrace: "
	subdir := "p"
	if namespace {
		prefix = "namespace: "
		subdir = "n"
	} else if ud {
		prefix = "d: "
		subdir = "d"
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

		if ud {
			f, _ := os.Open(code)
			m.CopyIn(f, "code.cc")
		}

		// compile
		cargs := []string{"/usr/bin/g++", "-lm", "-o", "exec", "code.cc"}
		u, err := run(cargs, p, "compiler", "stdin", "stdout", "stderr", 10, namespace, namespace, ud)
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
		u, err = run(args, p, "compiler", "stdin", "stdout", "stderr", 1, false, namespace, ud)
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
		if ud {
			m.Reset()
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
	Start() (specs.TraceResult, error)
}

func run(args []string, workPath, pType, stdin, stdout, stderr string, timeLimit uint64, showDetails, namespace, ud bool) (*Update, error) {
	var (
		err       error
		startTime = time.Now().UnixNano()
		runner    Runner
		rt        specs.TraceResult
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

	cg, err := cgroup.NewCGroup("judger")
	if err != nil {
		return nil, err
	}
	defer cg.Destroy()
	if err = cg.SetMemoryLimitInBytes(memoryLimit << 10); err != nil {
		return nil, err
	}

	syncFunc := func(pid int) error {
		if cg != nil {
			if err := cg.AddProc(pid); err != nil {
				return err
			}
		}
		return nil
	}

	rlims := rlimit.RLimits{
		CPU:      timeLimit,
		CPUHard:  timeLimit + 2,
		FileSize: outputLimit,
		Data:     memoryLimit + 16<<10,
		Stack:    memoryLimit,
	}

	if ud {
		var s *deamon.ExecveStatus
		sTime := time.Now()
		s, err = m.Execve(&deamon.ExecveParam{
			Args:     args,
			Envv:     []string{pathEnv},
			Fds:      fds,
			RLimits:  rlims.PrepareRLimit(),
			SyncFunc: syncFunc,
		})
		if err != nil {
			return nil, err
		}
		rTime := time.Now()
		tC := time.After(time.Duration(int64(timeLimit+1) * int64(time.Second)))
		select {
		case <-tC:
			s.Kill <- 1
			rt = <-s.Wait

		case rt = <-s.Wait:
			s.Kill <- 1
		}
		<-s.Wait
		eTime := time.Now()
		rt.SetUpTime = int64(rTime.Sub(sTime))
		rt.RunningTime = int64(eTime.Sub(rTime))
		if rt.TraceStatus > 0 {
			err = rt.TraceStatus
		}
	} else if namespace {
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
			ResLimits: specs.ResLimit{
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
			SyncFunc:    syncFunc,
		}
	} else {
		runner = &runprogram.RunProgram{
			Args:    h.Args,
			Env:     []string{pathEnv},
			WorkDir: workPath,
			RLimits: rlims,
			TraceLimit: specs.ResLimit{
				TimeLimit:     timeLimit * 1e3,
				RealTimeLimit: (timeLimit + 2) * 1e3,
				MemoryLimit:   memoryLimit,
			},
			Files:          fds,
			SyscallAllowed: h.SyscallAllow,
			SyscallTraced:  h.SyscallTrace,
			ShowDetails:    showDetails,
			Handler:        h,
			SyncFunc:       syncFunc,
		}
	}
	if runner != nil {
		rt, err = runner.Start()
	}
	status := "AC"
	if err != nil {
		status = err.Error()
	}
	if rt.ExitCode != 0 {
		status = "Exited: " + strconv.Itoa(rt.ExitCode)
	}
	cpu, err := cg.CpuacctUsage()
	if err != nil {
		return nil, err
	}
	memory, err := cg.MemoryMaxUsageInBytes()
	if err != nil {
		return nil, err
	}

	l := fmt.Sprintf("SetUpTime = %d ms\nRunningTime = %d ms\nRealTime = %d ms",
		rt.SetUpTime/int64(time.Millisecond), rt.RunningTime/int64(time.Millisecond),
		(time.Now().UnixNano()-startTime)/int64(time.Millisecond))
	return &Update{
		Status: status,
		Time:   uint64(cpu / uint64(time.Millisecond)),
		Memory: uint64(memory >> 10),
		Date:   uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		Stdin:  readfile(fin),
		Stdout: readfile(fout),
		Stderr: readfile(ferr),
		Log:    l,
	}, nil
}
