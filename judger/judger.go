package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/criyle/go-judge/language"
	"github.com/criyle/go-judge/runner"
	"github.com/criyle/go-judge/taskqueue/channel"
	"github.com/criyle/go-sandbox/daemon"
	"github.com/criyle/go-sandbox/pkg/mount"
)

const (
	envWebURL = "WEB_URL"
)

var (
	compileEnv = []string{
		pathEnv,
		"GOCACHE=/tmp",
	}
	runEnv = []string{pathEnv}
)

func init() {
	daemon.Init()
}

func main() {
	done := make(chan struct{})
	root, err := ioutil.TempDir("", "dm")
	if err != nil {
		panic(err)
	}
	q := channel.New()
	m, err := mount.NewBuilder().
		// basic exec and lib
		WithBind("/bin", "bin", true).
		WithBind("/lib", "lib", true).
		WithBind("/lib64", "lib64", true).
		WithBind("/usr", "usr", true).
		// java wants /proc/self/exe as it need relative path for lib
		// however, /proc gives interface like /proc/1/fd/3 ..
		// it is fine since open that file will be a EPERM
		// changing the fs uid and gid would be a good idea
		WithMount(mount.Mount{
			Source: "proc",
			Target: "proc",
			FsType: "proc",
			Flags:  syscall.MS_NOSUID,
		}).
		// some compiler have multiple version
		WithBind("/etc/alternatives", "etc/alternatives", true).
		// fpc wants /etc/fpc.cfg
		WithBind("/etc/fpc.cfg", "etc/fpc.cfg", true).
		// go wants /dev/null
		WithBind("/dev/null", "dev/null", false).
		// work dir
		WithTmpfs("w", "size=8m,nr_inodes=4k").
		// tmp dir
		WithTmpfs("tmp", "size=8m,nr_inodes=4k").
		// finished
		Build(true)

	if err != nil {
		panic(err)
	}
	b := &daemon.Builder{
		Root:   root,
		Mounts: m,
	}
	r := &runner.Runner{
		Builder:  b,
		Queue:    q,
		Language: &dumbLang{},
	}
	const parallism = 4
	for i := 0; i < parallism; i++ {
		go r.Loop(done)
	}

	retryTime := 3 * time.Second
	input := make(chan job, 64)
	output := make(chan Model, 64)

	// start run loop
	go runLoop(input, output, q)

	for {
		j, err := dialWS(os.Getenv(envWebURL))
		if err != nil {
			log.Println("ws:", err)
			time.Sleep(retryTime)
			continue
		}
		log.Println("ws connected")
		judgerLoop(j, input, output)
	}
}

type dumbLang struct{}

func (l *dumbLang) Get(n string, t language.Type) language.ExecParam {
	var d Language
	json.NewDecoder(strings.NewReader(n)).Decode(&d)
	switch t {
	case language.TypeCompile:
		return language.ExecParam{
			Args:              strings.Split(d.CompileCmd, " "),
			Env:               compileEnv,
			SourceFileName:    d.SourceFileName,
			CompiledFileNames: strings.Split(d.Executables, " "),
			TimeLimit:         10 * uint64(time.Millisecond),
			MemoryLimit:       512 << 10,
			ProcLimit:         100,
			OutputLimit:       64 << 10,
		}
	case language.TypeExec:
		// java, go, node needs more threads.. need a better way
		// may be add cpu bandwidth on cgroup..
		var procLimit uint64 = 1
		switch d.Name {
		case "java":
			procLimit = 25
		case "go":
			procLimit = 12
		case "javascript":
			procLimit = 12
		}
		return language.ExecParam{
			Args:              strings.Split(d.RunCmd, " "),
			Env:               runEnv,
			SourceFileName:    d.SourceFileName,
			CompiledFileNames: strings.Split(d.Executables, " "),
			ProcLimit:         procLimit,
		}
	}
	return language.ExecParam{}
}

func judgerLoop(j *judger, input chan job, output chan Model) {
	for {
		select {
		case <-j.disconnet:
			log.Println("ws disconneted")
			return

		case s := <-j.submit:
			log.Println("input: ", s)
			input <- s

		case o := <-output:
			log.Println("output: ", o)
			j.update <- o
		}
	}
}
