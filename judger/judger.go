package main

import (
	"log"
	"os"
	"time"
)

const (
	envWebURL = "WEB_URL"
)

func main() {
	retryTime := time.Second
	input := make(chan job, 64)
	defer close(input)
	input2 := make(chan job, 64)
	defer close(input2)
	output := make(chan Model, 64)
	defer close(output)

	// start run loop
	go runLoop(input, output, false)
	go runLoop(input2, output, true)

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

			case o := <-output:
				j.update <- o
			}
		}
	}
}
