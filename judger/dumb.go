package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/criyle/go-judge-client/language"
	"github.com/criyle/go-judge-client/problem"
	"github.com/criyle/go-judge/file"
	"github.com/flynn/go-shlex"
)

const (
	outputLimit = 64 << 10  // 64k
	memoryLimit = 256 << 20 // 256m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/bin:/usr/bin:/bin"
	noCase      = 12
	parallism   = 4
)

type dumbLang struct{}

func (l *dumbLang) Get(n string, t language.Type) language.ExecParam {
	var d Language
	json.NewDecoder(strings.NewReader(n)).Decode(&d)

	var (
		compileEnv = []string{
			pathEnv,
		}
		runEnv = []string{
			pathEnv,
		}
	)
	switch d.Name {
	case "java":
		// compileEnv = append(compileEnv, "LD_LIBRARY_PATH=/usr/lib/jvm/java-11-openjdk-amd64/lib/jli")
		// runEnv = append(runEnv, "LD_LIBRARY_PATH=/usr/lib/jvm/java-11-openjdk-amd64/lib/jli")
	case "go":
		compileEnv = append(compileEnv, "GOCACHE=/tmp")
	case "haskell":
		// compileEnv = append(compileEnv, "LD_LIBRARY_PATH=/usr/lib/ghc")
	}

	switch t {
	case language.TypeCompile:
		args, _ := shlex.Split(d.CompileCmd)
		return language.ExecParam{
			Args:              args,
			Env:               compileEnv,
			SourceFileName:    d.SourceFileName,
			CompiledFileNames: strings.Split(d.Executables, " "),
			TimeLimit:         20 * time.Second,
			MemoryLimit:       512 << 20,
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
		case "go", "javascript", "typescript", "ruby", "csharp", "perl":
			procLimit = 12
		}
		args, _ := shlex.Split(d.RunCmd)
		return language.ExecParam{
			Args:              args,
			Env:               runEnv,
			SourceFileName:    d.SourceFileName,
			CompiledFileNames: strings.Split(d.Executables, " "),
			ProcLimit:         procLimit,
		}
	}
	return language.ExecParam{}
}

type dumbBuilder struct {
}

func (b *dumbBuilder) Build([]file.File) (problem.Config, error) {
	c := make([]problem.Case, 0, noCase)
	for i := 0; i < noCase; i++ {
		inputContent := strconv.Itoa(i) + " " + strconv.Itoa(i)
		outputContent := strconv.Itoa(i + i)
		c = append(c, problem.Case{
			Input:  file.NewMemFile("input", []byte(inputContent)),
			Answer: file.NewMemFile("output", []byte(outputContent)),
		})
	}

	return problem.Config{
		Type: "standard",
		Subtasks: []problem.SubTask{
			problem.SubTask{
				ScoringType: "sum",
				Score:       100,
				Cases:       c,
			},
		},
	}, nil
}
