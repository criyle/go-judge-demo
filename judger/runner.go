package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/criyle/go-judge/file/memfile"
	"github.com/criyle/go-judge/taskqueue"
	"github.com/criyle/go-judge/types"
)

const (
	outputLimit = 64        // 64k
	memoryLimit = 256 << 10 // 256m
	runDir      = "run"
	pathEnv     = "PATH=/usr/local/bin:/usr/bin:/bin"
	noCase      = 8
)

func runLoop(input <-chan job, output chan<- Model, q taskqueue.Queue) {
	for {
		task, ok := <-input
		if !ok {
			break
		}

		var b strings.Builder
		results := make([]Result, 0, 2)
		json.NewEncoder(&b).Encode(task.Lang)
		lang := b.String()

		input := memfile.New("input", []byte{})

		output <- Model{
			ID:     task.ID,
			Type:   "progress",
			Status: "Compiling",
		}

		// compile
		result := make(chan types.RunTaskResult, noCase)
		q.Send(types.RunTask{
			Type:      "compile",
			Language:  lang,
			Code:      task.Source,
			InputFile: input,
		}, result)
		ret := <-result
		results = append(results, Result{
			Time:   ret.Time,
			Memory: ret.Memory,
			Stdout: string(ret.UserOutput),
			Stderr: string(ret.UserError),
		})
		if ret.Status != "" {
			output <- Model{
				ID:      task.ID,
				Type:    "finish",
				Status:  "Compile Failed: " + ret.Status,
				Results: results,
			}
			continue
		}

		// run
		for i := 0; i < noCase; i++ {
			input := memfile.New("input", []byte(fmt.Sprintf("%d %d", i+1, i+1)))
			answer := memfile.New("answer", []byte(fmt.Sprintf("%d", i+i+2)))
			q.Send(types.RunTask{
				Type:        "exec",
				Language:    lang,
				Code:        task.Source,
				TimeLimit:   1000,
				MemoryLimit: memoryLimit,
				ExecFiles:   ret.ExecFiles,
				InputFile:   input,
				AnswerFile:  answer,
			}, result)
		}

		var status string
		for i := 0; i < noCase; i++ {
			output <- Model{
				ID:     task.ID,
				Type:   "progress",
				Status: fmt.Sprintf("Judging (%d / %d)", i, noCase),
			}
			ret2 := <-result
			if status == "" {
				status = ret2.Status
			}
			results = append(results, Result{
				Time:   ret2.Time,
				Memory: ret2.Memory,
				Stdin:  string(ret2.Input),
				Stdout: string(ret2.UserOutput),
				Stderr: string(ret2.UserError),
				Log:    string(ret2.SpjOutput),
			})
		}
		if status == "" {
			status = "AC"
		}

		output <- Model{
			ID:      task.ID,
			Type:    "finish",
			Status:  status,
			Results: results,
		}
	}
}
