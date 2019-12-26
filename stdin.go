package main

import (
	"bufio"
	"os"
	"sync"
)

type stdin struct {
	run bool
}

// Config configures ingester
func (o *stdin) Config(cfg string) error {
	return nil
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(channel chan<- []byte, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		channel <- []byte(line)

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
