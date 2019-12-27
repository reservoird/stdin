package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/reservoird/icd"
)

type stdin struct {
	run bool
}

// Config configures ingester
func (o *stdin) Config(cfg string) error {
	return nil
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	fmt.Printf("stdin.ingest: into\n")
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	o.run = true
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Printf("stdin.ingest: before push\n")
		err = queue.Push([]byte(line))
		fmt.Printf("stdin.ingest: after push\n")
		if err != nil {
			return err
		}

		select {
		case <-done:
			o.run = false
		default:
		}
	}
	fmt.Printf("stdin.ingest: outof\n")
	return nil
}

// Ingester for stdin
var Ingester stdin
