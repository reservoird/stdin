package main

import (
	"bufio"
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
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	o.run = true
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		err = queue.Push([]byte(line))
		if err != nil {
			return err
		}

		select {
		case <-done:
			o.run = false
		default:
		}
	}
	return nil
}

// Ingester for stdin
var Ingester stdin
