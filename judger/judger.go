package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/criyle/go-judge/language"
	"github.com/criyle/go-judge/runner"
	"github.com/criyle/go-judge/taskqueue/channel"
	"github.com/criyle/go-sandbox/daemon"
)

const (
	envWebURL = "WEB_URL"
)

func main() {
	daemon.ContainerInit()

	done := make(chan struct{})
	root, err := ioutil.TempDir("", "dm")
	if err != nil {
		panic(err)
	}
	q := channel.New()
	r := &runner.Runner{
		Root:     root,
		Queue:    q,
		Language: &dumbLang{},
	}
	go r.Loop(done)
	go r.Loop(done)
	go r.Loop(done)
	go r.Loop(done)

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
			SourceFileName:    d.SourceFileName,
			CompiledFileNames: strings.Split(d.Executables, " "),
			TimeLimit:         10 * uint64(time.Millisecond),
			MemoryLimit:       512 << 10,
			ProcLimit:         25,
			OutputLimit:       64 << 10,
		}
	case language.TypeExec:
		return language.ExecParam{
			Args:              strings.Split(d.RunCmd, " "),
			SourceFileName:    d.SourceFileName,
			CompiledFileNames: strings.Split(d.Executables, " "),
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
