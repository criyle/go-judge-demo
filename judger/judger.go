package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/criyle/go-sandbox/daemon"
)

const (
	envWebURL = "WEB_URL"
)

// currently only allow one deamon to test
var m *daemon.Master

func initM() {
	var err error
	root, err := ioutil.TempDir("", "dm")
	log.Println("init: ", root, err)
	m, err = daemon.New(root)
	log.Println("init: ", m, err)
}

func main() {
	daemon.ContainerInit()
	initM()

	retryTime := time.Second
	input := make(chan job, 64)
	input2 := make(chan job, 64)
	input3 := make(chan job, 64)
	output := make(chan Model, 64)

	// start run loop
	go runLoop(input, output, false, false)
	go runLoop(input2, output, true, false)
	go runLoop(input3, output, false, true)

	for {
		j, err := dialWS(os.Getenv(envWebURL))
		if err != nil {
			log.Println("ws:", err)
			retryTime += time.Second
			time.Sleep(retryTime)
			continue
		} else {
			log.Println("ws connected")
		}
	loop:
		for {
			select {
			case <-j.disconnet:
				log.Println("ws disconneted")
				break loop

			case s := <-j.submit:
				input <- s
				input2 <- s
				input3 <- s

			case o := <-output:
				j.update <- o
			}
		}
	}
}
