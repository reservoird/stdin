package main

import (
	"bufio"
	"os"
	"sync"
)

type stdin struct {
	run  bool
	done <-chan struct{}
	wg   *sync.WaitGroup
}

// Config configures ingester
func (o *stdin) Config(cfg string, done <-chan struct{}, wg *sync.WaitGroup) error {
	o.done = done
	o.wg = wg
	return nil
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(channel chan<- []byte) error {
	reader := bufio.NewReader(os.Stdin)
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		channel <- []byte(line)

		select {
		case <-o.done:
			o.run = false
		default:
		}
	}
	return nil
}

// Ingester for stdin
var Ingester stdin
