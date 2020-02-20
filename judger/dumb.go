package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/criyle/go-judge/file"
	"github.com/criyle/go-judge/file/memfile"
	"github.com/criyle/go-judge/language"
	"github.com/criyle/go-judge/types"
)

const (
	outputLimit = 64 << 10  // 64k
	memoryLimit = 256 << 20 // 256m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/bin:/usr/bin:/bin"
	noCase      = 8
)

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
			TimeLimit:         10 * time.Second,
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

const total = 500

type dumbBuilder struct {
}

func (b *dumbBuilder) Build([]file.File) (types.ProblemConfig, error) {
	c := make([]types.Case, 0, total)
	for i := 0; i < total; i++ {
		inputContent := strconv.Itoa(i) + " " + strconv.Itoa(i)
		outputContent := strconv.Itoa(i + i)
		c = append(c, types.Case{
			Input:  memfile.New("input", []byte(inputContent)),
			Answer: memfile.New("output", []byte(outputContent)),
		})
	}

	return types.ProblemConfig{
		Type: "standard",
		Subtasks: []types.SubTask{
			types.SubTask{
				ScoringType: "sum",
				Score:       100,
				Cases:       c,
			},
		},
	}, nil
}
