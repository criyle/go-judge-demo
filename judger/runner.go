package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/criyle/go-sandbox/config"
	"github.com/criyle/go-sandbox/daemon"
	"github.com/criyle/go-sandbox/pkg/cgroup"
	"github.com/criyle/go-sandbox/pkg/rlimit"
	"github.com/criyle/go-sandbox/pkg/seccomp"
	"github.com/criyle/go-sandbox/pkg/seccomp/libseccomp"
	"github.com/criyle/go-sandbox/runner/ptrace"
	"github.com/criyle/go-sandbox/runner/unshare"
	"github.com/criyle/go-sandbox/types"
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
	Start(<-chan struct{}) (<-chan types.Result, error)
}

type daemonRunner struct {
	*daemon.Master
	*daemon.ExecveParam
}

func (r *daemonRunner) Start(done <-chan struct{}) (<-chan types.Result, error) {
	return r.Master.Execve(done, r.ExecveParam)
}

func run(args []string, workPath, pType, stdin, stdout, stderr string, timeLimit uint64, showDetails, namespace, ud bool) (*Update, error) {
	var (
		err       error
		startTime = time.Now()
		runner    Runner
		rt        types.Result
	)
	args, allow, trace, h := config.GetConf(pType, workPath, args, nil, nil, false)
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

	actionDefault := seccomp.ActionKill
	if showDetails {
		actionDefault = seccomp.ActionTrace.WithReturnCode(seccomp.MsgDisallow)
	}

	if ud {
		runner = &daemonRunner{
			Master: m,
			ExecveParam: &daemon.ExecveParam{
				Args:     args,
				Envv:     []string{pathEnv},
				Fds:      fds,
				RLimits:  rlims.PrepareRLimit(),
				SyncFunc: syncFunc,
			},
		}
	} else if namespace {
		builder := libseccomp.Builder{
			Allow:   append(allow, trace...),
			Default: actionDefault,
		}
		filter, err := builder.Build()
		root, err := ioutil.TempDir("", "ns")
		if err != nil {
			return nil, err
		}
		runner = &unshare.Runner{
			Args:    args,
			Env:     []string{pathEnv},
			WorkDir: "/w",
			Files:   fds,
			RLimits: rlims,
			Limit: types.Limit{
				TimeLimit:   timeLimit * 1e3,
				MemoryLimit: memoryLimit << 10,
			},
			Seccomp: filter,
			Root:    root,
			Mounts: unshare.GetDefaultMounts(root, []unshare.AddBind{
				{
					Source: workPath,
					Target: "w",
				},
			}),
			ShowDetails: true,
			SyncFunc:    syncFunc,
		}
	} else {
		builder := libseccomp.Builder{
			Allow:   allow,
			Trace:   trace,
			Default: actionDefault,
		}
		filter, err := builder.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to create seccomp filter %v", err)
		}
		runner = &ptrace.Runner{
			Args:    args,
			Env:     []string{pathEnv},
			WorkDir: workPath,
			RLimits: rlims,
			Limit: types.Limit{
				TimeLimit:   timeLimit * 1e3,
				MemoryLimit: memoryLimit << 10,
			},
			Files:       fds,
			Seccomp:     filter,
			ShowDetails: showDetails,
			Handler:     h,
			SyncFunc:    syncFunc,
		}
	}

	sTime := time.Now()
	done := make(chan struct{})
	s, err := runner.Start(done)
	rTime := time.Now()
	if err != nil {
		return nil, fmt.Errorf("failed to execve: %v", err)
	}
	tC := time.After(time.Duration(int64(timeLimit) * int64(time.Second)))
	select {
	case <-tC:
		close(done)
		rt = <-s

	case rt = <-s:
	}
	eTime := time.Now()

	if rt.SetUpTime == 0 {
		rt.SetUpTime = rTime.Sub(sTime)
		rt.RunningTime = eTime.Sub(rTime)
	}

	status := "AC"
	if err != nil {
		status = err.Error()
	}
	if rt.ExitStatus != 0 {
		status = "Exited: " + strconv.Itoa(rt.ExitStatus)
	}
	cpu, err := cg.CpuacctUsage()
	if err != nil {
		return nil, err
	}
	memory, err := cg.MemoryMaxUsageInBytes()
	if err != nil {
		return nil, err
	}

	l := fmt.Sprintf("SetUpTime=%v\nRunningTime=%v\nTotalTime=%v",
		time.Duration(rt.SetUpTime), time.Duration(rt.RunningTime),
		time.Since(startTime))
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
